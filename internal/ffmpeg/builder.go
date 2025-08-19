package ffmpeg

import (
	"fmt"
	"strings"
	"time"
)

// OptionsBuilder provides a fluent interface for building FFprobeOptions
type OptionsBuilder struct {
	options *FFprobeOptions
}

// NewOptionsBuilder creates a new options builder
func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{
		options: &FFprobeOptions{
			HideBanner: true, // Hide banner by default
		},
	}
}

// Input sets the input file or URL
func (b *OptionsBuilder) Input(input string) *OptionsBuilder {
	b.options.Input = input
	return b
}

// Format sets the input format
func (b *OptionsBuilder) Format(format string) *OptionsBuilder {
	b.options.Format = format
	return b
}

// OutputFormat sets the output format
func (b *OptionsBuilder) OutputFormat(format OutputFormat) *OptionsBuilder {
	b.options.OutputFormat = format
	return b
}

// JSON sets output format to JSON with pretty printing
func (b *OptionsBuilder) JSON() *OptionsBuilder {
	b.options.OutputFormat = OutputJSON
	b.options.PrettyPrint = true
	return b
}

// XML sets output format to XML
func (b *OptionsBuilder) XML() *OptionsBuilder {
	b.options.OutputFormat = OutputXML
	return b
}

// CSV sets output format to CSV
func (b *OptionsBuilder) CSV() *OptionsBuilder {
	b.options.OutputFormat = OutputCSV
	return b
}

// ShowFormat enables format information output
func (b *OptionsBuilder) ShowFormat() *OptionsBuilder {
	b.options.ShowFormat = true
	return b
}

// ShowStreams enables streams information output
func (b *OptionsBuilder) ShowStreams() *OptionsBuilder {
	b.options.ShowStreams = true
	return b
}

// ShowPackets enables packets information output
func (b *OptionsBuilder) ShowPackets() *OptionsBuilder {
	b.options.ShowPackets = true
	return b
}

// ShowFrames enables frames information output
func (b *OptionsBuilder) ShowFrames() *OptionsBuilder {
	b.options.ShowFrames = true
	return b
}

// ShowChapters enables chapters information output
func (b *OptionsBuilder) ShowChapters() *OptionsBuilder {
	b.options.ShowChapters = true
	return b
}

// ShowPrograms enables programs information output
func (b *OptionsBuilder) ShowPrograms() *OptionsBuilder {
	b.options.ShowPrograms = true
	return b
}

// ShowError enables error information output
func (b *OptionsBuilder) ShowError() *OptionsBuilder {
	b.options.ShowError = true
	return b
}

// ShowData enables data sections output
func (b *OptionsBuilder) ShowData() *OptionsBuilder {
	b.options.ShowData = true
	return b
}

// ShowPrivateData enables private data sections output
func (b *OptionsBuilder) ShowPrivateData() *OptionsBuilder {
	b.options.ShowPrivateData = true
	return b
}

// ShowDataHash enables data hash output
func (b *OptionsBuilder) ShowDataHash() *OptionsBuilder {
	b.options.ShowDataHash = true
	return b
}

// HashAlgorithm sets the hash algorithm for data hash calculation
func (b *OptionsBuilder) HashAlgorithm(algorithm string) *OptionsBuilder {
	b.options.HashAlgorithm = algorithm
	b.options.ShowDataHash = true // Auto-enable when hash is specified
	return b
}

// MD5Hash enables MD5 hash calculation
func (b *OptionsBuilder) MD5Hash() *OptionsBuilder {
	return b.HashAlgorithm("md5")
}

// CRC32Hash enables CRC32 hash calculation
func (b *OptionsBuilder) CRC32Hash() *OptionsBuilder {
	return b.HashAlgorithm("crc32")
}

// ErrorDetect sets error detection flags
func (b *OptionsBuilder) ErrorDetect(flags string) *OptionsBuilder {
	b.options.ErrorDetect = flags
	return b
}

// ErrorDetectAll enables all error detection
func (b *OptionsBuilder) ErrorDetectAll() *OptionsBuilder {
	b.options.ErrorDetect = "crccheck+bitstream+buffer+explode+careful+compliant+aggressive"
	return b
}

// ErrorDetectBroadcast enables broadcast-safe error detection
func (b *OptionsBuilder) ErrorDetectBroadcast() *OptionsBuilder {
	b.options.ErrorDetect = "crccheck+bitstream+buffer+careful+compliant"
	return b
}

// FormatErrorDetect sets format error detection flags
func (b *OptionsBuilder) FormatErrorDetect(flags string) *OptionsBuilder {
	b.options.FormatErrorDetect = flags
	return b
}

