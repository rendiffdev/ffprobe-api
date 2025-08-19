package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// AudioWrappingAnalyzer handles audio wrapping format detection and analysis
type AudioWrappingAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewAudioWrappingAnalyzer creates a new audio wrapping analyzer
func NewAudioWrappingAnalyzer(ffprobePath string, logger zerolog.Logger) *AudioWrappingAnalyzer {
	return &AudioWrappingAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger,
	}
}

// AudioWrappingAnalysis contains comprehensive audio wrapping analysis
type AudioWrappingAnalysis struct {
	AudioStreams        map[int]*AudioWrappingInfo   `json:"audio_streams,omitempty"`
	WrappingValidation  *WrappingValidation          `json:"wrapping_validation,omitempty"`
	ProfessionalFormats *ProfessionalFormats         `json:"professional_formats,omitempty"`
	BroadcastCompliance *BroadcastWrappingCompliance `json:"broadcast_compliance,omitempty"`
}

// AudioWrappingInfo contains detailed audio wrapping information for a stream
type AudioWrappingInfo struct {
	StreamIndex          int                   `json:"stream_index"`
	CodecName            string                `json:"codec_name"`
	WrappingFormat       string                `json:"wrapping_format"`
	WrappingType         string                `json:"wrapping_type"`       // "elementary", "packetized", "framed", "embedded"
	ContainerWrapping    string                `json:"container_wrapping"`  // How audio is wrapped in container
	TransportMechanism   string                `json:"transport_mechanism"` // "PES", "raw", "framed", "chunk"
	Packetization        *PacketizationInfo    `json:"packetization,omitempty"`
	FramingInfo          *FramingInfo          `json:"framing_info,omitempty"`
	Synchronization      *SynchronizationInfo  `json:"synchronization,omitempty"`
	ProfessionalWrapping *ProfessionalWrapping `json:"professional_wrapping,omitempty"`
	Issues               []string              `json:"issues,omitempty"`
	Recommendations      []string              `json:"recommendations,omitempty"`
}

// PacketizationInfo contains information about audio packetization
type PacketizationInfo struct {
	PacketSize       int     `json:"packet_size,omitempty"`
	PacketsPerFrame  int     `json:"packets_per_frame,omitempty"`
	PacketAlignment  string  `json:"packet_alignment"` // "byte", "word", "frame"
	HasPacketHeaders bool    `json:"has_packet_headers"`
	PacketHeaderSize int     `json:"packet_header_size,omitempty"`
	ErrorCorrection  bool    `json:"error_correction"`
	PacketBoundaries string  `json:"packet_boundaries"` // "fixed", "variable", "implicit"
	PacketOverhead   float64 `json:"packet_overhead_percent,omitempty"`
}

// FramingInfo contains information about audio framing
type FramingInfo struct {
	FrameSize            int     `json:"frame_size,omitempty"`        // Samples per frame
	FrameDuration        float64 `json:"frame_duration_ms,omitempty"` // Duration in milliseconds
	FrameAlignment       string  `json:"frame_alignment"`             // "byte", "bit", "sample"
	HasFrameHeaders      bool    `json:"has_frame_headers"`
	FrameHeaderSize      int     `json:"frame_header_size,omitempty"`
	FrameBoundaryMarkers bool    `json:"frame_boundary_markers"`
	VariableFrameSize    bool    `json:"variable_frame_size"`
	FrameOverhead        float64 `json:"frame_overhead_percent,omitempty"`
}

// SynchronizationInfo contains audio synchronization information
type SynchronizationInfo struct {
	SyncMethod            string   `json:"sync_method"` // "PTS", "DTS", "timestamp", "embedded", "none"
	HasTimestamps         bool     `json:"has_timestamps"`
	TimestampResolution   int      `json:"timestamp_resolution,omitempty"` // Ticks per second
	SyncAccuracy          string   `json:"sync_accuracy"`                  // "sample", "frame", "packet", "approximate"
	PresentationDelay     float64  `json:"presentation_delay_ms,omitempty"`
	SynchronizationIssues []string `json:"synchronization_issues,omitempty"`
}

// ProfessionalWrapping contains professional audio wrapping information
type ProfessionalWrapping struct {
	AESFormat   *AESWrapping   `json:"aes_format,omitempty"`
	BWFFormat   *BWFWrapping   `json:"bwf_format,omitempty"`
	MXFWrapping *MXFWrapping   `json:"mxf_wrapping,omitempty"`
	PCMWrapping *PCMWrapping   `json:"pcm_wrapping,omitempty"`
	DolbyFormat *DolbyWrapping `json:"dolby_format,omitempty"`
	DTSFormat   *DTSWrapping   `json:"dts_format,omitempty"`
}

