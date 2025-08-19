package ffmpeg

import (
	"time"
)

// FFprobeOptions contains all possible ffprobe command options
type FFprobeOptions struct {
	// Input options
	Input        string            `json:"input"`
	Format       string            `json:"format,omitempty"`        // -f
	InputOptions map[string]string `json:"input_options,omitempty"` // Additional input options

	// Output format options
	OutputFormat    OutputFormat `json:"output_format,omitempty"`     // -of
	PrettyPrint     bool         `json:"pretty_print,omitempty"`      // -pretty
	ShowFormat      bool         `json:"show_format,omitempty"`       // -show_format
	ShowStreams     bool         `json:"show_streams,omitempty"`      // -show_streams
	ShowPackets     bool         `json:"show_packets,omitempty"`      // -show_packets
	ShowFrames      bool         `json:"show_frames,omitempty"`       // -show_frames
	ShowChapters    bool         `json:"show_chapters,omitempty"`     // -show_chapters
	ShowPrograms    bool         `json:"show_programs,omitempty"`     // -show_programs
	ShowError       bool         `json:"show_error,omitempty"`        // -show_error
	ShowData        bool         `json:"show_data,omitempty"`         // -show_data
	ShowDataHash    bool         `json:"show_data_hash,omitempty"`    // -show_data_hash
	ShowPrivateData bool         `json:"show_private_data,omitempty"` // -show_private_data

	// Selection options
	SelectStreams string `json:"select_streams,omitempty"` // -select_streams
	ReadIntervals string `json:"read_intervals,omitempty"` // -read_intervals
	ShowEntries   string `json:"show_entries,omitempty"`   // -show_entries

	// Processing options
	CountFrames     bool  `json:"count_frames,omitempty"`     // -count_frames
	CountPackets    bool  `json:"count_packets,omitempty"`    // -count_packets
	ProbeSize       int64 `json:"probe_size,omitempty"`       // -probesize
	AnalyzeDuration int64 `json:"analyze_duration,omitempty"` // -analyzeduration

	// Error detection options
	ErrorDetect       string `json:"error_detect,omitempty"`   // -err_detect
	FormatErrorDetect string `json:"f_error_detect,omitempty"` // -f_err_detect
	HashAlgorithm     string `json:"hash,omitempty"`           // -hash (for show_data_hash)

	// Logging options
	LogLevel   LogLevel `json:"log_level,omitempty"`   // -loglevel
	HideBanner bool     `json:"hide_banner,omitempty"` // -hide_banner
	Report     bool     `json:"report,omitempty"`      // -report

	// Processing limits
	Timeout       time.Duration `json:"timeout,omitempty"`         // Custom timeout
	MaxOutputSize int64         `json:"max_output_size,omitempty"` // Custom limit

	// Custom arguments
	Args []string `json:"args,omitempty"` // Custom FFprobe arguments
}

// OutputFormat represents ffprobe output formats
type OutputFormat string

const (
	OutputDefault OutputFormat = "default"
	OutputCompact OutputFormat = "compact"
	OutputCSV     OutputFormat = "csv"
	OutputFlat    OutputFormat = "flat"
	OutputINI     OutputFormat = "ini"
	OutputJSON    OutputFormat = "json"
	OutputXML     OutputFormat = "xml"
)

// LogLevel represents ffprobe log levels
type LogLevel string

const (
	LogQuiet   LogLevel = "quiet"
	LogPanic   LogLevel = "panic"
	LogFatal   LogLevel = "fatal"
	LogError   LogLevel = "error"
	LogWarning LogLevel = "warning"
	LogInfo    LogLevel = "info"
	LogVerbose LogLevel = "verbose"
	LogDebug   LogLevel = "debug"
	LogTrace   LogLevel = "trace"
)

// FFprobeResult contains the result of ffprobe execution
type FFprobeResult struct {
	// Output data
	Output   string        `json:"output"`
	Format   *FormatInfo   `json:"format,omitempty"`
	Streams  []StreamInfo  `json:"streams,omitempty"`
	Packets  []PacketInfo  `json:"packets,omitempty"`
	Frames   []FrameInfo   `json:"frames,omitempty"`
	Chapters []ChapterInfo `json:"chapters,omitempty"`
	Programs []ProgramInfo `json:"programs,omitempty"`
	Error    *ErrorInfo    `json:"error,omitempty"`

	// Enhanced analysis data
	EnhancedAnalysis *EnhancedAnalysis `json:"enhanced_analysis,omitempty"`

	// Execution metadata
	Command       []string      `json:"command"`
	ExecutionTime time.Duration `json:"execution_time"`
	Success       bool          `json:"success"`
	ExitCode      int           `json:"exit_code"`
	StdErr        string        `json:"stderr,omitempty"`
}

