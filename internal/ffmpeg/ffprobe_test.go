package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewFFprobe(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("with empty path uses default", func(t *testing.T) {
		ffprobe := NewFFprobe("", logger)
		if ffprobe == nil {
			t.Fatal("Expected FFprobe instance, got nil")
		}
		if ffprobe.binaryPath != "ffprobe" {
			t.Errorf("Expected default binary path 'ffprobe', got %q", ffprobe.binaryPath)
		}
	})

	t.Run("with custom path", func(t *testing.T) {
		ffprobe := NewFFprobe("/usr/local/bin/ffprobe", logger)
		if ffprobe == nil {
			t.Fatal("Expected FFprobe instance, got nil")
		}
		if ffprobe.binaryPath != "/usr/local/bin/ffprobe" {
			t.Errorf("Expected custom binary path, got %q", ffprobe.binaryPath)
		}
	})

	t.Run("default timeout is 5 minutes", func(t *testing.T) {
		ffprobe := NewFFprobe("", logger)
		expected := 5 * time.Minute
		if ffprobe.defaultTimeout != expected {
			t.Errorf("Expected default timeout %v, got %v", expected, ffprobe.defaultTimeout)
		}
	})

	t.Run("default max output size is 100MB", func(t *testing.T) {
		ffprobe := NewFFprobe("", logger)
		expected := int64(100 * 1024 * 1024)
		if ffprobe.maxOutputSize != expected {
			t.Errorf("Expected default max output size %d, got %d", expected, ffprobe.maxOutputSize)
		}
	})

	t.Run("content analysis disabled by default", func(t *testing.T) {
		ffprobe := NewFFprobe("", logger)
		if ffprobe.enableContentAnalysis {
			t.Error("Expected content analysis to be disabled by default")
		}
	})

	t.Run("enhanced analyzer is initialized", func(t *testing.T) {
		ffprobe := NewFFprobe("", logger)
		if ffprobe.enhancedAnalyzer == nil {
			t.Error("Expected enhanced analyzer to be initialized")
		}
	})
}

func TestSetDefaultTimeout(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	newTimeout := 10 * time.Minute
	ffprobe.SetDefaultTimeout(newTimeout)

	if ffprobe.defaultTimeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, ffprobe.defaultTimeout)
	}
}

func TestSetMaxOutputSize(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	newSize := int64(50 * 1024 * 1024)
	ffprobe.SetMaxOutputSize(newSize)

	if ffprobe.maxOutputSize != newSize {
		t.Errorf("Expected max output size %d, got %d", newSize, ffprobe.maxOutputSize)
	}
}

func TestEnableContentAnalysis(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	ffprobe.EnableContentAnalysis()

	if !ffprobe.enableContentAnalysis {
		t.Error("Expected content analysis to be enabled")
	}
}

func TestDisableContentAnalysis(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	ffprobe.EnableContentAnalysis()
	ffprobe.DisableContentAnalysis()

	if ffprobe.enableContentAnalysis {
		t.Error("Expected content analysis to be disabled")
	}
}