// AESWrapping contains AES/EBU digital audio wrapping information
type AESWrapping struct {
	AESStandard       string `json:"aes_standard"` // "AES3", "AES47", "AES67"
	ChannelStatusData string `json:"channel_status_data,omitempty"`
	UserData          string `json:"user_data,omitempty"`
	ValidityBits      bool   `json:"validity_bits"`
	SubframeStructure string `json:"subframe_structure"` // "32-bit", "24-bit"
	BlockStructure    string `json:"block_structure"`    // "192 frames"
	PreamblePattern   string `json:"preamble_pattern,omitempty"`
}

// BWFWrapping contains Broadcast Wave Format wrapping information
type BWFWrapping struct {
	BWFVersion          string `json:"bwf_version"`
	BextChunk           bool   `json:"has_bext_chunk"`
	CodingHistory       string `json:"coding_history,omitempty"`
	Originator          string `json:"originator,omitempty"`
	OriginatorReference string `json:"originator_reference,omitempty"`
	TimeReference       string `json:"time_reference,omitempty"`
	UMID                string `json:"umid,omitempty"`
	LoudnessInfo        string `json:"loudness_info,omitempty"`
}

// MXFWrapping contains MXF audio wrapping information
type MXFWrapping struct {
	EssenceContainer   string `json:"essence_container"`
	EssenceEncoding    string `json:"essence_encoding"`
	WrappingType       string `json:"wrapping_type"`      // "frame", "clip", "custom"
	KAGSize            int    `json:"kag_size,omitempty"` // KLV Alignment Grid
	EditUnitSize       int    `json:"edit_unit_size,omitempty"`
	IndexTable         bool   `json:"has_index_table"`
	PartitionStructure string `json:"partition_structure,omitempty"`
}

// PCMWrapping contains PCM wrapping format information
type PCMWrapping struct {
	SampleFormat  string   `json:"sample_format"`
	Interleaving  string   `json:"interleaving"`   // "interleaved", "planar", "packed"
	ByteAlignment string   `json:"byte_alignment"` // "left", "right", "msb"
	Endianness    string   `json:"endianness"`     // "little", "big"
	SignFormat    string   `json:"sign_format"`    // "signed", "unsigned", "offset_binary"
	PaddingBits   int      `json:"padding_bits,omitempty"`
	ChannelOrder  []string `json:"channel_order,omitempty"`
}

// DolbyWrapping contains Dolby-specific wrapping information
type DolbyWrapping struct {
	DolbyFormat          string `json:"dolby_format"` // "AC-3", "E-AC-3", "TrueHD", "Atmos"
	BitstreamMode        int    `json:"bitstream_mode,omitempty"`
	DialogNormalization  int    `json:"dialog_normalization,omitempty"`
	ChannelConfiguration string `json:"channel_configuration,omitempty"`
	LFEPresent           bool   `json:"lfe_present"`
	CouplingStrategy     string `json:"coupling_strategy,omitempty"`
	DataRate             string `json:"data_rate,omitempty"`
}

// DTSWrapping contains DTS-specific wrapping information
type DTSWrapping struct {
	DTSFormat         string `json:"dts_format"` // "DTS", "DTS-HD", "DTS:X"
	CoreBitrate       int    `json:"core_bitrate,omitempty"`
	ExtensionBitrate  int    `json:"extension_bitrate,omitempty"`
	ChannelLayout     string `json:"channel_layout,omitempty"`
	LosslessMode      bool   `json:"lossless_mode"`
	MultiAssetMode    bool   `json:"multi_asset_mode"`
	CoreChannels      int    `json:"core_channels,omitempty"`
	ExtensionChannels int    `json:"extension_channels,omitempty"`
}

// WrappingValidation contains wrapping validation results
type WrappingValidation struct {
	IsValid             bool     `json:"is_valid"`
	IsStandardCompliant bool     `json:"is_standard_compliant"`
	HasWrappingIssues   bool     `json:"has_wrapping_issues"`
	Issues              []string `json:"issues,omitempty"`
	Warnings            []string `json:"warnings,omitempty"`
	Recommendations     []string `json:"recommendations,omitempty"`
	CompatibilityLevel  string   `json:"compatibility_level"` // "high", "medium", "low"
}

// ProfessionalFormats contains information about professional audio formats
type ProfessionalFormats struct {
	HasProfessionalWrapping bool     `json:"has_professional_wrapping"`
	DetectedFormats         []string `json:"detected_formats,omitempty"`
	BroadcastReady          bool     `json:"broadcast_ready"`
	PostProductionReady     bool     `json:"post_production_ready"`
	ArchivalQuality         bool     `json:"archival_quality"`
	Recommendations         []string `json:"recommendations,omitempty"`
}

