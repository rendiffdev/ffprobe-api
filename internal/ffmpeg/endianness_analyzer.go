package ffmpeg

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// EndiannessAnalyzer handles endianness detection through binary analysis
type EndiannessAnalyzer struct {
	logger zerolog.Logger
}

// NewEndiannessAnalyzer creates a new endianness analyzer
func NewEndiannessAnalyzer(logger zerolog.Logger) *EndiannessAnalyzer {
	return &EndiannessAnalyzer{
		logger: logger,
	}
}

// EndiannessAnalysis contains comprehensive endianness analysis
type EndiannessAnalysis struct {
	ContainerEndianness    string                     `json:"container_endianness"`
	AudioStreamEndianness  map[int]*AudioEndianness   `json:"audio_stream_endianness,omitempty"`
	VideoStreamEndianness  map[int]*VideoEndianness   `json:"video_stream_endianness,omitempty"`
	RawDataEndianness      *RawDataEndianness         `json:"raw_data_endianness,omitempty"`
	FileHeaderAnalysis     *FileHeaderAnalysis        `json:"file_header_analysis,omitempty"`
	ByteOrderMarks         []ByteOrderMark            `json:"byte_order_marks,omitempty"`
	EndiannessValidation   *EndiannessValidation      `json:"endianness_validation,omitempty"`
	PlatformCompatibility  *PlatformCompatibility     `json:"platform_compatibility,omitempty"`
}

// AudioEndianness contains audio-specific endianness information
type AudioEndianness struct {
	StreamIndex      int                 `json:"stream_index"`
	SampleFormat     string              `json:"sample_format"`
	DetectedEndian   string              `json:"detected_endian"`    // "little", "big", "unknown"
	ByteOrder        string              `json:"byte_order"`         // "LE", "BE"
	BitsPerSample    int                 `json:"bits_per_sample"`
	IsPacked         bool                `json:"is_packed"`
	IsPlanar         bool                `json:"is_planar"`
	RequiresSwapping bool                `json:"requires_swapping"`
	ConfidenceLevel  float64             `json:"confidence_level"`
	DetectionMethod  string              `json:"detection_method"`
	Issues           []string            `json:"issues,omitempty"`
}

// VideoEndianness contains video-specific endianness information
type VideoEndianness struct {
	StreamIndex     int      `json:"stream_index"`
	PixelFormat     string   `json:"pixel_format"`
	DetectedEndian  string   `json:"detected_endian"`
	ByteOrder       string   `json:"byte_order"`
	BitsPerPixel    int      `json:"bits_per_pixel"`
	ColorComponents []string `json:"color_components,omitempty"`
	PackingFormat   string   `json:"packing_format"`
	ConfidenceLevel float64  `json:"confidence_level"`
	DetectionMethod string   `json:"detection_method"`
	Issues          []string `json:"issues,omitempty"`
}

// RawDataEndianness contains raw binary data endianness analysis
type RawDataEndianness struct {
	FileSignature       string      `json:"file_signature"`
	MagicNumbers        []MagicNumber `json:"magic_numbers,omitempty"`
	DetectedEndian      string      `json:"detected_endian"`
	SampleAnalysis      []SampleData `json:"sample_analysis,omitempty"`
	StatisticalAnalysis *StatisticalEndianness `json:"statistical_analysis,omitempty"`
	ConfidenceLevel     float64     `json:"confidence_level"`
}

// FileHeaderAnalysis contains file header-based endianness detection
type FileHeaderAnalysis struct {
	FileFormat       string        `json:"file_format"`
	HeaderEndianness string        `json:"header_endianness"`
	BOMs             []ByteOrderMark `json:"boms,omitempty"`
	FormatSpecific   map[string]interface{} `json:"format_specific,omitempty"`
	IsConsistent     bool          `json:"is_consistent"`
	Issues           []string      `json:"issues,omitempty"`
}