// FormatInfo represents container/format information
type FormatInfo struct {
	Filename       string            `json:"filename"`
	NBStreams      int               `json:"nb_streams"`
	NBPrograms     int               `json:"nb_programs"`
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	StartTime      string            `json:"start_time,omitempty"`
	Duration       string            `json:"duration,omitempty"`
	Size           string            `json:"size,omitempty"`
	BitRate        string            `json:"bit_rate,omitempty"`
	ProbeScore     int               `json:"probe_score"`
	Tags           map[string]string `json:"tags,omitempty"`
}

// StreamInfo represents stream information
type StreamInfo struct {
	Index              int                    `json:"index"`
	CodecName          string                 `json:"codec_name"`
	CodecLongName      string                 `json:"codec_long_name"`
	Profile            string                 `json:"profile,omitempty"`
	CodecType          string                 `json:"codec_type"`
	CodecTimeBase      string                 `json:"codec_time_base,omitempty"`
	CodecTagString     string                 `json:"codec_tag_string,omitempty"`
	CodecTag           string                 `json:"codec_tag,omitempty"`
	Width              int                    `json:"width,omitempty"`
	Height             int                    `json:"height,omitempty"`
	CodedWidth         int                    `json:"coded_width,omitempty"`
	CodedHeight        int                    `json:"coded_height,omitempty"`
	ClosedCaptions     int                    `json:"closed_captions,omitempty"`
	HasBFrames         int                    `json:"has_b_frames,omitempty"`
	SampleAspectRatio  string                 `json:"sample_aspect_ratio,omitempty"`
	DisplayAspectRatio string                 `json:"display_aspect_ratio,omitempty"`
	PixFmt             string                 `json:"pix_fmt,omitempty"`
	Level              int                    `json:"level,omitempty"`
	ColorRange         string                 `json:"color_range,omitempty"`
	ColorSpace         string                 `json:"color_space,omitempty"`
	ColorTransfer      string                 `json:"color_transfer,omitempty"`
	ColorPrimaries     string                 `json:"color_primaries,omitempty"`
	ChromaLocation     string                 `json:"chroma_location,omitempty"`
	FieldOrder         string                 `json:"field_order,omitempty"`
	Refs               int                    `json:"refs,omitempty"`
	IsAVC              string                 `json:"is_avc,omitempty"`
	NALLengthSize      string                 `json:"nal_length_size,omitempty"`
	RFrameRate         string                 `json:"r_frame_rate,omitempty"`
	AvgFrameRate       string                 `json:"avg_frame_rate,omitempty"`
	TimeBase           string                 `json:"time_base,omitempty"`
	StartPts           int64                  `json:"start_pts,omitempty"`
	StartTime          string                 `json:"start_time,omitempty"`
	DurationTs         int64                  `json:"duration_ts,omitempty"`
	Duration           string                 `json:"duration,omitempty"`
	BitRate            string                 `json:"bit_rate,omitempty"`
	BitsPerRawSample   string                 `json:"bits_per_raw_sample,omitempty"`
	NBFrames           string                 `json:"nb_frames,omitempty"`
	SampleFmt          string                 `json:"sample_fmt,omitempty"`
	SampleRate         string                 `json:"sample_rate,omitempty"`
	Channels           int                    `json:"channels,omitempty"`
	ChannelLayout      string                 `json:"channel_layout,omitempty"`
	BitsPerSample      int                    `json:"bits_per_sample,omitempty"`
	Disposition        map[string]int         `json:"disposition,omitempty"`
	Tags               map[string]string      `json:"tags,omitempty"`
	ExtraData          map[string]interface{} `json:"extradata,omitempty"`
}

// PacketInfo represents packet information
type PacketInfo struct {
	CodecType    string `json:"codec_type"`
	StreamIndex  int    `json:"stream_index"`
	Pts          int64  `json:"pts"`
	PtsTime      string `json:"pts_time"`
	Dts          int64  `json:"dts"`
	DtsTime      string `json:"dts_time"`
	Duration     int64  `json:"duration"`
	DurationTime string `json:"duration_time"`
	Size         string `json:"size"`
	Pos          string `json:"pos"`
	Flags        string `json:"flags"`
}