// BroadcastWrappingCompliance contains broadcast compliance information
type BroadcastWrappingCompliance struct {
	EBUCompliant     bool     `json:"ebu_compliant"`   // EBU R68, R128
	ATSCCompliant    bool     `json:"atsc_compliant"`  // ATSC A/52, A/85
	DVBCompliant     bool     `json:"dvb_compliant"`   // DVB standards
	ARIBCompliant    bool     `json:"arib_compliant"`  // ARIB standards (Japan)
	ITUCompliant     bool     `json:"itu_compliant"`   // ITU-R BS standards
	AES67Compliant   bool     `json:"aes67_compliant"` // AES67 Audio-over-IP
	ComplianceIssues []string `json:"compliance_issues,omitempty"`
	ComplianceLevel  string   `json:"compliance_level"` // "full", "partial", "non_compliant"
}

// Audio wrapping format definitions
var audioWrappingFormats = map[string]string{
	"pcm_s16le": "PCM Little Endian 16-bit",
	"pcm_s16be": "PCM Big Endian 16-bit",
	"pcm_s24le": "PCM Little Endian 24-bit",
	"pcm_s24be": "PCM Big Endian 24-bit",
	"pcm_s32le": "PCM Little Endian 32-bit",
	"pcm_s32be": "PCM Big Endian 32-bit",
	"pcm_f32le": "PCM Float 32-bit Little Endian",
	"pcm_f32be": "PCM Float 32-bit Big Endian",
	"pcm_f64le": "PCM Float 64-bit Little Endian",
	"pcm_f64be": "PCM Float 64-bit Big Endian",
	"aac":       "AAC (Advanced Audio Coding)",
	"ac3":       "AC-3 (Dolby Digital)",
	"eac3":      "E-AC-3 (Dolby Digital Plus)",
	"truehd":    "Dolby TrueHD",
	"dts":       "DTS (Digital Theater Systems)",
	"mp3":       "MP3 (MPEG-1 Audio Layer III)",
	"mp2":       "MP2 (MPEG-1 Audio Layer II)",
	"opus":      "Opus",
	"vorbis":    "Vorbis",
	"flac":      "FLAC (Free Lossless Audio Codec)",
}

// AnalyzeAudioWrapping performs comprehensive audio wrapping analysis
func (awa *AudioWrappingAnalyzer) AnalyzeAudioWrapping(ctx context.Context, filePath string, streams []StreamInfo, format *FormatInfo) (*AudioWrappingAnalysis, error) {
	analysis := &AudioWrappingAnalysis{
		AudioStreams: make(map[int]*AudioWrappingInfo),
	}

	// Analyze each audio stream
	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) == "audio" {
			wrappingInfo, err := awa.analyzeStreamWrapping(ctx, filePath, stream, format)
			if err != nil {
				awa.logger.Warn().Err(err).Int("stream", stream.Index).Msg("Failed to analyze stream wrapping")
				continue
			}
			analysis.AudioStreams[stream.Index] = wrappingInfo
		}
	}

	// Validate wrapping formats
	analysis.WrappingValidation = awa.validateWrapping(analysis)

	// Analyze professional formats
	analysis.ProfessionalFormats = awa.analyzeProfessionalFormats(analysis)

	// Check broadcast compliance
	analysis.BroadcastCompliance = awa.checkBroadcastCompliance(analysis)

	return analysis, nil
}

