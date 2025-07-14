package ffmpeg

import (
	"encoding/json"
	"time"
)

// FFprobeOptions contains all possible ffprobe command options
type FFprobeOptions struct {
	// Input options
	Input          string            `json:"input"`
	Format         string            `json:"format,omitempty"`         // -f
	InputOptions   map[string]string `json:"input_options,omitempty"`  // Additional input options

	// Output format options
	OutputFormat   OutputFormat      `json:"output_format,omitempty"`  // -of
	PrettyPrint    bool              `json:"pretty_print,omitempty"`   // -pretty
	ShowFormat     bool              `json:"show_format,omitempty"`    // -show_format
	ShowStreams    bool              `json:"show_streams,omitempty"`   // -show_streams
	ShowPackets    bool              `json:"show_packets,omitempty"`   // -show_packets
	ShowFrames     bool              `json:"show_frames,omitempty"`    // -show_frames
	ShowChapters   bool              `json:"show_chapters,omitempty"`  // -show_chapters
	ShowPrograms   bool              `json:"show_programs,omitempty"`  // -show_programs
	ShowError      bool              `json:"show_error,omitempty"`     // -show_error
	ShowData       bool              `json:"show_data,omitempty"`      // -show_data
	ShowPrivateData bool             `json:"show_private_data,omitempty"` // -show_private_data

	// Selection options
	SelectStreams  string            `json:"select_streams,omitempty"` // -select_streams
	ReadIntervals  string            `json:"read_intervals,omitempty"` // -read_intervals
	ShowEntries    string            `json:"show_entries,omitempty"`   // -show_entries

	// Processing options
	CountFrames    bool              `json:"count_frames,omitempty"`   // -count_frames
	CountPackets   bool              `json:"count_packets,omitempty"`  // -count_packets
	ProbeSize      int64             `json:"probe_size,omitempty"`     // -probesize
	AnalyzeDuration int64            `json:"analyze_duration,omitempty"` // -analyzeduration

	// Logging options
	LogLevel       LogLevel          `json:"log_level,omitempty"`      // -loglevel
	HideBanner     bool              `json:"hide_banner,omitempty"`    // -hide_banner
	Report         bool              `json:"report,omitempty"`         // -report

	// Processing limits
	Timeout        time.Duration     `json:"timeout,omitempty"`        // Custom timeout
	MaxOutputSize  int64             `json:"max_output_size,omitempty"` // Custom limit
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
	Output     string            `json:"output"`
	Format     *FormatInfo       `json:"format,omitempty"`
	Streams    []StreamInfo      `json:"streams,omitempty"`
	Packets    []PacketInfo      `json:"packets,omitempty"`
	Frames     []FrameInfo       `json:"frames,omitempty"`
	Chapters   []ChapterInfo     `json:"chapters,omitempty"`
	Programs   []ProgramInfo     `json:"programs,omitempty"`
	Error      *ErrorInfo        `json:"error,omitempty"`

	// Execution metadata
	Command        []string      `json:"command"`
	ExecutionTime  time.Duration `json:"execution_time"`
	Success        bool          `json:"success"`
	ExitCode       int           `json:"exit_code"`
	StdErr         string        `json:"stderr,omitempty"`
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
	MediaType       string            `json:"media_type"`
	StreamIndex     int               `json:"stream_index"`
	KeyFrame        int               `json:"key_frame"`
	Pts             int64             `json:"pts"`
	PtsTime         string            `json:"pts_time"`
	PktPts          int64             `json:"pkt_pts"`
	PktPtsTime      string            `json:"pkt_pts_time"`
	PktDts          int64             `json:"pkt_dts"`
	PktDtsTime      string            `json:"pkt_dts_time"`
	BestEffortTimestamp     int64     `json:"best_effort_timestamp"`
	BestEffortTimestampTime string    `json:"best_effort_timestamp_time"`
	PktDuration     int64             `json:"pkt_duration"`
	PktDurationTime string            `json:"pkt_duration_time"`
	PktPos          string            `json:"pkt_pos"`
	PktSize         string            `json:"pkt_size"`
	Width           int               `json:"width,omitempty"`
	Height          int               `json:"height,omitempty"`
	PixFmt          string            `json:"pix_fmt,omitempty"`
	SampleAspectRatio string          `json:"sample_aspect_ratio,omitempty"`
	PictType        string            `json:"pict_type,omitempty"`
	CodedPictureNumber int            `json:"coded_picture_number,omitempty"`
	DisplayPictureNumber int          `json:"display_picture_number,omitempty"`
	InterlacedFrame int               `json:"interlaced_frame,omitempty"`
	TopFieldFirst   int               `json:"top_field_first,omitempty"`
	RepeatPict      int               `json:"repeat_pict,omitempty"`
	SampleFmt       string            `json:"sample_fmt,omitempty"`
	NBSamples       int               `json:"nb_samples,omitempty"`
	Channels        int               `json:"channels,omitempty"`
	ChannelLayout   string            `json:"channel_layout,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
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
	ProgramID   int               `json:"program_id"`
	ProgramNum  int               `json:"program_num"`
	NBStreams   int               `json:"nb_streams"`
	PmtPID      int               `json:"pmt_pid"`
	PcrPID      int               `json:"pcr_pid"`
	StartPts    int64             `json:"start_pts"`
	StartTime   string            `json:"start_time"`
	EndPts      int64             `json:"end_pts"`
	EndTime     string            `json:"end_time"`
	Tags        map[string]string `json:"tags,omitempty"`
	Streams     []int             `json:"streams,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code   int    `json:"code"`
	String string `json:"string"`
}

// StreamSpecifier represents stream selection options
type StreamSpecifier struct {
	Type     string `json:"type,omitempty"`     // v, a, s, d (video, audio, subtitle, data)
	Index    int    `json:"index,omitempty"`    // Stream index
	ProgramID int   `json:"program_id,omitempty"` // Program ID
	Metadata string `json:"metadata,omitempty"` // Metadata tag
	Usable   bool   `json:"usable,omitempty"`   // Only usable streams
}

// ReadInterval represents time interval specification  
type ReadInterval struct {
	Start    string `json:"start,omitempty"`    // Start time/percentage
	End      string `json:"end,omitempty"`      // End time/percentage  
	Duration string `json:"duration,omitempty"` // Duration
}