// FrameInfo represents frame information
type FrameInfo struct {
	MediaType               string            `json:"media_type"`
	StreamIndex             int               `json:"stream_index"`
	KeyFrame                int               `json:"key_frame"`
	Pts                     int64             `json:"pts"`
	PtsTime                 string            `json:"pts_time"`
	PktPts                  int64             `json:"pkt_pts"`
	PktPtsTime              string            `json:"pkt_pts_time"`
	PktDts                  int64             `json:"pkt_dts"`
	PktDtsTime              string            `json:"pkt_dts_time"`
	BestEffortTimestamp     int64             `json:"best_effort_timestamp"`
	BestEffortTimestampTime string            `json:"best_effort_timestamp_time"`
	PktDuration             int64             `json:"pkt_duration"`
	PktDurationTime         string            `json:"pkt_duration_time"`
	PktPos                  string            `json:"pkt_pos"`
	PktSize                 string            `json:"pkt_size"`
	Width                   int               `json:"width,omitempty"`
	Height                  int               `json:"height,omitempty"`
	PixFmt                  string            `json:"pix_fmt,omitempty"`
	SampleAspectRatio       string            `json:"sample_aspect_ratio,omitempty"`
	PictType                string            `json:"pict_type,omitempty"`
	CodedPictureNumber      int               `json:"coded_picture_number,omitempty"`
	DisplayPictureNumber    int               `json:"display_picture_number,omitempty"`
	InterlacedFrame         int               `json:"interlaced_frame,omitempty"`
	TopFieldFirst           int               `json:"top_field_first,omitempty"`
	RepeatPict              int               `json:"repeat_pict,omitempty"`
	SampleFmt               string            `json:"sample_fmt,omitempty"`
	NBSamples               int               `json:"nb_samples,omitempty"`
	Channels                int               `json:"channels,omitempty"`
	ChannelLayout           string            `json:"channel_layout,omitempty"`
	Tags                    map[string]string `json:"tags,omitempty"`
}

// ChapterInfo represents chapter information
type ChapterInfo struct {
	ID        int               `json:"id"`
	TimeBase  string            `json:"time_base"`
	Start     int64             `json:"start"`
	StartTime string            `json:"start_time"`
	End       int64             `json:"end"`
	EndTime   string            `json:"end_time"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// ProgramInfo represents program information
type ProgramInfo struct {
	ProgramID  int               `json:"program_id"`
	ProgramNum int               `json:"program_num"`
	NBStreams  int               `json:"nb_streams"`
	PmtPID     int               `json:"pmt_pid"`
	PcrPID     int               `json:"pcr_pid"`
	StartPts   int64             `json:"start_pts"`
	StartTime  string            `json:"start_time"`
	EndPts     int64             `json:"end_pts"`
	EndTime    string            `json:"end_time"`
	Tags       map[string]string `json:"tags,omitempty"`
	Streams    []int             `json:"streams,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code   int    `json:"code"`
	String string `json:"string"`
}

// StreamSpecifier represents stream selection options
type StreamSpecifier struct {
	Type      string `json:"type,omitempty"`       // v, a, s, d (video, audio, subtitle, data)
	Index     int    `json:"index,omitempty"`      // Stream index
	ProgramID int    `json:"program_id,omitempty"` // Program ID
	Metadata  string `json:"metadata,omitempty"`   // Metadata tag
	Usable    bool   `json:"usable,omitempty"`     // Only usable streams
}

// ReadInterval represents time interval specification
type ReadInterval struct {
	Start    string `json:"start,omitempty"`    // Start time/percentage
	End      string `json:"end,omitempty"`      // End time/percentage
	Duration string `json:"duration,omitempty"` // Duration
}

// EnhancedAnalysis contains additional quality control checks
type EnhancedAnalysis struct {
	StreamCounts              *StreamCounts              `json:"stream_counts,omitempty"`
	VideoAnalysis             *VideoAnalysis             `json:"video_analysis,omitempty"`
	AudioAnalysis             *AudioAnalysis             `json:"audio_analysis,omitempty"`
	GOPAnalysis               *GOPAnalysis               `json:"gop_analysis,omitempty"`
	FrameStatistics           *FrameStatistics           `json:"frame_statistics,omitempty"`
	ContentAnalysis           *ContentAnalysis           `json:"content_analysis,omitempty"`
	BitDepthAnalysis          *BitDepthAnalysis          `json:"bit_depth_analysis,omitempty"`
	ResolutionAnalysis        *ResolutionAnalysis        `json:"resolution_analysis,omitempty"`
	FrameRateAnalysis         *FrameRateAnalysis         `json:"frame_rate_analysis,omitempty"`
	CodecAnalysis             *CodecAnalysis             `json:"codec_analysis,omitempty"`
	ContainerAnalysis         *ContainerAnalysis         `json:"container_analysis,omitempty"`
	LLMReport                 *LLMEnhancedReport         `json:"llm_report,omitempty"`
	TimecodeAnalysis          *TimecodeAnalysis          `json:"timecode_analysis,omitempty"`
	AFDAnalysis               *AFDAnalysis               `json:"afd_analysis,omitempty"`
	TransportStreamAnalysis   *TransportStreamAnalysis   `json:"transport_stream_analysis,omitempty"`
	EndiannessAnalysis        *EndiannessAnalysis        `json:"endianness_analysis,omitempty"`
	AudioWrappingAnalysis     *AudioWrappingAnalysis     `json:"audio_wrapping_analysis,omitempty"`
	IMFAnalysis               *IMFAnalysis               `json:"imf_analysis,omitempty"`
	MXFAnalysis               *MXFAnalysis               `json:"mxf_analysis,omitempty"`
	DeadPixelAnalysis         *DeadPixelAnalysis         `json:"dead_pixel_analysis,omitempty"`
	PSEAnalysis               *PSEAnalysis               `json:"pse_analysis,omitempty"`
	StreamDispositionAnalysis *StreamDispositionAnalysis `json:"stream_disposition_analysis,omitempty"`
	DataIntegrityAnalysis     *DataIntegrityAnalysis     `json:"data_integrity_analysis,omitempty"`
}