// analyzeStreamWrapping analyzes wrapping for a specific audio stream
func (awa *AudioWrappingAnalyzer) analyzeStreamWrapping(ctx context.Context, filePath string, stream StreamInfo, format *FormatInfo) (*AudioWrappingInfo, error) {
	info := &AudioWrappingInfo{
		StreamIndex:     stream.Index,
		CodecName:       stream.CodecName,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Determine wrapping format from codec and container
	info.WrappingFormat = awa.determineWrappingFormat(stream.CodecName)
	info.WrappingType = awa.determineWrappingType(stream.CodecName, format)
	info.ContainerWrapping = awa.determineContainerWrapping(format)
	info.TransportMechanism = awa.determineTransportMechanism(format, stream)

	// Analyze packetization
	if err := awa.analyzePacketization(ctx, filePath, stream, info); err != nil {
		awa.logger.Warn().Err(err).Msg("Failed to analyze packetization")
	}

	// Analyze framing
	if err := awa.analyzeFraming(ctx, filePath, stream, info); err != nil {
		awa.logger.Warn().Err(err).Msg("Failed to analyze framing")
	}

	// Analyze synchronization
	info.Synchronization = awa.analyzeSynchronization(stream)

	// Analyze professional wrapping formats
	info.ProfessionalWrapping = awa.analyzeProfessionalWrapping(stream, format)

	return info, nil
}

// determineWrappingFormat determines the audio wrapping format
func (awa *AudioWrappingAnalyzer) determineWrappingFormat(codecName string) string {
	if format, exists := audioWrappingFormats[strings.ToLower(codecName)]; exists {
		return format
	}
	return fmt.Sprintf("Unknown (%s)", codecName)
}

// determineWrappingType determines how the audio is wrapped
func (awa *AudioWrappingAnalyzer) determineWrappingType(codecName string, format *FormatInfo) string {
	codec := strings.ToLower(codecName)

	// PCM is typically elementary
	if strings.HasPrefix(codec, "pcm") {
		return "elementary"
	}

	// Compressed formats are typically framed
	if codec == "aac" || codec == "mp3" || codec == "ac3" || codec == "eac3" {
		return "framed"
	}

	// Professional formats
	if codec == "truehd" || codec == "dts" {
		return "packetized"
	}

	// Container-dependent
	if format != nil {
		formatName := strings.ToLower(format.FormatName)
		if strings.Contains(formatName, "mp4") || strings.Contains(formatName, "mov") {
			return "embedded"
		}
		if strings.Contains(formatName, "mpegts") {
			return "packetized"
		}
	}

	return "unknown"
}

// determineContainerWrapping determines how audio is wrapped in the container
func (awa *AudioWrappingAnalyzer) determineContainerWrapping(format *FormatInfo) string {
	if format == nil {
		return "unknown"
	}

	formatName := strings.ToLower(format.FormatName)

	wrappingMap := map[string]string{
		"mp4":      "MP4 sample-based",
		"mov":      "QuickTime sample-based",
		"avi":      "AVI chunk-based",
		"wav":      "RIFF chunk-based",
		"mpegts":   "MPEG-TS PES packets",
		"matroska": "Matroska block-based",
		"webm":     "WebM block-based",
		"ogg":      "Ogg page-based",
		"flv":      "FLV tag-based",
		"asf":      "ASF packet-based",
		"mxf":      "MXF KLV-based",
	}

	for container, wrapping := range wrappingMap {
		if strings.Contains(formatName, container) {
			return wrapping
		}
	}

	return "Unknown container wrapping"
}

// determineTransportMechanism determines the transport mechanism
func (awa *AudioWrappingAnalyzer) determineTransportMechanism(format *FormatInfo, stream StreamInfo) string {
	if format == nil {
		return "unknown"
	}

	formatName := strings.ToLower(format.FormatName)

	if strings.Contains(formatName, "mpegts") {
		return "PES (Packetized Elementary Stream)"
	}
	if strings.Contains(formatName, "mp4") || strings.Contains(formatName, "mov") {
		return "Sample-based access units"
	}
	if strings.Contains(formatName, "avi") {
		return "Interleaved chunks"
	}
	if strings.Contains(formatName, "wav") {
		return "Raw PCM data"
	}
	if strings.Contains(formatName, "matroska") || strings.Contains(formatName, "webm") {
		return "Block structure"
	}
	if strings.Contains(formatName, "mxf") {
		return "KLV packets"
	}

	return "Container-specific framing"
}

// analyzePacketization analyzes audio packetization
func (awa *AudioWrappingAnalyzer) analyzePacketization(ctx context.Context, filePath string, stream StreamInfo, info *AudioWrappingInfo) error {
	packetInfo := &PacketizationInfo{
		PacketAlignment:  "unknown",
		HasPacketHeaders: false,
		ErrorCorrection:  false,
		PacketBoundaries: "unknown",
	}

	// Extract packet information using ffprobe
	cmd := []string{
		awa.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "packet=stream_index,size,duration_time,flags",
		"-select_streams", fmt.Sprintf("a:%d", stream.Index),
		"-read_intervals", "%+#10", // First 10 packets
		filePath,
	}

	output, err := awa.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to analyze packetization: %w", err)
	}

	var result struct {
		Packets []struct {
			StreamIndex  int    `json:"stream_index"`
			Size         string `json:"size"`
			DurationTime string `json:"duration_time"`
			Flags        string `json:"flags"`
		} `json:"packets"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("failed to parse packet JSON: %w", err)
	}

	if len(result.Packets) > 0 {
		// Analyze packet sizes
		sizes := make([]int, 0, len(result.Packets))
		for _, packet := range result.Packets {
			if size, err := strconv.Atoi(packet.Size); err == nil {
				sizes = append(sizes, size)
			}
		}

		if len(sizes) > 0 {
			// Calculate statistics
			minSize, maxSize, avgSize := awa.calculatePacketStatistics(sizes)
			packetInfo.PacketSize = avgSize

			// Determine packet boundaries
			if minSize == maxSize {
				packetInfo.PacketBoundaries = "fixed"
			} else {
				packetInfo.PacketBoundaries = "variable"
			}

			// Estimate packet overhead
			if avgSize > 0 {
				// Very rough estimate - actual payload vs total packet size
				estimatedPayload := float64(avgSize) * 0.9 // Assume 10% overhead
				packetInfo.PacketOverhead = (float64(avgSize) - estimatedPayload) / float64(avgSize) * 100.0
			}
		}
	}

	info.Packetization = packetInfo
	return nil
}

// analyzeFraming analyzes audio framing structure
func (awa *AudioWrappingAnalyzer) analyzeFraming(ctx context.Context, filePath string, stream StreamInfo, info *AudioWrappingInfo) error {
	framingInfo := &FramingInfo{
		FrameAlignment:       "unknown",
		HasFrameHeaders:      false,
		FrameBoundaryMarkers: false,
		VariableFrameSize:    false,
	}

	// Extract frame information using ffprobe
	cmd := []string{
		awa.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "frame=stream_index,nb_samples,pkt_size,pkt_duration_time",
		"-select_streams", fmt.Sprintf("a:%d", stream.Index),
		"-read_intervals", "%+#10", // First 10 frames
		filePath,
	}

	output, err := awa.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to analyze framing: %w", err)
	}

	var result struct {
		Frames []struct {
			StreamIndex     int    `json:"stream_index"`
			NBSamples       int    `json:"nb_samples"`
			PktSize         string `json:"pkt_size"`
			PktDurationTime string `json:"pkt_duration_time"`
		} `json:"frames"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("failed to parse frame JSON: %w", err)
	}

	if len(result.Frames) > 0 {
		// Analyze frame properties
		sampleCounts := make([]int, 0, len(result.Frames))
		durations := make([]float64, 0, len(result.Frames))

		for _, frame := range result.Frames {
			if frame.NBSamples > 0 {
				sampleCounts = append(sampleCounts, frame.NBSamples)
			}
			if duration, err := strconv.ParseFloat(frame.PktDurationTime, 64); err == nil {
				durations = append(durations, duration*1000) // Convert to ms
			}
		}

		// Determine frame characteristics
		if len(sampleCounts) > 0 {
			minSamples, maxSamples, avgSamples := awa.calculateFrameStatistics(sampleCounts)
			framingInfo.FrameSize = avgSamples
			framingInfo.VariableFrameSize = (minSamples != maxSamples)
		}

		if len(durations) > 0 {
			_, _, avgDuration := awa.calculateDurationStatistics(durations)
			framingInfo.FrameDuration = avgDuration
		}

		// Analyze codec-specific framing
		awa.analyzeCodecSpecificFraming(stream.CodecName, framingInfo)
	}

	info.FramingInfo = framingInfo
	return nil
}