func TestValidateInput(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	t.Run("URL input is valid", func(t *testing.T) {
		err := ffprobe.ValidateInput("https://example.com/video.mp4")
		if err != nil {
			t.Errorf("Expected URL to be valid, got error: %v", err)
		}
	})

	t.Run("RTMP URL is valid", func(t *testing.T) {
		err := ffprobe.ValidateInput("rtmp://stream.example.com/live")
		if err != nil {
			t.Errorf("Expected RTMP URL to be valid, got error: %v", err)
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		err := ffprobe.ValidateInput("/nonexistent/path/video.mp4")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("existing file is valid", func(t *testing.T) {
		// Use the current test file as a valid existing file
		err := ffprobe.ValidateInput("ffprobe_test.go")
		if err != nil {
			t.Errorf("Expected existing file to be valid, got error: %v", err)
		}
	})
}

func TestBuildArgs(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	t.Run("empty input returns error", func(t *testing.T) {
		options := &FFprobeOptions{}
		_, err := ffprobe.buildArgs(options)
		if err == nil {
			t.Error("Expected error for empty input")
		}
	})

	t.Run("basic options", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:      "test.mp4",
			HideBanner: true,
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that hide_banner is included
		found := false
		for _, arg := range args {
			if arg == "-hide_banner" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected -hide_banner in args")
		}

		// Check that input is last
		if len(args) < 2 || args[len(args)-2] != "-i" || args[len(args)-1] != "test.mp4" {
			t.Error("Expected input to be last with -i flag")
		}
	})

	t.Run("JSON output format", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:        "test.mp4",
			OutputFormat: OutputJSON,
			PrettyPrint:  true,
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasFormat := false
		hasPretty := false
		for _, arg := range args {
			if arg == "-of" {
				hasFormat = true
			}
			if arg == "-pretty" {
				hasPretty = true
			}
		}
		if !hasFormat {
			t.Error("Expected -of flag for output format")
		}
		if !hasPretty {
			t.Error("Expected -pretty flag for pretty print")
		}
	})

	t.Run("show options", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:       "test.mp4",
			ShowFormat:  true,
			ShowStreams: true,
			ShowPackets: true,
			ShowFrames:  true,
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedFlags := []string{"-show_format", "-show_streams", "-show_packets", "-show_frames"}
		for _, expected := range expectedFlags {
			found := false
			for _, arg := range args {
				if arg == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected %s in args", expected)
			}
		}
	})

	t.Run("probe size and analyze duration", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:           "test.mp4",
			ProbeSize:       1024 * 1024,
			AnalyzeDuration: 5000000,
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasProbeSize := false
		hasAnalyzeDuration := false
		for i, arg := range args {
			if arg == "-probesize" && i+1 < len(args) && args[i+1] == "1048576" {
				hasProbeSize = true
			}
			if arg == "-analyzeduration" && i+1 < len(args) && args[i+1] == "5000000" {
				hasAnalyzeDuration = true
			}
		}
		if !hasProbeSize {
			t.Error("Expected -probesize in args")
		}
		if !hasAnalyzeDuration {
			t.Error("Expected -analyzeduration in args")
		}
	})

	t.Run("log level", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:    "test.mp4",
			LogLevel: LogQuiet,
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasLogLevel := false
		for i, arg := range args {
			if arg == "-loglevel" && i+1 < len(args) && args[i+1] == "quiet" {
				hasLogLevel = true
				break
			}
		}
		if !hasLogLevel {
			t.Error("Expected -loglevel quiet in args")
		}
	})

	t.Run("select streams", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:         "test.mp4",
			SelectStreams: "v:0",
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasSelectStreams := false
		for i, arg := range args {
			if arg == "-select_streams" && i+1 < len(args) && args[i+1] == "v:0" {
				hasSelectStreams = true
				break
			}
		}
		if !hasSelectStreams {
			t.Error("Expected -select_streams v:0 in args")
		}
	})

	t.Run("count frames and packets", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:        "test.mp4",
			CountFrames:  true,
			CountPackets: true,
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasCountFrames := false
		hasCountPackets := false
		for _, arg := range args {
			if arg == "-count_frames" {
				hasCountFrames = true
			}
			if arg == "-count_packets" {
				hasCountPackets = true
			}
		}
		if !hasCountFrames {
			t.Error("Expected -count_frames in args")
		}
		if !hasCountPackets {
			t.Error("Expected -count_packets in args")
		}
	})

	t.Run("error detection", func(t *testing.T) {
		options := &FFprobeOptions{
			Input:       "test.mp4",
			ErrorDetect: "crccheck+bitstream+buffer",
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasErrorDetect := false
		for i, arg := range args {
			if arg == "-err_detect" && i+1 < len(args) {
				hasErrorDetect = true
				break
			}
		}
		if !hasErrorDetect {
			t.Error("Expected -err_detect in args")
		}
	})

	t.Run("input options", func(t *testing.T) {
		options := &FFprobeOptions{
			Input: "test.mp4",
			InputOptions: map[string]string{
				"thread_queue_size": "512",
			},
		}
		args, err := ffprobe.buildArgs(options)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		hasInputOption := false
		for i, arg := range args {
			if arg == "-thread_queue_size" && i+1 < len(args) && args[i+1] == "512" {
				hasInputOption = true
				break
			}
		}
		if !hasInputOption {
			t.Error("Expected -thread_queue_size 512 in args")
		}
	})
}