// StreamCounts provides detailed stream counting
type StreamCounts struct {
	TotalStreams      int `json:"total_streams"`
	VideoStreams      int `json:"video_streams"`
	AudioStreams      int `json:"audio_streams"`
	SubtitleStreams   int `json:"subtitle_streams"`
	DataStreams       int `json:"data_streams"`
	AttachmentStreams int `json:"attachment_streams"`
}

// VideoAnalysis provides enhanced video analysis
type VideoAnalysis struct {
	ChromaSubsampling  *string `json:"chroma_subsampling,omitempty"`
	MatrixCoefficients *string `json:"matrix_coefficients,omitempty"`
	BitRateMode        *string `json:"bit_rate_mode,omitempty"`
	HasClosedCaptions  bool    `json:"has_closed_captions"`
}

// AudioAnalysis provides enhanced audio analysis
type AudioAnalysis struct {
	BitRateMode *string `json:"bit_rate_mode,omitempty"`
}

// GOPAnalysis provides Group of Pictures analysis
type GOPAnalysis struct {
	AverageGOPSize  *float64 `json:"average_gop_size,omitempty"`
	MaxGOPSize      *int     `json:"max_gop_size,omitempty"`
	MinGOPSize      *int     `json:"min_gop_size,omitempty"`
	KeyFrameCount   int      `json:"keyframe_count"`
	TotalFrameCount int      `json:"total_frame_count"`
	GOPPattern      *string  `json:"gop_pattern,omitempty"`
}

// FrameStatistics provides comprehensive frame-level statistics
type FrameStatistics struct {
	TotalFrames      int            `json:"total_frames"`
	IFrames          int            `json:"i_frames"`
	PFrames          int            `json:"p_frames"`
	BFrames          int            `json:"b_frames"`
	FrameTypes       map[string]int `json:"frame_types,omitempty"`
	AverageFrameSize *float64       `json:"average_frame_size,omitempty"`
	MaxFrameSize     *int64         `json:"max_frame_size,omitempty"`
	MinFrameSize     *int64         `json:"min_frame_size,omitempty"`
}

// ContentAnalysis provides content-based quality analysis
type ContentAnalysis struct {
	BlackFrames   *BlackFrameAnalysis    `json:"black_frames,omitempty"`
	FreezeFrames  *FreezeFrameAnalysis   `json:"freeze_frames,omitempty"`
	AudioClipping *AudioClippingAnalysis `json:"audio_clipping,omitempty"`
	Blockiness    *BlockinessAnalysis    `json:"blockiness,omitempty"`
	Blurriness    *BlurrinessAnalysis    `json:"blurriness,omitempty"`
	InterlaceInfo *InterlaceAnalysis     `json:"interlace_info,omitempty"`
	NoiseLevel    *NoiseAnalysis         `json:"noise_level,omitempty"`
	LoudnessMeter *LoudnessAnalysis      `json:"loudness_meter,omitempty"`
	HDRAnalysis   *HDRAnalysis           `json:"hdr_analysis,omitempty"`
}

// BlackFrameAnalysis detects black or nearly black frames
type BlackFrameAnalysis struct {
	DetectedFrames int     `json:"detected_frames"`
	Percentage     float64 `json:"percentage"`
	Threshold      float64 `json:"threshold"`
}

// FreezeFrameAnalysis detects static/frozen frames
type FreezeFrameAnalysis struct {
	DetectedFrames int     `json:"detected_frames"`
	Percentage     float64 `json:"percentage"`
	Threshold      float64 `json:"threshold"`
}

// AudioClippingAnalysis detects audio clipping
type AudioClippingAnalysis struct {
	ClippedSamples int     `json:"clipped_samples"`
	Percentage     float64 `json:"percentage"`
	PeakLevel      float64 `json:"peak_level_db"`
}

// BlockinessAnalysis measures compression blockiness
type BlockinessAnalysis struct {
	AverageBlockiness float64 `json:"average_blockiness"`
	MaxBlockiness     float64 `json:"max_blockiness"`
	Threshold         float64 `json:"threshold"`
}

// BlurrinessAnalysis measures image sharpness
type BlurrinessAnalysis struct {
	AverageSharpness float64 `json:"average_sharpness"`
	MinSharpness     float64 `json:"min_sharpness"`
	BlurDetected     bool    `json:"blur_detected"`
}

// InterlaceAnalysis detects interlacing artifacts
type InterlaceAnalysis struct {
	InterlaceDetected bool    `json:"interlace_detected"`
	ProgressiveFrames int     `json:"progressive_frames"`
	InterlacedFrames  int     `json:"interlaced_frames"`
	Confidence        float64 `json:"confidence"`
}

