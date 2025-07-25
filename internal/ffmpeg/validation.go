package ffmpeg

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// ValidateOptions validates FFprobe options for security and correctness
func ValidateOptions(opts *FFprobeOptions) error {
	if opts == nil {
		return fmt.Errorf("options cannot be nil")
	}

	// Validate input
	if err := validateInput(opts.Input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	// Validate timeout
	if opts.Timeout > 0 && opts.Timeout > 60*time.Minute {
		return fmt.Errorf("timeout cannot exceed 60 minutes")
	}

	// Validate output size
	if opts.MaxOutputSize > 0 && opts.MaxOutputSize > 1024*1024*1024 {
		return fmt.Errorf("max output size cannot exceed 1GB")
	}

	// Validate probe size
	if opts.ProbeSize > 0 && opts.ProbeSize > 1024*1024*1024 {
		return fmt.Errorf("probe size cannot exceed 1GB")
	}

	// Validate analyze duration
	if opts.AnalyzeDuration > 0 && opts.AnalyzeDuration > 3600*1000000 { // 1 hour in microseconds
		return fmt.Errorf("analyze duration cannot exceed 1 hour")
	}

	// Validate show entries format
	if opts.ShowEntries != "" {
		if err := validateShowEntries(opts.ShowEntries); err != nil {
			return fmt.Errorf("invalid show_entries: %w", err)
		}
	}

	// Validate select streams format
	if opts.SelectStreams != "" {
		if err := validateSelectStreams(opts.SelectStreams); err != nil {
			return fmt.Errorf("invalid select_streams: %w", err)
		}
	}

	// Validate read intervals format
	if opts.ReadIntervals != "" {
		if err := validateReadIntervals(opts.ReadIntervals); err != nil {
			return fmt.Errorf("invalid read_intervals: %w", err)
		}
	}

	return nil
}

// validateInput validates the input path or URL
func validateInput(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("input cannot be empty")
	}

	// Check for potential command injection
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(input, char) {
			return fmt.Errorf("input contains dangerous character: %s", char)
		}
	}

	// If it's a local file, check it exists and is readable
	if !strings.Contains(input, "://") {
		if info, err := os.Stat(input); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", input)
			}
			return fmt.Errorf("cannot access file: %w", err)
		} else if info.IsDir() {
			return fmt.Errorf("input is a directory, not a file")
		}
	}

	return nil
}

// validateShowEntries validates the show_entries parameter format
func validateShowEntries(entries string) error {
	// Basic format validation for show_entries
	// Format: stream=index,codec_name:format=duration,size
	validPattern := regexp.MustCompile(`^[a-zA-Z_,=:]+$`)
	if !validPattern.MatchString(entries) {
		return fmt.Errorf("invalid characters in show_entries")
	}

	// Check for known valid prefixes
	validPrefixes := []string{"stream", "format", "packet", "frame", "chapter", "program"}
	parts := strings.Split(entries, ":")
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		found := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(part, prefix+"=") || part == prefix {
				found = true
				break
			}
		}
		
		if !found {
			return fmt.Errorf("unknown section in show_entries: %s", part)
		}
	}

	return nil
}

// validateSelectStreams validates the select_streams parameter format
func validateSelectStreams(streams string) error {
	// Valid stream selectors: v:0, a:1, s, v:0,a:1, etc.
	validPattern := regexp.MustCompile(`^[vasdt]?(:[0-9]+)?(,[vasdt]?(:[0-9]+)?)*$`)
	if !validPattern.MatchString(streams) {
		return fmt.Errorf("invalid stream selector format")
	}

	return nil
}

// validateReadIntervals validates the read_intervals parameter format
func validateReadIntervals(intervals string) error {
	// Basic validation for read intervals
	// Format: %+#-start,+#end or timestamps
	if strings.Contains(intervals, ";") || strings.Contains(intervals, "&") {
		return fmt.Errorf("invalid characters in read_intervals")
	}

	return nil
}

// ValidateResult validates FFprobe result for consistency
func ValidateResult(result *FFprobeResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	// Check if successful execution matches output
	if result.Success && result.ExitCode != 0 {
		return fmt.Errorf("inconsistent result: success=true but exit_code=%d", result.ExitCode)
	}

	if !result.Success && result.ExitCode == 0 {
		return fmt.Errorf("inconsistent result: success=false but exit_code=0")
	}

	// Validate parsed data consistency
	if result.Format != nil {
		if err := validateFormatInfo(result.Format); err != nil {
			return fmt.Errorf("invalid format info: %w", err)
		}
	}

	for i, stream := range result.Streams {
		if err := validateStreamInfo(&stream); err != nil {
			return fmt.Errorf("invalid stream %d: %w", i, err)
		}
	}

	return nil
}

// validateFormatInfo validates format information
func validateFormatInfo(format *FormatInfo) error {
	if format.NBStreams < 0 {
		return fmt.Errorf("negative number of streams")
	}

	if format.ProbeScore < 0 || format.ProbeScore > 100 {
		return fmt.Errorf("invalid probe score: %d", format.ProbeScore)
	}

	return nil
}

// validateStreamInfo validates stream information
func validateStreamInfo(stream *StreamInfo) error {
	if stream.Index < 0 {
		return fmt.Errorf("negative stream index")
	}

	// Validate video stream specifics
	if stream.CodecType == "video" {
		if stream.Width < 0 || stream.Height < 0 {
			return fmt.Errorf("invalid video dimensions: %dx%d", stream.Width, stream.Height)
		}
	}

	// Validate audio stream specifics
	if stream.CodecType == "audio" {
		if stream.Channels < 0 {
			return fmt.Errorf("invalid number of audio channels: %d", stream.Channels)
		}
	}

	return nil
}