// FormatErrorDetectAll enables all format error detection
func (b *OptionsBuilder) FormatErrorDetectAll() *OptionsBuilder {
	b.options.FormatErrorDetect = "crccheck+bitstream+buffer+explode+careful+compliant+aggressive"
	return b
}

// ShowAll enables all information output
func (b *OptionsBuilder) ShowAll() *OptionsBuilder {
	return b.ShowFormat().ShowStreams().ShowChapters().ShowPrograms()
}

// SelectStreams sets stream selection criteria
func (b *OptionsBuilder) SelectStreams(specifier string) *OptionsBuilder {
	b.options.SelectStreams = specifier
	return b
}

// SelectVideoStreams selects only video streams
func (b *OptionsBuilder) SelectVideoStreams() *OptionsBuilder {
	b.options.SelectStreams = "v"
	return b
}

// SelectAudioStreams selects only audio streams
func (b *OptionsBuilder) SelectAudioStreams() *OptionsBuilder {
	b.options.SelectStreams = "a"
	return b
}

// SelectSubtitleStreams selects only subtitle streams
func (b *OptionsBuilder) SelectSubtitleStreams() *OptionsBuilder {
	b.options.SelectStreams = "s"
	return b
}

// SelectStreamByIndex selects stream by index
func (b *OptionsBuilder) SelectStreamByIndex(index int) *OptionsBuilder {
	b.options.SelectStreams = fmt.Sprintf("%d", index)
	return b
}

// ReadIntervals sets read intervals
func (b *OptionsBuilder) ReadIntervals(intervals string) *OptionsBuilder {
	b.options.ReadIntervals = intervals
	return b
}

// ReadInterval sets a single read interval
func (b *OptionsBuilder) ReadInterval(start, duration string) *OptionsBuilder {
	if duration != "" {
		b.options.ReadIntervals = fmt.Sprintf("%s+%s", start, duration)
	} else {
		b.options.ReadIntervals = start
	}
	return b
}

// ReadPercentage reads a percentage of the file
func (b *OptionsBuilder) ReadPercentage(startPercent, durationPercent int) *OptionsBuilder {
	if durationPercent > 0 {
		b.options.ReadIntervals = fmt.Sprintf("%d%%+%d%%", startPercent, durationPercent)
	} else {
		b.options.ReadIntervals = fmt.Sprintf("%d%%", startPercent)
	}
	return b
}

// ShowEntries sets specific entries to show
func (b *OptionsBuilder) ShowEntries(entries string) *OptionsBuilder {
	b.options.ShowEntries = entries
	return b
}

// ShowStreamEntries shows specific stream entries
func (b *OptionsBuilder) ShowStreamEntries(entries ...string) *OptionsBuilder {
	b.options.ShowEntries = "stream=" + strings.Join(entries, ",")
	return b
}

// ShowFormatEntries shows specific format entries
func (b *OptionsBuilder) ShowFormatEntries(entries ...string) *OptionsBuilder {
	b.options.ShowEntries = "format=" + strings.Join(entries, ",")
	return b
}

// CountFrames enables frame counting
func (b *OptionsBuilder) CountFrames() *OptionsBuilder {
	b.options.CountFrames = true
	return b
}

// CountPackets enables packet counting
func (b *OptionsBuilder) CountPackets() *OptionsBuilder {
	b.options.CountPackets = true
	return b
}

// ProbeSize sets the probe size in bytes
func (b *OptionsBuilder) ProbeSize(size int64) *OptionsBuilder {
	b.options.ProbeSize = size
	return b
}

// ProbeSizeMB sets the probe size in megabytes
func (b *OptionsBuilder) ProbeSizeMB(sizeMB int) *OptionsBuilder {
	b.options.ProbeSize = int64(sizeMB) * 1024 * 1024
	return b
}

// AnalyzeDuration sets the analyze duration in microseconds
func (b *OptionsBuilder) AnalyzeDuration(duration int64) *OptionsBuilder {
	b.options.AnalyzeDuration = duration
	return b
}

// AnalyzeDurationSeconds sets the analyze duration in seconds
func (b *OptionsBuilder) AnalyzeDurationSeconds(seconds int) *OptionsBuilder {
	b.options.AnalyzeDuration = int64(seconds) * 1000000 // Convert to microseconds
	return b
}

// LogLevel sets the log level
func (b *OptionsBuilder) LogLevel(level LogLevel) *OptionsBuilder {
	b.options.LogLevel = level
	return b
}

// Quiet sets log level to quiet
func (b *OptionsBuilder) Quiet() *OptionsBuilder {
	b.options.LogLevel = LogQuiet
	return b
}

// Verbose sets log level to verbose
func (b *OptionsBuilder) Verbose() *OptionsBuilder {
	b.options.LogLevel = LogVerbose
	return b
}