// analyzeSynchronization analyzes audio synchronization
func (awa *AudioWrappingAnalyzer) analyzeSynchronization(stream StreamInfo) *SynchronizationInfo {
	syncInfo := &SynchronizationInfo{
		SyncMethod:            "unknown",
		HasTimestamps:         false,
		SyncAccuracy:          "unknown",
		SynchronizationIssues: []string{},
	}

	// Analyze based on stream metadata
	if stream.StartTime != "" {
		syncInfo.HasTimestamps = true
		syncInfo.SyncMethod = "container_timestamps"
		syncInfo.SyncAccuracy = "packet"
	}

	// Check for codec-specific sync mechanisms
	codec := strings.ToLower(stream.CodecName)
	switch codec {
	case "aac":
		syncInfo.SyncMethod = "ADTS_headers"
		syncInfo.SyncAccuracy = "frame"
	case "mp3":
		syncInfo.SyncMethod = "frame_headers"
		syncInfo.SyncAccuracy = "frame"
	case "ac3", "eac3":
		syncInfo.SyncMethod = "sync_words"
		syncInfo.SyncAccuracy = "frame"
	case "pcm_s16le", "pcm_s24le", "pcm_s32le":
		syncInfo.SyncMethod = "sample_based"
		syncInfo.SyncAccuracy = "sample"
	}

	// Check for potential synchronization issues
	if stream.Duration == "" {
		syncInfo.SynchronizationIssues = append(syncInfo.SynchronizationIssues, "No duration information available")
	}

	return syncInfo
}

// analyzeProfessionalWrapping analyzes professional audio wrapping formats
func (awa *AudioWrappingAnalyzer) analyzeProfessionalWrapping(stream StreamInfo, format *FormatInfo) *ProfessionalWrapping {
	professional := &ProfessionalWrapping{}

	// Analyze based on codec and container
	codec := strings.ToLower(stream.CodecName)
	var formatName string
	if format != nil {
		formatName = strings.ToLower(format.FormatName)
	}

	// PCM Analysis
	if strings.HasPrefix(codec, "pcm") {
		professional.PCMWrapping = awa.analyzePCMWrapping(stream)
	}

	// AES Analysis
	if awa.isAESFormat(stream, format) {
		professional.AESFormat = awa.analyzeAESWrapping(stream)
	}

	// BWF Analysis
	if strings.Contains(formatName, "wav") && awa.hasBWFMetadata(stream) {
		professional.BWFFormat = awa.analyzeBWFWrapping(stream)
	}

	// MXF Analysis
	if strings.Contains(formatName, "mxf") {
		professional.MXFWrapping = awa.analyzeMXFWrapping(stream)
	}

	// Dolby Analysis
	if codec == "ac3" || codec == "eac3" || codec == "truehd" {
		professional.DolbyFormat = awa.analyzeDolbyWrapping(stream)
	}

	// DTS Analysis
	if strings.HasPrefix(codec, "dts") {
		professional.DTSFormat = awa.analyzeDTSWrapping(stream)
	}

	return professional
}