// NoiseAnalysis measures video noise levels
type NoiseAnalysis struct {
	AverageNoise float64 `json:"average_noise"`
	MaxNoise     float64 `json:"max_noise"`
	NoiseProfile string  `json:"noise_profile"`
}

// LoudnessAnalysis provides broadcast loudness compliance
type LoudnessAnalysis struct {
	IntegratedLoudness float64 `json:"integrated_loudness_lufs"`
	LoudnessRange      float64 `json:"loudness_range_lu"`
	TruePeak           float64 `json:"true_peak_dbtp"`
	Compliant          bool    `json:"broadcast_compliant"`
	Standard           string  `json:"standard"`
}

// HDRAnalysis provides comprehensive HDR metadata analysis
type HDRAnalysis struct {
	IsHDR             bool                      `json:"is_hdr"`
	HDRFormat         string                    `json:"hdr_format,omitempty"`      // HDR10, HDR10+, Dolby Vision, HLG
	ColorPrimaries    string                    `json:"color_primaries,omitempty"` // bt2020, etc.
	ColorTransfer     string                    `json:"color_transfer,omitempty"`  // smpte2084, arib-std-b67, etc.
	ColorSpace        string                    `json:"color_space,omitempty"`     // bt2020nc, etc.
	MasteringDisplay  *MasteringDisplayMetadata `json:"mastering_display,omitempty"`
	ContentLightLevel *ContentLightLevelData    `json:"content_light_level,omitempty"`
	DolbyVision       *DolbyVisionMetadata      `json:"dolby_vision,omitempty"`
	HDR10Plus         *HDR10PlusMetadata        `json:"hdr10_plus,omitempty"`
	HLGCompatible     bool                      `json:"hlg_compatible"`
	Validation        *HDRValidation            `json:"validation,omitempty"`
}

// MasteringDisplayMetadata contains mastering display color volume information
type MasteringDisplayMetadata struct {
	DisplayPrimariesX   [3]float64 `json:"display_primaries_x"` // Red, Green, Blue X coordinates
	DisplayPrimariesY   [3]float64 `json:"display_primaries_y"` // Red, Green, Blue Y coordinates
	WhitePointX         float64    `json:"white_point_x"`
	WhitePointY         float64    `json:"white_point_y"`
	MaxDisplayLuminance float64    `json:"max_display_luminance"` // nits
	MinDisplayLuminance float64    `json:"min_display_luminance"` // nits
	HasMasteringDisplay bool       `json:"has_mastering_display"`
}

// ContentLightLevelData contains content light level information
type ContentLightLevelData struct {
	MaxCLL               int  `json:"max_cll"`  // Maximum Content Light Level (nits)
	MaxFALL              int  `json:"max_fall"` // Maximum Frame-Average Light Level (nits)
	HasContentLightLevel bool `json:"has_content_light_level"`
}

// DolbyVisionMetadata contains Dolby Vision specific metadata
type DolbyVisionMetadata struct {
	Profile                 int  `json:"profile"`
	Level                   int  `json:"level"`
	RPUPresent              bool `json:"rpu_present"` // Reference Processing Unit
	ELPresent               bool `json:"el_present"`  // Enhancement Layer
	BLPresent               bool `json:"bl_present"`  // Base Layer
	BLSignalCompatibilityID int  `json:"bl_signal_compatibility_id"`
}

// HDR10PlusMetadata contains HDR10+ dynamic metadata information
type HDR10PlusMetadata struct {
	Present                                  bool    `json:"present"`
	ApplicationVersion                       int     `json:"application_version,omitempty"`
	NumWindows                               int     `json:"num_windows,omitempty"`
	TargetedSystemDisplayActualPeakLuminance float64 `json:"targeted_system_display_actual_peak_luminance,omitempty"`
}

// HDRValidation contains HDR compliance validation results
type HDRValidation struct {
	IsCompliant     bool     `json:"is_compliant"`
	Standard        string   `json:"standard,omitempty"` // HDR10, Dolby Vision, HLG
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
	GamutCoverage   float64  `json:"gamut_coverage,omitempty"` // Percentage of Rec.2020 gamut covered
}

// BitDepthAnalysis provides comprehensive bit depth analysis
type BitDepthAnalysis struct {
	VideoStreams     map[int]*VideoBitDepth `json:"video_streams,omitempty"`
	AudioStreams     map[int]*AudioBitDepth `json:"audio_streams,omitempty"`
	MaxVideoBitDepth int                    `json:"max_video_bit_depth"`
	MaxAudioBitDepth int                    `json:"max_audio_bit_depth"`
	IsHDR            bool                   `json:"is_hdr"`            // Based on bit depth characteristics
	IsHighBitDepth   bool                   `json:"is_high_bit_depth"` // >8-bit video or >16-bit audio
	Validation       *BitDepthValidation    `json:"validation,omitempty"`
}