// Debug sets log level to debug
func (b *OptionsBuilder) Debug() *OptionsBuilder {
	b.options.LogLevel = LogDebug
	return b
}

// HideBanner controls banner visibility
func (b *OptionsBuilder) HideBanner(hide bool) *OptionsBuilder {
	b.options.HideBanner = hide
	return b
}

// ShowBanner shows the ffprobe banner
func (b *OptionsBuilder) ShowBanner() *OptionsBuilder {
	b.options.HideBanner = false
	return b
}

// Report enables detailed reporting
func (b *OptionsBuilder) Report() *OptionsBuilder {
	b.options.Report = true
	return b
}

// Timeout sets a custom timeout
func (b *OptionsBuilder) Timeout(timeout time.Duration) *OptionsBuilder {
	b.options.Timeout = timeout
	return b
}

// TimeoutSeconds sets timeout in seconds
func (b *OptionsBuilder) TimeoutSeconds(seconds int) *OptionsBuilder {
	b.options.Timeout = time.Duration(seconds) * time.Second
	return b
}

// MaxOutputSize sets maximum output size
func (b *OptionsBuilder) MaxOutputSize(size int64) *OptionsBuilder {
	b.options.MaxOutputSize = size
	return b
}

// MaxOutputSizeMB sets maximum output size in megabytes
func (b *OptionsBuilder) MaxOutputSizeMB(sizeMB int) *OptionsBuilder {
	b.options.MaxOutputSize = int64(sizeMB) * 1024 * 1024
	return b
}

// InputOption adds a custom input option
func (b *OptionsBuilder) InputOption(key, value string) *OptionsBuilder {
	if b.options.InputOptions == nil {
		b.options.InputOptions = make(map[string]string)
	}
	b.options.InputOptions[key] = value
	return b
}

// InputOptions adds multiple input options
func (b *OptionsBuilder) InputOptions(options map[string]string) *OptionsBuilder {
	if b.options.InputOptions == nil {
		b.options.InputOptions = make(map[string]string)
	}
	for key, value := range options {
		b.options.InputOptions[key] = value
	}
	return b
}

// Build returns the configured FFprobeOptions
func (b *OptionsBuilder) Build() *FFprobeOptions {
	return b.options
}

// Convenience methods for common use cases

// BasicInfo configures options for basic media information
func (b *OptionsBuilder) BasicInfo() *OptionsBuilder {
	return b.JSON().ShowFormat().ShowStreams()
}

// DetailedInfo configures options for detailed media information
func (b *OptionsBuilder) DetailedInfo() *OptionsBuilder {
	return b.JSON().ShowAll().CountFrames()
}

// VideoInfo configures options for video-specific information
func (b *OptionsBuilder) VideoInfo() *OptionsBuilder {
	return b.JSON().ShowFormat().ShowStreams().SelectVideoStreams()
}

// AudioInfo configures options for audio-specific information
func (b *OptionsBuilder) AudioInfo() *OptionsBuilder {
	return b.JSON().ShowFormat().ShowStreams().SelectAudioStreams()
}

// QuickInfo configures options for fast analysis
func (b *OptionsBuilder) QuickInfo() *OptionsBuilder {
	return b.JSON().ShowFormat().ShowStreams().ProbeSizeMB(1).AnalyzeDurationSeconds(1)
}

// DeepAnalysis configures options for comprehensive analysis
func (b *OptionsBuilder) DeepAnalysis() *OptionsBuilder {
	return b.JSON().ShowAll().CountFrames().CountPackets().ProbeSizeMB(50).AnalyzeDurationSeconds(30)
}

// QualityControlAnalysis configures options for professional QC analysis
func (b *OptionsBuilder) QualityControlAnalysis() *OptionsBuilder {
	return b.JSON().ShowAll().ShowError().ShowData().ShowDataHash().ShowPrivateData().
		CountFrames().CountPackets().ErrorDetectBroadcast().FormatErrorDetectAll().
		CRC32Hash().ProbeSizeMB(100).AnalyzeDurationSeconds(60)
}

// BroadcastQC configures options for broadcast compliance analysis
func (b *OptionsBuilder) BroadcastQC() *OptionsBuilder {
	return b.JSON().ShowAll().ShowError().ShowPrivateData().
		ErrorDetectBroadcast().FormatErrorDetectAll().
		CountFrames().CountPackets().ProbeSizeMB(50).AnalyzeDurationSeconds(30)
}

// StreamingQC configures options for streaming platform compliance
func (b *OptionsBuilder) StreamingQC() *OptionsBuilder {
	return b.JSON().ShowAll().ShowError().ShowDataHash().
		ErrorDetect("crccheck+bitstream+buffer+careful").
		MD5Hash().CountFrames().ProbeSizeMB(25).AnalyzeDurationSeconds(15)
}