// validateWrapping validates audio wrapping formats
func (awa *AudioWrappingAnalyzer) validateWrapping(analysis *AudioWrappingAnalysis) *WrappingValidation {
	validation := &WrappingValidation{
		IsValid:             true,
		IsStandardCompliant: true,
		HasWrappingIssues:   false,
		Issues:              []string{},
		Warnings:            []string{},
		Recommendations:     []string{},
		CompatibilityLevel:  "high",
	}

	issueCount := 0
	streamCount := len(analysis.AudioStreams)

	for _, streamInfo := range analysis.AudioStreams {
		// Count issues
		issueCount += len(streamInfo.Issues)

		// Check for common problems
		if streamInfo.WrappingFormat == "Unknown" {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Stream %d has unknown wrapping format", streamInfo.StreamIndex))
			validation.HasWrappingIssues = true
		}

		if streamInfo.TransportMechanism == "unknown" {
			validation.Warnings = append(validation.Warnings,
				fmt.Sprintf("Stream %d transport mechanism unclear", streamInfo.StreamIndex))
		}
	}

	// Determine overall validity
	if issueCount > 0 {
		validation.HasWrappingIssues = true
		if issueCount > streamCount {
			validation.IsValid = false
		}
	}

	// Determine compatibility level
	if issueCount == 0 {
		validation.CompatibilityLevel = "high"
	} else if issueCount <= streamCount {
		validation.CompatibilityLevel = "medium"
	} else {
		validation.CompatibilityLevel = "low"
	}

	return validation
}

// analyzeProfessionalFormats analyzes professional audio formats
func (awa *AudioWrappingAnalyzer) analyzeProfessionalFormats(analysis *AudioWrappingAnalysis) *ProfessionalFormats {
	formats := &ProfessionalFormats{
		HasProfessionalWrapping: false,
		DetectedFormats:         []string{},
		BroadcastReady:          false,
		PostProductionReady:     false,
		ArchivalQuality:         false,
		Recommendations:         []string{},
	}

	for _, streamInfo := range analysis.AudioStreams {
		if streamInfo.ProfessionalWrapping != nil {
			formats.HasProfessionalWrapping = true

			// Check for specific professional formats
			if streamInfo.ProfessionalWrapping.AESFormat != nil {
				formats.DetectedFormats = append(formats.DetectedFormats, "AES/EBU")
				formats.BroadcastReady = true
			}
			if streamInfo.ProfessionalWrapping.BWFFormat != nil {
				formats.DetectedFormats = append(formats.DetectedFormats, "BWF")
				formats.BroadcastReady = true
				formats.ArchivalQuality = true
			}
			if streamInfo.ProfessionalWrapping.MXFWrapping != nil {
				formats.DetectedFormats = append(formats.DetectedFormats, "MXF")
				formats.PostProductionReady = true
				formats.ArchivalQuality = true
			}
		}

		// Check for high-quality PCM
		if strings.Contains(strings.ToLower(streamInfo.CodecName), "pcm") {
			if streamInfo.ProfessionalWrapping != nil && streamInfo.ProfessionalWrapping.PCMWrapping != nil {
				pcm := streamInfo.ProfessionalWrapping.PCMWrapping
				if strings.Contains(pcm.SampleFormat, "24") || strings.Contains(pcm.SampleFormat, "32") {
					formats.PostProductionReady = true
					formats.ArchivalQuality = true
				}
			}
		}
	}

	// Generate recommendations
	if !formats.HasProfessionalWrapping {
		formats.Recommendations = append(formats.Recommendations, "Consider using professional audio wrapping formats for broadcast/post-production")
	}
	if !formats.BroadcastReady {
		formats.Recommendations = append(formats.Recommendations, "Add BWF or AES/EBU wrapping for broadcast compliance")
	}

	return formats
}