func TestParseJSONOutput(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	t.Run("empty output returns error", func(t *testing.T) {
		result := &FFprobeResult{Output: ""}
		err := ffprobe.parseJSONOutput(result)
		// Empty string is not valid JSON, so it returns an error
		if err == nil {
			t.Error("Expected error for empty output (invalid JSON)")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		result := &FFprobeResult{Output: "not json"}
		err := ffprobe.parseJSONOutput(result)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("valid format info", func(t *testing.T) {
		jsonOutput := `{
			"format": {
				"filename": "test.mp4",
				"nb_streams": 2,
				"format_name": "mov,mp4,m4a,3gp,3g2,mj2",
				"format_long_name": "QuickTime / MOV",
				"duration": "10.000000",
				"size": "1024000",
				"bit_rate": "819200"
			}
		}`
		result := &FFprobeResult{Output: jsonOutput}
		err := ffprobe.parseJSONOutput(result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Format == nil {
			t.Fatal("Expected Format to be parsed")
		}
		if result.Format.Filename != "test.mp4" {
			t.Errorf("Expected filename 'test.mp4', got %q", result.Format.Filename)
		}
		if result.Format.NBStreams != 2 {
			t.Errorf("Expected 2 streams, got %d", result.Format.NBStreams)
		}
	})

	t.Run("valid streams info", func(t *testing.T) {
		jsonOutput := `{
			"streams": [
				{
					"index": 0,
					"codec_name": "h264",
					"codec_type": "video",
					"width": 1920,
					"height": 1080
				},
				{
					"index": 1,
					"codec_name": "aac",
					"codec_type": "audio",
					"sample_rate": "48000",
					"channels": 2
				}
			]
		}`
		result := &FFprobeResult{Output: jsonOutput}
		err := ffprobe.parseJSONOutput(result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(result.Streams) != 2 {
			t.Fatalf("Expected 2 streams, got %d", len(result.Streams))
		}
		if result.Streams[0].CodecName != "h264" {
			t.Errorf("Expected codec_name 'h264', got %q", result.Streams[0].CodecName)
		}
		if result.Streams[0].Width != 1920 {
			t.Errorf("Expected width 1920, got %d", result.Streams[0].Width)
		}
		if result.Streams[1].CodecType != "audio" {
			t.Errorf("Expected codec_type 'audio', got %q", result.Streams[1].CodecType)
		}
	})

	t.Run("valid chapters info", func(t *testing.T) {
		jsonOutput := `{
			"chapters": [
				{
					"id": 1,
					"time_base": "1/1000",
					"start": 0,
					"end": 60000
				}
			]
		}`
		result := &FFprobeResult{Output: jsonOutput}
		err := ffprobe.parseJSONOutput(result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(result.Chapters) != 1 {
			t.Errorf("Expected 1 chapter, got %d", len(result.Chapters))
		}
	})

	t.Run("valid error info", func(t *testing.T) {
		jsonOutput := `{
			"error": {
				"code": -2,
				"string": "No such file or directory"
			}
		}`
		result := &FFprobeResult{Output: jsonOutput}
		err := ffprobe.parseJSONOutput(result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Error == nil {
			t.Fatal("Expected Error to be parsed")
		}
	})
}

func TestExecuteFFprobeCommand(t *testing.T) {
	t.Run("empty command returns error", func(t *testing.T) {
		ctx := context.Background()
		_, err := executeFFprobeCommand(ctx, []string{})
		if err == nil {
			t.Error("Expected error for empty command")
		}
	})

	t.Run("valid echo command", func(t *testing.T) {
		ctx := context.Background()
		output, err := executeFFprobeCommand(ctx, []string{"echo", "test"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if output != "test\n" {
			t.Errorf("Expected 'test\\n', got %q", output)
		}
	})

	t.Run("invalid command returns error", func(t *testing.T) {
		ctx := context.Background()
		_, err := executeFFprobeCommand(ctx, []string{"nonexistent_command_12345"})
		if err == nil {
			t.Error("Expected error for invalid command")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		_, err := executeFFprobeCommand(ctx, []string{"sleep", "10"})
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

// Tests that require ffprobe binary
func TestFFprobeIntegration(t *testing.T) {
	// Skip if ffprobe is not available or not executable
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not available, skipping integration tests")
	}

	// Try to actually execute ffprobe to verify it works
	cmd := exec.Command(ffprobePath, "-version")
	if err := cmd.Run(); err != nil {
		t.Skipf("ffprobe found but not executable (possibly wrong architecture): %v", err)
	}

	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	t.Run("GetVersion", func(t *testing.T) {
		ctx := context.Background()
		version, err := ffprobe.GetVersion(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if version == "" {
			t.Error("Expected non-empty version string")
		}
	})

	t.Run("CheckBinary", func(t *testing.T) {
		ctx := context.Background()
		err := ffprobe.CheckBinary(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("ValidateBinaryAtStartup", func(t *testing.T) {
		ctx := context.Background()
		err := ffprobe.ValidateBinaryAtStartup(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestValidateBinaryAtStartup_InvalidPath(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("/nonexistent/path/ffprobe", logger)

	ctx := context.Background()
	err := ffprobe.ValidateBinaryAtStartup(ctx)
	if err == nil {
		t.Error("Expected error for invalid binary path")
	}
}

func TestProbeWithProgress(t *testing.T) {
	// Skip if ffprobe is not available or not executable
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not available, skipping integration tests")
	}

	// Try to actually execute ffprobe to verify it works
	cmd := exec.Command(ffprobePath, "-version")
	if err := cmd.Run(); err != nil {
		t.Skipf("ffprobe found but not executable (possibly wrong architecture): %v", err)
	}

	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	t.Run("progress callback called", func(t *testing.T) {
		// Create a temp file for testing
		tmpFile, err := os.CreateTemp("", "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		ctx := context.Background()
		options := &FFprobeOptions{
			Input:        tmpFile.Name(),
			OutputFormat: OutputJSON,
			ShowFormat:   true,
		}

		progressCalled := false
		progressCallback := func(progress float64) {
			progressCalled = true
		}

		// This will fail because it's not a valid media file, but the callback should still be called
		_, _ = ffprobe.ProbeWithProgress(ctx, options, progressCallback)

		if !progressCalled {
			t.Error("Expected progress callback to be called")
		}
	})

	t.Run("nil callback doesn't panic", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		ctx := context.Background()
		options := &FFprobeOptions{
			Input:        tmpFile.Name(),
			OutputFormat: OutputJSON,
		}

		// Should not panic with nil callback
		_, _ = ffprobe.ProbeWithProgress(ctx, options, nil)
	})
}

func TestValidateOptions(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		err := ValidateOptions(nil)
		if err == nil {
			t.Error("Expected error for nil options")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		options := &FFprobeOptions{}
		err := ValidateOptions(options)
		if err == nil {
			t.Error("Expected error for empty input")
		}
	})

	t.Run("valid options with URL", func(t *testing.T) {
		// URLs bypass file existence check
		options := &FFprobeOptions{
			Input: "https://example.com/video.mp4",
		}
		err := ValidateOptions(options)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("non-existent local file", func(t *testing.T) {
		options := &FFprobeOptions{
			Input: "/nonexistent/path/test.mp4",
		}
		err := ValidateOptions(options)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("valid options with existing file", func(t *testing.T) {
		// Use the current test file as a valid existing file
		options := &FFprobeOptions{
			Input: "ffprobe_test.go",
		}
		err := ValidateOptions(options)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestValidateResult(t *testing.T) {
	t.Run("nil result", func(t *testing.T) {
		err := ValidateResult(nil)
		if err == nil {
			t.Error("Expected error for nil result")
		}
	})

	t.Run("consistent unsuccessful result", func(t *testing.T) {
		// Success=false with ExitCode=1 is consistent (not an error)
		result := &FFprobeResult{
			Success:  false,
			ExitCode: 1,
		}
		err := ValidateResult(result)
		if err != nil {
			t.Errorf("Unexpected error for consistent unsuccessful result: %v", err)
		}
	})

	t.Run("inconsistent result success true exit code non-zero", func(t *testing.T) {
		// Success=true but ExitCode != 0 is inconsistent
		result := &FFprobeResult{
			Success:  true,
			ExitCode: 1,
		}
		err := ValidateResult(result)
		if err == nil {
			t.Error("Expected error for inconsistent result (success=true, exit_code=1)")
		}
	})

	t.Run("inconsistent result success false exit code zero", func(t *testing.T) {
		// Success=false but ExitCode == 0 is inconsistent
		result := &FFprobeResult{
			Success:  false,
			ExitCode: 0,
		}
		err := ValidateResult(result)
		if err == nil {
			t.Error("Expected error for inconsistent result (success=false, exit_code=0)")
		}
	})

	t.Run("successful result", func(t *testing.T) {
		result := &FFprobeResult{
			Success:  true,
			ExitCode: 0,
		}
		err := ValidateResult(result)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestSetLLMAnalyzer(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	// Should not panic with nil
	ffprobe.SetLLMAnalyzer(nil)

	// Test that it doesn't panic when enhanced analyzer exists
	if ffprobe.enhancedAnalyzer == nil {
		t.Error("Expected enhanced analyzer to be initialized")
	}
}

func TestParseOutput(t *testing.T) {
	logger := zerolog.Nop()
	ffprobe := NewFFprobe("", logger)

	t.Run("non-JSON format keeps raw output", func(t *testing.T) {
		result := &FFprobeResult{Output: "some raw output"}
		options := &FFprobeOptions{OutputFormat: OutputFlat}
		err := ffprobe.parseOutput(result, options)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// Raw output should be preserved
		if result.Output != "some raw output" {
			t.Error("Expected raw output to be preserved")
		}
	})

	t.Run("JSON format parses output", func(t *testing.T) {
		jsonOutput := `{"format": {"filename": "test.mp4"}}`
		result := &FFprobeResult{Output: jsonOutput}
		options := &FFprobeOptions{OutputFormat: OutputJSON}
		err := ffprobe.parseOutput(result, options)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.Format == nil {
			t.Error("Expected format to be parsed")
		}
	})
}