// VideoBitDepth contains video bit depth information
type VideoBitDepth struct {
	BitDepth              int    `json:"bit_depth"`
	Source                string `json:"source"` // pixel_format, bits_per_raw_sample, codec_profile, default
	PixelFormat           string `json:"pixel_format,omitempty"`
	ProfileIndicatedDepth int    `json:"profile_indicated_depth,omitempty"` // Bit depth indicated by codec profile
	IsConsistent          bool   `json:"is_consistent"`                     // Whether all indicators agree
}

// AudioBitDepth contains audio bit depth information
type AudioBitDepth struct {
	BitDepth     int    `json:"bit_depth"`
	Source       string `json:"source"` // sample_format, bits_per_sample, bits_per_raw_sample, default
	SampleFormat string `json:"sample_format,omitempty"`
	IsConsistent bool   `json:"is_consistent"` // Whether all indicators agree
}

// BitDepthValidation contains bit depth validation results
type BitDepthValidation struct {
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// ResolutionAnalysis provides comprehensive resolution and aspect ratio analysis
type ResolutionAnalysis struct {
	VideoStreams           map[int]*VideoResolution `json:"video_streams,omitempty"`
	MaxWidth               int                      `json:"max_width"`
	MaxHeight              int                      `json:"max_height"`
	PrimaryResolution      string                   `json:"primary_resolution,omitempty"`
	IsHighDefinition       bool                     `json:"is_high_definition"`
	IsUltraHighDefinition  bool                     `json:"is_ultra_high_definition"`
	IsWidescreen           bool                     `json:"is_widescreen"`
	HasMultipleResolutions bool                     `json:"has_multiple_resolutions"`
	Validation             *ResolutionValidation    `json:"validation,omitempty"`
}

// VideoResolution contains detailed resolution information for a video stream
type VideoResolution struct {
	Width               int     `json:"width"`
	Height              int     `json:"height"`
	PixelCount          int     `json:"pixel_count"`
	StandardResolution  string  `json:"standard_resolution"` // "Full HD", "4K UHD", etc.
	ResolutionClass     string  `json:"resolution_class"`    // "HD", "4K", "8K", etc.
	SampleAspectRatio   float64 `json:"sample_aspect_ratio"`
	DisplayAspectRatio  float64 `json:"display_aspect_ratio"`
	PixelAspectRatio    float64 `json:"pixel_aspect_ratio"`
	IsAnamorphic        bool    `json:"is_anamorphic"`
	AspectRatioCategory string  `json:"aspect_ratio_category"` // "16:9 (Widescreen)", "4:3 (Standard)", etc.
	Orientation         string  `json:"orientation"`           // "Landscape", "Portrait", "Square"
	IsConsistent        bool    `json:"is_consistent"`         // Whether metadata is consistent
}

// ResolutionValidation contains resolution validation results
type ResolutionValidation struct {
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// FrameRateAnalysis provides comprehensive frame rate analysis
type FrameRateAnalysis struct {
	VideoStreams             map[int]*VideoFrameRate `json:"video_streams,omitempty"`
	MaxFrameRate             float64                 `json:"max_frame_rate"`
	MinFrameRate             float64                 `json:"min_frame_rate"`
	PrimaryFrameRateStandard string                  `json:"primary_frame_rate_standard,omitempty"`
	IsVariableFrameRate      bool                    `json:"is_variable_frame_rate"`
	IsHighFrameRate          bool                    `json:"is_high_frame_rate"` // >= 60 fps
	HasMultipleFrameRates    bool                    `json:"has_multiple_frame_rates"`
	IsInterlaced             bool                    `json:"is_interlaced"`
	Validation               *FrameRateValidation    `json:"validation,omitempty"`
}

// VideoFrameRate contains detailed frame rate information for a video stream
type VideoFrameRate struct {
	RealFrameRate       float64 `json:"real_frame_rate"`      // r_frame_rate
	AverageFrameRate    float64 `json:"average_frame_rate"`   // avg_frame_rate
	EffectiveFrameRate  float64 `json:"effective_frame_rate"` // The frame rate to use
	Source              string  `json:"source"`               // Which field was used for effective rate
	Standard            string  `json:"standard"`             // "24p", "29.97p", "60p", etc.
	Category            string  `json:"category"`             // "Cinema", "Standard", "High Frame Rate", etc.
	IsVariableFrameRate bool    `json:"is_variable_frame_rate"`
	IsInterlaced        bool    `json:"is_interlaced"`
	FrameDuration       float64 `json:"frame_duration_ms"` // Duration of one frame in milliseconds
	IsConsistent        bool    `json:"is_consistent"`     // Whether metadata is consistent
}

// FrameRateValidation contains frame rate validation results
type FrameRateValidation struct {
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// CodecAnalysis provides comprehensive codec analysis
type CodecAnalysis struct {
	VideoCodecs          map[int]*VideoCodecInfo `json:"video_codecs,omitempty"`
	AudioCodecs          map[int]*AudioCodecInfo `json:"audio_codecs,omitempty"`
	SupportedVideoCodecs []string                `json:"supported_video_codecs,omitempty"`
	SupportedAudioCodecs []string                `json:"supported_audio_codecs,omitempty"`
	CodecSummary         map[string]int          `json:"codec_summary,omitempty"` // codec_name -> count
	HasModernCodecs      bool                    `json:"has_modern_codecs"`
	HasLegacyCodecs      bool                    `json:"has_legacy_codecs"`
	IsStreamingOptimized bool                    `json:"is_streaming_optimized"`
	Validation           *CodecValidation        `json:"validation,omitempty"`
}

// VideoCodecInfo contains detailed video codec information
type VideoCodecInfo struct {
	CodecName       string       `json:"codec_name"`
	CodecLongName   string       `json:"codec_long_name"`
	CodecFamily     string       `json:"codec_family"` // "H.264/AVC", "H.265/HEVC", etc.
	Profile         string       `json:"profile"`
	Level           int          `json:"level"`
	ProfileInfo     *ProfileInfo `json:"profile_info,omitempty"`
	LevelInfo       *LevelInfo   `json:"level_info,omitempty"`
	Generation      string       `json:"generation"`         // "4th Generation", "5th Generation"
	Features        []string     `json:"features,omitempty"` // Codec-specific features
	HardwareSupport []string     `json:"hardware_support,omitempty"`
	IsValid         bool         `json:"is_valid"` // Whether profile/level combination is valid
}

// AudioCodecInfo contains detailed audio codec information
type AudioCodecInfo struct {
	CodecName       string       `json:"codec_name"`
	CodecLongName   string       `json:"codec_long_name"`
	CodecFamily     string       `json:"codec_family"` // "AAC", "MP3", etc.
	Profile         string       `json:"profile"`
	ProfileInfo     *ProfileInfo `json:"profile_info,omitempty"`
	SampleFormat    string       `json:"sample_format"`
	SampleRate      string       `json:"sample_rate"`
	Channels        int          `json:"channels"`
	ChannelLayout   string       `json:"channel_layout"`
	IsLossless      bool         `json:"is_lossless"`
	IsSurround      bool         `json:"is_surround"`
	HardwareSupport []string     `json:"hardware_support,omitempty"`
	IsValid         bool         `json:"is_valid"`
}

// ProfileInfo contains codec profile information
type ProfileInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// LevelInfo contains codec level information
type LevelInfo struct {
	Level         int    `json:"level"`
	Description   string `json:"description,omitempty"`
	MaxResolution string `json:"max_resolution,omitempty"`
	MaxFrameRate  string `json:"max_frame_rate,omitempty"`
}

// CodecValidation contains codec validation results
type CodecValidation struct {
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// ContainerAnalysis provides comprehensive container format analysis
type ContainerAnalysis struct {
	FormatName          string               `json:"format_name"`
	FormatLongName      string               `json:"format_long_name"`
	FileName            string               `json:"file_name"`
	ContainerFamily     string               `json:"container_family"` // "MP4", "Matroska", etc.
	ContainerInfo       *ContainerInfo       `json:"container_info,omitempty"`
	StreamCount         int                  `json:"stream_count"`
	ProgramCount        int                  `json:"program_count"`
	Duration            float64              `json:"duration_seconds"`
	FileSize            int64                `json:"file_size_bytes"`
	OverallBitRate      int64                `json:"overall_bit_rate"`
	ProbeScore          int                  `json:"probe_score"`
	IsStreamingFriendly bool                 `json:"is_streaming_friendly"`
	SupportedCodecs     *SupportedCodecs     `json:"supported_codecs,omitempty"`
	Features            []string             `json:"features,omitempty"`
	UseCases            []string             `json:"use_cases,omitempty"`
	Tags                map[string]string    `json:"tags,omitempty"`
	Validation          *ContainerValidation `json:"validation,omitempty"`
}

// ContainerInfo contains detailed container format information
type ContainerInfo struct {
	Description    string   `json:"description"`
	MimeType       string   `json:"mime_type,omitempty"`
	Extensions     []string `json:"extensions,omitempty"`
	StandardizedBy string   `json:"standardized_by,omitempty"`
	YearIntroduced int      `json:"year_introduced,omitempty"`
	IsOpenStandard bool     `json:"is_open_standard"`
}

// SupportedCodecs contains codec support information for container
type SupportedCodecs struct {
	Video    []string `json:"video,omitempty"`
	Audio    []string `json:"audio,omitempty"`
	Subtitle []string `json:"subtitle,omitempty"`
}

// ContainerValidation contains container validation results
type ContainerValidation struct {
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// StreamDispositionAnalysis provides comprehensive stream disposition analysis
type StreamDispositionAnalysis struct {
	VideoStreams         map[int]*StreamDisposition `json:"video_streams,omitempty"`
	AudioStreams         map[int]*StreamDisposition `json:"audio_streams,omitempty"`
	SubtitleStreams      map[int]*StreamDisposition `json:"subtitle_streams,omitempty"`
	HasMainStreams       bool                       `json:"has_main_streams"`
	HasAlternateStreams  bool                       `json:"has_alternate_streams"`
	HasForcedSubtitles   bool                       `json:"has_forced_subtitles"`
	HasSDHSubtitles      bool                       `json:"has_sdh_subtitles"` // Subtitles for Deaf/Hard-of-hearing
	HasCommentary        bool                       `json:"has_commentary"`
	HasDescriptiveAudio  bool                       `json:"has_descriptive_audio"` // Audio for visually impaired
	LanguageDistribution map[string]int             `json:"language_distribution,omitempty"`
	AccessibilityScore   int                        `json:"accessibility_score"` // 0-100 based on accessibility features
	Validation           *DispositionValidation     `json:"validation,omitempty"`
}

// StreamDisposition contains detailed disposition information for a stream
type StreamDisposition struct {
	StreamIndex       int             `json:"stream_index"`
	Default           bool            `json:"default"`
	Dub               bool            `json:"dub"`
	Original          bool            `json:"original"`
	Comment           bool            `json:"comment"`
	Lyrics            bool            `json:"lyrics"`
	Karaoke           bool            `json:"karaoke"`
	Forced            bool            `json:"forced"`
	HearingImpaired   bool            `json:"hearing_impaired"`
	VisualImpaired    bool            `json:"visual_impaired"`
	CleanEffects      bool            `json:"clean_effects"`
	AttachedPic       bool            `json:"attached_pic"`
	TimedThumbnails   bool            `json:"timed_thumbnails"`
	CaptionService    bool            `json:"caption_service"`
	DispositionFlags  map[string]bool `json:"disposition_flags,omitempty"`
	Language          string          `json:"language,omitempty"`
	Title             string          `json:"title,omitempty"`
	Role              string          `json:"role,omitempty"`               // main, alternate, commentary, etc.
	AccessibilityType string          `json:"accessibility_type,omitempty"` // none, sdh, audio_description, etc.
	IsCompliant       bool            `json:"is_compliant"`                 // Follows accessibility standards
}

// DispositionValidation contains stream disposition validation results
type DispositionValidation struct {
	IsValid                bool     `json:"is_valid"`
	AccessibilityCompliant bool     `json:"accessibility_compliant"`
	Issues                 []string `json:"issues,omitempty"`
	Recommendations        []string `json:"recommendations,omitempty"`
	Standards              []string `json:"standards,omitempty"` // ADA, Section 508, WCAG, etc.
}

// DataIntegrityAnalysis provides comprehensive data integrity validation
type DataIntegrityAnalysis struct {
	FormatErrors         int                      `json:"format_errors"`
	BitstreamErrors      int                      `json:"bitstream_errors"`
	PacketErrors         int                      `json:"packet_errors"`
	ContinuityErrors     int                      `json:"continuity_errors"`
	DataHashes           map[string]string        `json:"data_hashes,omitempty"` // algorithm -> hash
	ErrorSummary         *ErrorSummary            `json:"error_summary,omitempty"`
	IntegrityScore       int                      `json:"integrity_score"` // 0-100, 100 = perfect integrity
	IsCorrupted          bool                     `json:"is_corrupted"`
	IsBroadcastCompliant bool                     `json:"is_broadcast_compliant"`
	Validation           *DataIntegrityValidation `json:"validation,omitempty"`
}

// ErrorSummary contains categorized error information
type ErrorSummary struct {
	CriticalErrors []ErrorDetail  `json:"critical_errors,omitempty"`
	MajorErrors    []ErrorDetail  `json:"major_errors,omitempty"`
	MinorErrors    []ErrorDetail  `json:"minor_errors,omitempty"`
	Warnings       []ErrorDetail  `json:"warnings,omitempty"`
	TotalErrors    int            `json:"total_errors"`
	ErrorsByType   map[string]int `json:"errors_by_type,omitempty"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Type       string `json:"type"`
	Code       int    `json:"code,omitempty"`
	Message    string `json:"message"`
	Location   string `json:"location,omitempty"` // timestamp, frame number, etc.
	Severity   string `json:"severity"`           // critical, major, minor, warning
	Suggestion string `json:"suggestion,omitempty"`
}

// DataIntegrityValidation contains data integrity validation results
type DataIntegrityValidation struct {
	IsValid            bool     `json:"is_valid"`
	BroadcastCompliant bool     `json:"broadcast_compliant"`
	StreamingCompliant bool     `json:"streaming_compliant"`
	Issues             []string `json:"issues,omitempty"`
	Recommendations    []string `json:"recommendations,omitempty"`
	RequiredActions    []string `json:"required_actions,omitempty"`
}