// checkBroadcastCompliance checks broadcast compliance
func (awa *AudioWrappingAnalyzer) checkBroadcastCompliance(analysis *AudioWrappingAnalysis) *BroadcastWrappingCompliance {
	compliance := &BroadcastWrappingCompliance{
		EBUCompliant:     true,
		ATSCCompliant:    true,
		DVBCompliant:     true,
		ARIBCompliant:    true,
		ITUCompliant:     true,
		AES67Compliant:   false,
		ComplianceIssues: []string{},
		ComplianceLevel:  "full",
	}

	for _, streamInfo := range analysis.AudioStreams {
		// Check codec compliance
		codec := strings.ToLower(streamInfo.CodecName)

		// ATSC compliance
		if !awa.isATSCCompliantCodec(codec) {
			compliance.ATSCCompliant = false
			compliance.ComplianceIssues = append(compliance.ComplianceIssues,
				fmt.Sprintf("Stream %d codec not ATSC compliant", streamInfo.StreamIndex))
		}

		// EBU compliance
		if !awa.isEBUCompliantWrapping(streamInfo) {
			compliance.EBUCompliant = false
			compliance.ComplianceIssues = append(compliance.ComplianceIssues,
				fmt.Sprintf("Stream %d wrapping not EBU compliant", streamInfo.StreamIndex))
		}

		// AES67 compliance
		if awa.isAES67Compliant(streamInfo) {
			compliance.AES67Compliant = true
		}
	}

	// Determine overall compliance level
	issueCount := len(compliance.ComplianceIssues)
	if issueCount == 0 {
		compliance.ComplianceLevel = "full"
	} else if issueCount <= 2 {
		compliance.ComplianceLevel = "partial"
	} else {
		compliance.ComplianceLevel = "non_compliant"
	}

	return compliance
}

// Helper methods for professional format analysis

func (awa *AudioWrappingAnalyzer) analyzePCMWrapping(stream StreamInfo) *PCMWrapping {
	pcm := &PCMWrapping{
		SampleFormat:  stream.SampleFmt,
		Interleaving:  "unknown",
		ByteAlignment: "unknown",
		Endianness:    "unknown",
		SignFormat:    "unknown",
		ChannelOrder:  []string{},
	}

	sampleFmt := strings.ToLower(stream.SampleFmt)

	// Determine endianness
	if strings.HasSuffix(sampleFmt, "le") {
		pcm.Endianness = "little"
	} else if strings.HasSuffix(sampleFmt, "be") {
		pcm.Endianness = "big"
	}

	// Determine sign format
	if strings.HasPrefix(sampleFmt, "s") {
		pcm.SignFormat = "signed"
	} else if strings.HasPrefix(sampleFmt, "u") {
		pcm.SignFormat = "unsigned"
	}

	// Determine interleaving
	if strings.Contains(sampleFmt, "p") {
		pcm.Interleaving = "planar"
	} else {
		pcm.Interleaving = "interleaved"
	}

	return pcm
}

func (awa *AudioWrappingAnalyzer) isAESFormat(stream StreamInfo, format *FormatInfo) bool {
	// Check for AES/EBU indicators
	if format != nil {
		formatName := strings.ToLower(format.FormatName)
		if strings.Contains(formatName, "aes") || strings.Contains(formatName, "ebu") {
			return true
		}
	}

	// Check for AES-specific sample rates
	if stream.SampleRate == "48000" || stream.SampleRate == "96000" {
		if stream.Channels == 2 && strings.Contains(strings.ToLower(stream.SampleFmt), "24") {
			return true
		}
	}

	return false
}

func (awa *AudioWrappingAnalyzer) analyzeAESWrapping(stream StreamInfo) *AESWrapping {
	aes := &AESWrapping{
		AESStandard:       "AES3",
		ValidityBits:      true,
		SubframeStructure: "32-bit",
		BlockStructure:    "192 frames",
	}

	// Determine AES standard based on characteristics
	if stream.SampleRate == "48000" {
		aes.AESStandard = "AES3"
	} else if stream.SampleRate == "96000" {
		aes.AESStandard = "AES3"
	}

	return aes
}

func (awa *AudioWrappingAnalyzer) hasBWFMetadata(stream StreamInfo) bool {
	// Check for BWF-specific metadata
	if bext, exists := stream.Tags["bext"]; exists && bext != "" {
		return true
	}
	if originator, exists := stream.Tags["originator"]; exists && originator != "" {
		return true
	}
	return false
}

func (awa *AudioWrappingAnalyzer) analyzeBWFWrapping(stream StreamInfo) *BWFWrapping {
	bwf := &BWFWrapping{
		BWFVersion: "2",
		BextChunk:  true,
	}

	// Extract BWF metadata
	if originator, exists := stream.Tags["originator"]; exists {
		bwf.Originator = originator
	}
	if originatorRef, exists := stream.Tags["originator_reference"]; exists {
		bwf.OriginatorReference = originatorRef
	}
	if timeRef, exists := stream.Tags["time_reference"]; exists {
		bwf.TimeReference = timeRef
	}

	return bwf
}

func (awa *AudioWrappingAnalyzer) analyzeMXFWrapping(stream StreamInfo) *MXFWrapping {
	mxf := &MXFWrapping{
		EssenceContainer:   "Unknown",
		EssenceEncoding:    stream.CodecName,
		WrappingType:       "frame",
		IndexTable:         false,
		PartitionStructure: "single_partition",
	}

	// Determine essence container based on codec
	codec := strings.ToLower(stream.CodecName)
	if strings.HasPrefix(codec, "pcm") {
		mxf.EssenceContainer = "Generic Container Multiple Mappings"
	} else {
		mxf.EssenceContainer = "Generic Container"
	}

	return mxf
}