// ByteOrderMark represents detected byte order marks
type ByteOrderMark struct {
	Offset      int64   `json:"offset"`
	Pattern     string  `json:"pattern"`
	Type        string  `json:"type"`        // "BOM", "magic_number", "signature"
	Endianness  string  `json:"endianness"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// MagicNumber represents magic number patterns that indicate endianness
type MagicNumber struct {
	Offset      int64  `json:"offset"`
	Value       string `json:"value"`
	Format      string `json:"format"`
	Endianness  string `json:"endianness"`
	Description string `json:"description"`
}

// SampleData represents sample binary data analysis
type SampleData struct {
	Offset            int64   `json:"offset"`
	DataLength        int     `json:"data_length"`
	LittleEndianProb  float64 `json:"little_endian_probability"`
	BigEndianProb     float64 `json:"big_endian_probability"`
	DetectedPattern   string  `json:"detected_pattern"`
	AnalysisMethod    string  `json:"analysis_method"`
}

// StatisticalEndianness contains statistical analysis of byte patterns
type StatisticalEndianness struct {
	ByteFrequency      map[string]int `json:"byte_frequency"`
	SequencePatterns   []string       `json:"sequence_patterns,omitempty"`
	EntropyAnalysis    float64        `json:"entropy_analysis"`
	LittleEndianScore  float64        `json:"little_endian_score"`
	BigEndianScore     float64        `json:"big_endian_score"`
	NeutralScore       float64        `json:"neutral_score"`
}

// EndiannessValidation contains validation results
type EndiannessValidation struct {
	IsConsistent         bool     `json:"is_consistent"`
	HasConflicts         bool     `json:"has_conflicts"`
	ConflictingStreams   []int    `json:"conflicting_streams,omitempty"`
	Issues               []string `json:"issues,omitempty"`
	Warnings             []string `json:"warnings,omitempty"`
	Recommendations      []string `json:"recommendations,omitempty"`
	OverallConfidence    float64  `json:"overall_confidence"`
}

// PlatformCompatibility contains platform-specific compatibility information
type PlatformCompatibility struct {
	IntelCompatible  bool     `json:"intel_compatible"`     // x86/x64 (little endian)
	ARMCompatible    bool     `json:"arm_compatible"`       // ARM (can be both)
	PowerPCCompatible bool    `json:"powerpc_compatible"`   // PowerPC (big endian)
	SPARCCompatible  bool     `json:"sparc_compatible"`     // SPARC (big endian)
	MIPSCompatible   bool     `json:"mips_compatible"`      // MIPS (can be both)
	Issues           []string `json:"issues,omitempty"`
	Recommendations  []string `json:"recommendations,omitempty"`
}

// Known file signatures and their endianness
var fileSignatures = map[string]string{
	"ftyp":     "big",    // MP4 file type box (big endian)
	"RIFF":     "little", // RIFF header (little endian)
	"RIFX":     "big",    // RIFF big endian variant
	"FORM":     "big",    // IFF FORM (big endian)
	"\x1A\x45\xDF\xA3": "little", // Matroska/WebM EBML header
	"OggS":     "little", // Ogg header (little endian)
	"\xFF\xFB": "variable", // MP3 frame header (no inherent endianness)
	"\xFF\xFA": "variable", // MP3 frame header (no inherent endianness)
}

// Known magic numbers and their endianness implications
var magicNumbers = map[string]MagicNumber{
	"\x00\x00\x00\x20ftyp": {0, "00000020667479700", "MP4", "big", "MP4 file type box (big endian)"},
	"\x52\x49\x46\x46":     {0, "52494646", "RIFF", "little", "RIFF header (little endian)"},
	"\x52\x49\x46\x58":     {0, "52494658", "RIFX", "big", "RIFF header (big endian)"},
	"\x46\x4F\x52\x4D":     {0, "464F524D", "IFF", "big", "IFF FORM header (big endian)"},
	"\x1A\x45\xDF\xA3":     {0, "1A45DFA3", "EBML", "little", "EBML header (little endian)"},
	"\x4F\x67\x67\x53":     {0, "4F676753", "Ogg", "little", "Ogg page header (little endian)"},
	"\x4D\x54\x68\x64":     {0, "4D546864", "MIDI", "big", "MIDI header (big endian)"},
}

// PCM sample format endianness indicators
var sampleFormatEndianness = map[string]string{
	"s16le": "little",
	"s16be": "big",
	"s24le": "little", 
	"s24be": "big",
	"s32le": "little",
	"s32be": "big",
	"f32le": "little",
	"f32be": "big",
	"f64le": "little",
	"f64be": "big",
}

// Pixel format endianness indicators
var pixelFormatEndianness = map[string]string{
	"rgb565le":  "little",
	"rgb565be":  "big",
	"bgr565le":  "little",
	"bgr565be":  "big",
	"rgb555le":  "little",
	"rgb555be":  "big",
	"bgr555le":  "little",
	"bgr555be":  "big",
	"gray16le":  "little",
	"gray16be":  "big",
	"yuv420p16le": "little",
	"yuv420p16be": "big",
	"yuv422p16le": "little",
	"yuv422p16be": "big",
	"yuv444p16le": "little",
	"yuv444p16be": "big",
}

// AnalyzeEndianness performs comprehensive endianness analysis
func (ea *EndiannessAnalyzer) AnalyzeEndianness(ctx context.Context, filePath string, streams []StreamInfo, format *FormatInfo) (*EndiannessAnalysis, error) {
	analysis := &EndiannessAnalysis{
		AudioStreamEndianness: make(map[int]*AudioEndianness),
		VideoStreamEndianness: make(map[int]*VideoEndianness),
		ByteOrderMarks:        []ByteOrderMark{},
	}

	// Step 1: Analyze file header and container format
	if err := ea.analyzeFileHeader(filePath, analysis); err != nil {
		ea.logger.Warn().Err(err).Msg("Failed to analyze file header")
	}

	// Step 2: Analyze stream-specific endianness
	ea.analyzeStreamEndianness(streams, analysis)

	// Step 3: Perform raw binary analysis
	if err := ea.analyzeRawBinaryData(filePath, analysis); err != nil {
		ea.logger.Warn().Err(err).Msg("Failed to analyze raw binary data")
	}

	// Step 4: Detect byte order marks and magic numbers
	if err := ea.detectByteOrderMarks(filePath, analysis); err != nil {
		ea.logger.Warn().Err(err).Msg("Failed to detect byte order marks")
	}

	// Step 5: Determine container endianness
	ea.determineContainerEndianness(format, analysis)

	// Step 6: Validate endianness consistency
	analysis.EndiannessValidation = ea.validateEndianness(analysis)

	// Step 7: Analyze platform compatibility
	analysis.PlatformCompatibility = ea.analyzePlatformCompatibility(analysis)

	return analysis, nil
}

// analyzeFileHeader analyzes the file header for endianness indicators
func (ea *EndiannessAnalyzer) analyzeFileHeader(filePath string, analysis *EndiannessAnalysis) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 1KB of file for header analysis
	headerData := make([]byte, 1024)
	n, err := file.Read(headerData)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file header: %w", err)
	}
	headerData = headerData[:n]

	headerAnalysis := &FileHeaderAnalysis{
		FormatSpecific: make(map[string]interface{}),
		BOMs:           []ByteOrderMark{},
		Issues:         []string{},
		IsConsistent:   true,
	}

	// Detect file format from header
	headerAnalysis.FileFormat = ea.detectFileFormat(headerData)
	
	// Analyze format-specific endianness
	switch headerAnalysis.FileFormat {
	case "MP4":
		headerAnalysis.HeaderEndianness = ea.analyzeMP4Header(headerData)
	case "RIFF":
		headerAnalysis.HeaderEndianness = ea.analyzeRIFFHeader(headerData)
	case "Matroska":
		headerAnalysis.HeaderEndianness = ea.analyzeMatroskaHeader(headerData)
	case "Ogg":
		headerAnalysis.HeaderEndianness = "little"
	default:
		headerAnalysis.HeaderEndianness = "unknown"
	}

	analysis.FileHeaderAnalysis = headerAnalysis
	return nil
}

// analyzeStreamEndianness analyzes endianness for individual streams
func (ea *EndiannessAnalyzer) analyzeStreamEndianness(streams []StreamInfo, analysis *EndiannessAnalysis) {
	for _, stream := range streams {
		switch strings.ToLower(stream.CodecType) {
		case "audio":
			audioEndian := ea.analyzeAudioStreamEndianness(stream)
			if audioEndian != nil {
				analysis.AudioStreamEndianness[stream.Index] = audioEndian
			}

		case "video":
			videoEndian := ea.analyzeVideoStreamEndianness(stream)
			if videoEndian != nil {
				analysis.VideoStreamEndianness[stream.Index] = videoEndian
			}
		}
	}
}

// analyzeAudioStreamEndianness analyzes audio stream endianness
func (ea *EndiannessAnalyzer) analyzeAudioStreamEndianness(stream StreamInfo) *AudioEndianness {
	audioEndian := &AudioEndianness{
		StreamIndex:     stream.Index,
		SampleFormat:    stream.SampleFmt,
		BitsPerSample:   stream.BitsPerSample,
		ConfidenceLevel: 0.0,
		DetectionMethod: "sample_format_analysis",
		Issues:          []string{},
	}

	// Detect endianness from sample format
	if endian, exists := sampleFormatEndianness[strings.ToLower(stream.SampleFmt)]; exists {
		audioEndian.DetectedEndian = endian
		if endian == "little" {
			audioEndian.ByteOrder = "LE"
		} else {
			audioEndian.ByteOrder = "BE"
		}
		audioEndian.ConfidenceLevel = 0.95
	} else {
		// Try to infer from format name patterns
		sampleFmt := strings.ToLower(stream.SampleFmt)
		if strings.HasSuffix(sampleFmt, "le") {
			audioEndian.DetectedEndian = "little"
			audioEndian.ByteOrder = "LE"
			audioEndian.ConfidenceLevel = 0.9
		} else if strings.HasSuffix(sampleFmt, "be") {
			audioEndian.DetectedEndian = "big"
			audioEndian.ByteOrder = "BE"
			audioEndian.ConfidenceLevel = 0.9
		} else {
			audioEndian.DetectedEndian = "unknown"
			audioEndian.ConfidenceLevel = 0.0
			audioEndian.Issues = append(audioEndian.Issues, "Cannot determine endianness from sample format")
		}
	}

	// Analyze packed vs planar
	audioEndian.IsPlanar = strings.Contains(strings.ToLower(stream.SampleFmt), "p")
	audioEndian.IsPacked = !audioEndian.IsPlanar

	// Check if byte swapping would be required on current platform
	audioEndian.RequiresSwapping = ea.requiresByteSwapping(audioEndian.DetectedEndian)

	return audioEndian
}

// analyzeVideoStreamEndianness analyzes video stream endianness  
func (ea *EndiannessAnalyzer) analyzeVideoStreamEndianness(stream StreamInfo) *VideoEndianness {
	videoEndian := &VideoEndianness{
		StreamIndex:     stream.Index,
		PixelFormat:     stream.PixFmt,
		ConfidenceLevel: 0.0,
		DetectionMethod: "pixel_format_analysis",
		Issues:          []string{},
	}

	// Detect endianness from pixel format
	if endian, exists := pixelFormatEndianness[strings.ToLower(stream.PixFmt)]; exists {
		videoEndian.DetectedEndian = endian
		if endian == "little" {
			videoEndian.ByteOrder = "LE"
		} else {
			videoEndian.ByteOrder = "BE"
		}
		videoEndian.ConfidenceLevel = 0.95
	} else {
		// Try to infer from format name patterns
		pixFmt := strings.ToLower(stream.PixFmt)
		if strings.HasSuffix(pixFmt, "le") {
			videoEndian.DetectedEndian = "little"
			videoEndian.ByteOrder = "LE"
			videoEndian.ConfidenceLevel = 0.9
		} else if strings.HasSuffix(pixFmt, "be") {
			videoEndian.DetectedEndian = "big"
			videoEndian.ByteOrder = "BE"
			videoEndian.ConfidenceLevel = 0.9
		} else {
			// Most common pixel formats are implicitly little endian or endian-neutral
			videoEndian.DetectedEndian = "neutral"
			videoEndian.ConfidenceLevel = 0.3
		}
	}

	// Analyze pixel format characteristics
	videoEndian.PackingFormat = ea.determinePackingFormat(stream.PixFmt)
	videoEndian.ColorComponents = ea.extractColorComponents(stream.PixFmt)

	return videoEndian
}

// analyzeRawBinaryData performs statistical analysis of raw binary data
func (ea *EndiannessAnalyzer) analyzeRawBinaryData(filePath string, analysis *EndiannessAnalysis) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for binary analysis: %w", err)
	}
	defer file.Close()

	// Sample data from different parts of the file
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()
	sampleSize := int64(1024) // 1KB samples
	sampleCount := 5

	rawData := &RawDataEndianness{
		SampleAnalysis: []SampleData{},
	}

	// Take samples from beginning, middle, and end of file
	offsets := []int64{
		0,                           // Beginning
		fileSize / 4,                // 25%
		fileSize / 2,                // Middle
		fileSize * 3 / 4,            // 75%
		fileSize - sampleSize,       // End
	}

	var allData []byte
	for i, offset := range offsets {
		if offset < 0 {
			offset = 0
		}
		if offset+sampleSize > fileSize {
			sampleSize = fileSize - offset
		}

		file.Seek(offset, 0)
		sampleData := make([]byte, sampleSize)
		n, err := file.Read(sampleData)
		if err != nil && err != io.EOF {
			continue
		}
		sampleData = sampleData[:n]
		allData = append(allData, sampleData...)

		// Analyze this sample
		sample := ea.analyzeSampleEndianness(sampleData, offset)
		rawData.SampleAnalysis = append(rawData.SampleAnalysis, sample)

		if i >= sampleCount-1 {
			break
		}
	}

	// Perform statistical analysis on all collected data
	rawData.StatisticalAnalysis = ea.performStatisticalAnalysis(allData)

	// Determine overall endianness from statistical analysis
	if rawData.StatisticalAnalysis.LittleEndianScore > rawData.StatisticalAnalysis.BigEndianScore {
		rawData.DetectedEndian = "little"
		rawData.ConfidenceLevel = rawData.StatisticalAnalysis.LittleEndianScore
	} else if rawData.StatisticalAnalysis.BigEndianScore > rawData.StatisticalAnalysis.LittleEndianScore {
		rawData.DetectedEndian = "big"
		rawData.ConfidenceLevel = rawData.StatisticalAnalysis.BigEndianScore
	} else {
		rawData.DetectedEndian = "neutral"
		rawData.ConfidenceLevel = rawData.StatisticalAnalysis.NeutralScore
	}

	analysis.RawDataEndianness = rawData
	return nil
}

// detectByteOrderMarks detects byte order marks and magic numbers
func (ea *EndiannessAnalyzer) detectByteOrderMarks(filePath string, analysis *EndiannessAnalysis) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for BOM detection: %w", err)
	}
	defer file.Close()

	// Read first 64KB for comprehensive magic number detection
	headerData := make([]byte, 65536)
	n, err := file.Read(headerData)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file for BOM detection: %w", err)
	}
	headerData = headerData[:n]

	// Search for known magic numbers
	for pattern, magic := range magicNumbers {
		if index := bytes.Index(headerData, []byte(pattern)); index >= 0 {
			bom := ByteOrderMark{
				Offset:      int64(index),
				Pattern:     magic.Value,
				Type:        "magic_number",
				Endianness:  magic.Endianness,
				Description: magic.Description,
				Confidence:  0.9,
			}
			analysis.ByteOrderMarks = append(analysis.ByteOrderMarks, bom)
		}
	}

	// Search for UTF BOMs
	utfBOMs := map[string]ByteOrderMark{
		"\xFF\xFE":     {0, "FFFE", "utf16_le_bom", "little", "UTF-16 Little Endian BOM", 1.0},
		"\xFE\xFF":     {0, "FEFF", "utf16_be_bom", "big", "UTF-16 Big Endian BOM", 1.0},
		"\xFF\xFE\x00\x00": {0, "FFFE0000", "utf32_le_bom", "little", "UTF-32 Little Endian BOM", 1.0},
		"\x00\x00\xFE\xFF": {0, "0000FEFF", "utf32_be_bom", "big", "UTF-32 Big Endian BOM", 1.0},
	}

	for pattern, bom := range utfBOMs {
		if bytes.HasPrefix(headerData, []byte(pattern)) {
			analysis.ByteOrderMarks = append(analysis.ByteOrderMarks, bom)
		}
	}

	return nil
}

// determineContainerEndianness determines overall container endianness
func (ea *EndiannessAnalyzer) determineContainerEndianness(format *FormatInfo, analysis *EndiannessAnalysis) {
	// Priority: File header > Magic numbers > Stream analysis
	
	if analysis.FileHeaderAnalysis != nil && analysis.FileHeaderAnalysis.HeaderEndianness != "unknown" {
		analysis.ContainerEndianness = analysis.FileHeaderAnalysis.HeaderEndianness
		return
	}

	// Check byte order marks
	for _, bom := range analysis.ByteOrderMarks {
		if bom.Type == "magic_number" && bom.Confidence > 0.8 {
			analysis.ContainerEndianness = bom.Endianness
			return
		}
	}

	// Check format name
	if format != nil {
		formatName := strings.ToLower(format.FormatName)
		if strings.Contains(formatName, "mp4") || strings.Contains(formatName, "mov") {
			analysis.ContainerEndianness = "big"
		} else if strings.Contains(formatName, "avi") || strings.Contains(formatName, "wav") {
			analysis.ContainerEndianness = "little"
		} else if strings.Contains(formatName, "matroska") || strings.Contains(formatName, "webm") {
			analysis.ContainerEndianness = "little"
		} else {
			analysis.ContainerEndianness = "unknown"
		}
	} else {
		analysis.ContainerEndianness = "unknown"
	}
}

// validateEndianness validates endianness consistency across streams
func (ea *EndiannessAnalyzer) validateEndianness(analysis *EndiannessAnalysis) *EndiannessValidation {
	validation := &EndiannessValidation{
		IsConsistent:      true,
		HasConflicts:      false,
		ConflictingStreams: []int{},
		Issues:            []string{},
		Warnings:          []string{},
		Recommendations:   []string{},
		OverallConfidence: 0.0,
	}

	// Check consistency between container and streams
	containerEndian := analysis.ContainerEndianness
	
	var totalConfidence float64
	var confidenceCount int

	// Validate audio streams
	for _, audioStream := range analysis.AudioStreamEndianness {
		if audioStream.DetectedEndian != "unknown" && audioStream.DetectedEndian != containerEndian {
			if containerEndian != "unknown" {
				validation.HasConflicts = true
				validation.ConflictingStreams = append(validation.ConflictingStreams, audioStream.StreamIndex)
				validation.Issues = append(validation.Issues, 
					fmt.Sprintf("Audio stream %d endianness (%s) conflicts with container (%s)", 
						audioStream.StreamIndex, audioStream.DetectedEndian, containerEndian))
			}
		}
		totalConfidence += audioStream.ConfidenceLevel
		confidenceCount++
	}

	// Validate video streams
	for _, videoStream := range analysis.VideoStreamEndianness {
		if videoStream.DetectedEndian != "unknown" && videoStream.DetectedEndian != "neutral" && 
		   videoStream.DetectedEndian != containerEndian {
			if containerEndian != "unknown" {
				validation.HasConflicts = true
				validation.ConflictingStreams = append(validation.ConflictingStreams, videoStream.StreamIndex)
				validation.Issues = append(validation.Issues, 
					fmt.Sprintf("Video stream %d endianness (%s) conflicts with container (%s)", 
						videoStream.StreamIndex, videoStream.DetectedEndian, containerEndian))
			}
		}
		totalConfidence += videoStream.ConfidenceLevel
		confidenceCount++
	}

	// Calculate overall confidence
	if confidenceCount > 0 {
		validation.OverallConfidence = totalConfidence / float64(confidenceCount)
	}

	// Check if conflicts exist
	if validation.HasConflicts {
		validation.IsConsistent = false
		validation.Recommendations = append(validation.Recommendations, 
			"Verify endianness consistency across all streams")
	}

	// Add general recommendations
	if validation.OverallConfidence < 0.7 {
		validation.Warnings = append(validation.Warnings, "Low confidence in endianness detection")
		validation.Recommendations = append(validation.Recommendations, 
			"Consider using tools with deeper binary analysis capabilities")
	}

	return validation
}

// analyzePlatformCompatibility analyzes platform-specific compatibility
func (ea *EndiannessAnalyzer) analyzePlatformCompatibility(analysis *EndiannessAnalysis) *PlatformCompatibility {
	compat := &PlatformCompatibility{
		Issues:          []string{},
		Recommendations: []string{},
	}

	containerEndian := analysis.ContainerEndianness

	// Intel/AMD x86/x64 (Little Endian)
	compat.IntelCompatible = (containerEndian == "little" || containerEndian == "unknown")

	// ARM (can be both, typically little endian in modern implementations)
	compat.ARMCompatible = true

	// PowerPC (Big Endian)
	compat.PowerPCCompatible = (containerEndian == "big" || containerEndian == "unknown")

	// SPARC (Big Endian)
	compat.SPARCCompatible = (containerEndian == "big" || containerEndian == "unknown")

	// MIPS (can be both)
	compat.MIPSCompatible = true

	// Generate compatibility issues and recommendations
	if containerEndian == "big" {
		compat.Issues = append(compat.Issues, "Big endian format may require byte swapping on Intel/AMD platforms")
		compat.Recommendations = append(compat.Recommendations, "Consider little endian formats for better x86/x64 performance")
	} else if containerEndian == "little" {
		compat.Issues = append(compat.Issues, "Little endian format may require byte swapping on PowerPC/SPARC platforms")
	}

	return compat
}

// Helper methods

func (ea *EndiannessAnalyzer) detectFileFormat(headerData []byte) string {
	if len(headerData) < 8 {
		return "unknown"
	}

	// Check for common file signatures
	if bytes.HasPrefix(headerData[4:], []byte("ftyp")) {
		return "MP4"
	}
	if bytes.HasPrefix(headerData, []byte("RIFF")) || bytes.HasPrefix(headerData, []byte("RIFX")) {
		return "RIFF"
	}
	if bytes.HasPrefix(headerData, []byte("\x1A\x45\xDF\xA3")) {
		return "Matroska"
	}
	if bytes.HasPrefix(headerData, []byte("OggS")) {
		return "Ogg"
	}
	if bytes.HasPrefix(headerData, []byte("FORM")) {
		return "IFF"
	}

	return "unknown"
}

func (ea *EndiannessAnalyzer) analyzeMP4Header(headerData []byte) string {
	// MP4 uses big endian for box headers
	return "big"
}

func (ea *EndiannessAnalyzer) analyzeRIFFHeader(headerData []byte) string {
	// RIFF uses little endian, RIFX uses big endian
	if bytes.HasPrefix(headerData, []byte("RIFX")) {
		return "big"
	}
	return "little"
}

func (ea *EndiannessAnalyzer) analyzeMatroskaHeader(headerData []byte) string {
	// Matroska/WebM EBML uses little endian for most fields
	return "little"
}

func (ea *EndiannessAnalyzer) requiresByteSwapping(endianness string) bool {
	// Assuming we're running on a little endian system (most common)
	return endianness == "big"
}

func (ea *EndiannessAnalyzer) determinePackingFormat(pixelFormat string) string {
	pixFmt := strings.ToLower(pixelFormat)
	
	if strings.Contains(pixFmt, "packed") {
		return "packed"
	} else if strings.Contains(pixFmt, "planar") || strings.HasSuffix(pixFmt, "p") {
		return "planar"
	} else if strings.Contains(pixFmt, "rgb") || strings.Contains(pixFmt, "bgr") {
		return "interleaved"
	}
	
	return "unknown"
}

func (ea *EndiannessAnalyzer) extractColorComponents(pixelFormat string) []string {
	pixFmt := strings.ToLower(pixelFormat)
	
	if strings.Contains(pixFmt, "yuv") {
		return []string{"Y", "U", "V"}
	} else if strings.Contains(pixFmt, "rgb") {
		return []string{"R", "G", "B"}
	} else if strings.Contains(pixFmt, "bgr") {
		return []string{"B", "G", "R"}
	} else if strings.Contains(pixFmt, "gray") {
		return []string{"Y"}
	}
	
	return []string{"unknown"}
}

func (ea *EndiannessAnalyzer) analyzeSampleEndianness(data []byte, offset int64) SampleData {
	sample := SampleData{
		Offset:          offset,
		DataLength:      len(data),
		AnalysisMethod:  "statistical_analysis",
	}

	if len(data) < 4 {
		return sample
	}

	// Simple heuristic: look for patterns that suggest endianness
	littleEndianScore := 0.0
	bigEndianScore := 0.0

	// Analyze 16-bit and 32-bit word patterns
	for i := 0; i < len(data)-3; i += 4 {
		// Read as little endian 32-bit
		leVal := binary.LittleEndian.Uint32(data[i:i+4])
		// Read as big endian 32-bit  
		beVal := binary.BigEndian.Uint32(data[i:i+4])

		// Score based on how "natural" the values look
		littleEndianScore += ea.scoreNaturalness(leVal)
		bigEndianScore += ea.scoreNaturalness(beVal)
	}

	total := littleEndianScore + bigEndianScore
	if total > 0 {
		sample.LittleEndianProb = littleEndianScore / total
		sample.BigEndianProb = bigEndianScore / total
	}

	// Determine detected pattern
	if sample.LittleEndianProb > sample.BigEndianProb {
		sample.DetectedPattern = "little_endian_likely"
	} else if sample.BigEndianProb > sample.LittleEndianProb {
		sample.DetectedPattern = "big_endian_likely"
	} else {
		sample.DetectedPattern = "neutral"
	}

	return sample
}

func (ea *EndiannessAnalyzer) performStatisticalAnalysis(data []byte) *StatisticalEndianness {
	stats := &StatisticalEndianness{
		ByteFrequency:    make(map[string]int),
		SequencePatterns: []string{},
	}

	if len(data) == 0 {
		return stats
	}

	// Calculate byte frequency
	for _, b := range data {
		key := fmt.Sprintf("%02X", b)
		stats.ByteFrequency[key]++
	}

	// Calculate entropy
	stats.EntropyAnalysis = ea.calculateEntropy(data)

	// Analyze endianness patterns
	littleScore := 0.0
	bigScore := 0.0
	neutralScore := 0.0

	// Look for endianness-indicating patterns
	for i := 0; i < len(data)-3; i += 2 {
		if i+3 < len(data) {
			// Check 16-bit patterns
			leVal := binary.LittleEndian.Uint16(data[i:i+2])
			beVal := binary.BigEndian.Uint16(data[i:i+2])

			littleScore += ea.scoreNaturalness(uint32(leVal))
			bigScore += ea.scoreNaturalness(uint32(beVal))
		}
	}

	// Normalize scores
	total := littleScore + bigScore
	if total > 0 {
		stats.LittleEndianScore = littleScore / total
		stats.BigEndianScore = bigScore / total
	} else {
		stats.NeutralScore = 1.0
	}

	return stats
}

func (ea *EndiannessAnalyzer) calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	frequency := make(map[byte]int)
	for _, b := range data {
		frequency[b]++
	}

	entropy := 0.0
	length := float64(len(data))

	for _, count := range frequency {
		if count > 0 {
			prob := float64(count) / length
			entropy -= prob * ea.log2(prob)
		}
	}

	return entropy
}

func (ea *EndiannessAnalyzer) scoreNaturalness(value uint32) float64 {
	// Simple heuristic to score how "natural" a value looks
	// Lower values and powers of 2 are more common in structured data
	
	if value == 0 {
		return 1.0
	}
	
	// Score based on magnitude (smaller values are more common)
	magnitudeScore := 1.0 / (1.0 + float64(value)/1000000.0)
	
	// Score based on whether it's a power of 2 or common value
	commonValueScore := 0.0
	if value&(value-1) == 0 { // Power of 2
		commonValueScore = 0.5
	}
	
	// Check for common file format values
	if value == 0x00000020 || value == 0x66747970 || value == 0x52494646 {
		commonValueScore = 1.0
	}
	
	return magnitudeScore + commonValueScore
}

// math.Log2 helper function for older Go versions
func (ea *EndiannessAnalyzer) log2(x float64) float64 {
	return math.Log(x) / math.Log(2)
}