func (awa *AudioWrappingAnalyzer) analyzeDolbyWrapping(stream StreamInfo) *DolbyWrapping {
	dolby := &DolbyWrapping{
		DolbyFormat:      stream.CodecName,
		LFEPresent:       false,
		CouplingStrategy: "unknown",
	}

	// Extract Dolby-specific information
	if stream.Channels >= 6 {
		dolby.LFEPresent = true
		dolby.ChannelConfiguration = "5.1"
	} else if stream.Channels == 2 {
		dolby.ChannelConfiguration = "2.0"
	}

	return dolby
}

func (awa *AudioWrappingAnalyzer) analyzeDTSWrapping(stream StreamInfo) *DTSWrapping {
	dts := &DTSWrapping{
		DTSFormat:      stream.CodecName,
		LosslessMode:   false,
		MultiAssetMode: false,
		CoreChannels:   stream.Channels,
	}

	// Determine DTS format specifics
	codec := strings.ToLower(stream.CodecName)
	if strings.Contains(codec, "hd") {
		dts.DTSFormat = "DTS-HD"
		dts.LosslessMode = strings.Contains(codec, "ma")
	}

	return dts
}

// Helper methods for calculations and validation

func (awa *AudioWrappingAnalyzer) calculatePacketStatistics(sizes []int) (int, int, int) {
	if len(sizes) == 0 {
		return 0, 0, 0
	}

	min, max := sizes[0], sizes[0]
	sum := 0
	for _, size := range sizes {
		if size < min {
			min = size
		}
		if size > max {
			max = size
		}
		sum += size
	}
	avg := sum / len(sizes)
	return min, max, avg
}

func (awa *AudioWrappingAnalyzer) calculateFrameStatistics(samples []int) (int, int, int) {
	if len(samples) == 0 {
		return 0, 0, 0
	}

	min, max := samples[0], samples[0]
	sum := 0
	for _, sample := range samples {
		if sample < min {
			min = sample
		}
		if sample > max {
			max = sample
		}
		sum += sample
	}
	avg := sum / len(samples)
	return min, max, avg
}

func (awa *AudioWrappingAnalyzer) calculateDurationStatistics(durations []float64) (float64, float64, float64) {
	if len(durations) == 0 {
		return 0, 0, 0
	}

	min, max := durations[0], durations[0]
	sum := 0.0
	for _, duration := range durations {
		if duration < min {
			min = duration
		}
		if duration > max {
			max = duration
		}
		sum += duration
	}
	avg := sum / float64(len(durations))
	return min, max, avg
}

func (awa *AudioWrappingAnalyzer) analyzeCodecSpecificFraming(codecName string, framingInfo *FramingInfo) {
	codec := strings.ToLower(codecName)

	switch codec {
	case "aac":
		framingInfo.FrameSize = 1024 // AAC typically uses 1024 samples per frame
		framingInfo.HasFrameHeaders = true
		framingInfo.FrameAlignment = "byte"
	case "mp3":
		framingInfo.FrameSize = 1152 // MP3 typically uses 1152 samples per frame
		framingInfo.HasFrameHeaders = true
		framingInfo.FrameAlignment = "byte"
	case "ac3":
		framingInfo.FrameSize = 1536 // AC-3 uses 1536 samples per frame
		framingInfo.HasFrameHeaders = true
		framingInfo.FrameAlignment = "byte"
	default:
		// Keep defaults
	}
}

func (awa *AudioWrappingAnalyzer) isATSCCompliantCodec(codec string) bool {
	compliantCodecs := []string{"ac3", "eac3", "aac"}
	for _, compliant := range compliantCodecs {
		if codec == compliant {
			return true
		}
	}
	return false
}

func (awa *AudioWrappingAnalyzer) isEBUCompliantWrapping(streamInfo *AudioWrappingInfo) bool {
	// EBU compliance requires specific wrapping formats
	if streamInfo.ProfessionalWrapping != nil {
		if streamInfo.ProfessionalWrapping.BWFFormat != nil || streamInfo.ProfessionalWrapping.AESFormat != nil {
			return true
		}
	}
	return false
}

func (awa *AudioWrappingAnalyzer) isAES67Compliant(streamInfo *AudioWrappingInfo) bool {
	// AES67 requires specific sample rates and formats
	if streamInfo.ProfessionalWrapping != nil && streamInfo.ProfessionalWrapping.AESFormat != nil {
		return streamInfo.ProfessionalWrapping.AESFormat.AESStandard == "AES67"
	}
	return false
}

func (awa *AudioWrappingAnalyzer) executeCommand(ctx context.Context, cmd []string) (string, error) {
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := executeFFprobeCommand(execCtx, cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
