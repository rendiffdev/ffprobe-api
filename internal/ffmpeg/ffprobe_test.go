package ffmpeg

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFFprobe(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	
	// Test with default binary path
	ffprobe := NewFFprobe("", logger)
	assert.Equal(t, "ffprobe", ffprobe.binaryPath)
	assert.Equal(t, 5*time.Minute, ffprobe.defaultTimeout)
	assert.Equal(t, int64(100*1024*1024), ffprobe.maxOutputSize)
	
	// Test with custom binary path
	ffprobe = NewFFprobe("/usr/bin/ffprobe", logger)
	assert.Equal(t, "/usr/bin/ffprobe", ffprobe.binaryPath)
}

func TestSetters(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ffprobe := NewFFprobe("", logger)
	
	// Test SetDefaultTimeout
	ffprobe.SetDefaultTimeout(10 * time.Minute)
	assert.Equal(t, 10*time.Minute, ffprobe.defaultTimeout)
	
	// Test SetMaxOutputSize
	ffprobe.SetMaxOutputSize(200 * 1024 * 1024)
	assert.Equal(t, int64(200*1024*1024), ffprobe.maxOutputSize)
}

func TestBuildArgs(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ffprobe := NewFFprobe("", logger)
	
	tests := []struct {
		name     string
		options  *FFprobeOptions
		expected []string
		wantErr  bool
	}{
		{
			name: "basic options",
			options: &FFprobeOptions{
				Input:       "test.mp4",
				OutputFormat: OutputJSON,
				ShowFormat:  true,
				ShowStreams: true,
				HideBanner:  true,
			},
			expected: []string{
				"-hide_banner",
				"-of", "json",
				"-show_format",
				"-show_streams",
				"-i", "test.mp4",
			},
			wantErr: false,
		},
		{
			name: "no input",
			options: &FFprobeOptions{
				ShowFormat: true,
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "complex options",
			options: &FFprobeOptions{
				Input:          "test.mp4",
				Format:         "matroska",
				OutputFormat:   OutputXML,
				ShowStreams:    true,
				ShowPackets:    true,
				SelectStreams:  "v:0",
				ReadIntervals:  "10%+20%",
				CountFrames:    true,
				ProbeSize:      1024000,
				AnalyzeDuration: 5000000,
				LogLevel:       LogDebug,
				HideBanner:     true,
			},
			expected: []string{
				"-hide_banner",
				"-loglevel", "debug",
				"-f", "matroska",
				"-probesize", "1024000",
				"-analyzeduration", "5000000",
				"-of", "xml",
				"-show_streams",
				"-show_packets",
				"-select_streams", "v:0",
				"-read_intervals", "10%+20%",
				"-count_frames",
				"-i", "test.mp4",
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := ffprobe.buildArgs(tt.options)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestValidateInput(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ffprobe := NewFFprobe("", logger)
	
	// Test URL (should pass)
	err := ffprobe.ValidateInput("https://example.com/video.mp4")
	assert.NoError(t, err)
	
	// Test non-existent file (should fail)
	err = ffprobe.ValidateInput("/non/existent/file.mp4")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
	
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_video_*.mp4")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// Test existing file (should pass)
	err = ffprobe.ValidateInput(tempFile.Name())
	assert.NoError(t, err)
}

func TestParseJSONOutput(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ffprobe := NewFFprobe("", logger)
	
	// Test valid JSON output
	jsonOutput := `{
		"streams": [
			{
				"index": 0,
				"codec_name": "h264",
				"codec_type": "video",
				"width": 1920,
				"height": 1080
			}
		],
		"format": {
			"filename": "test.mp4",
			"nb_streams": 1,
			"format_name": "mov,mp4,m4a,3gp,3g2,mj2",
			"duration": "120.5"
		}
	}`
	
	result := &FFprobeResult{
		Output: jsonOutput,
	}
	
	err := ffprobe.parseJSONOutput(result)
	require.NoError(t, err)
	
	// Check parsed format
	assert.NotNil(t, result.Format)
	assert.Equal(t, "test.mp4", result.Format.Filename)
	assert.Equal(t, 1, result.Format.NBStreams)
	assert.Equal(t, "mov,mp4,m4a,3gp,3g2,mj2", result.Format.FormatName)
	assert.Equal(t, "120.5", result.Format.Duration)
	
	// Check parsed streams
	assert.Len(t, result.Streams, 1)
	assert.Equal(t, 0, result.Streams[0].Index)
	assert.Equal(t, "h264", result.Streams[0].CodecName)
	assert.Equal(t, "video", result.Streams[0].CodecType)
	assert.Equal(t, 1920, result.Streams[0].Width)
	assert.Equal(t, 1080, result.Streams[0].Height)
}

func TestParseInvalidJSON(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ffprobe := NewFFprobe("", logger)
	
	result := &FFprobeResult{
		Output: "invalid json",
	}
	
	err := ffprobe.parseJSONOutput(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal JSON")
}

func TestOptionsBuilder(t *testing.T) {
	// Test basic builder functionality
	options := NewOptionsBuilder().
		Input("test.mp4").
		JSON().
		ShowFormat().
		ShowStreams().
		Build()
	
	assert.Equal(t, "test.mp4", options.Input)
	assert.Equal(t, OutputJSON, options.OutputFormat)
	assert.True(t, options.ShowFormat)
	assert.True(t, options.ShowStreams)
	assert.True(t, options.PrettyPrint)
	assert.True(t, options.HideBanner)
}

func TestStreamSelection(t *testing.T) {
	// Test video stream selection
	options := NewOptionsBuilder().
		Input("test.mp4").
		SelectVideoStreams().
		Build()
	
	assert.Equal(t, "v", options.SelectStreams)
	
	// Test audio stream selection
	options = NewOptionsBuilder().
		Input("test.mp4").
		SelectAudioStreams().
		Build()
	
	assert.Equal(t, "a", options.SelectStreams)
	
	// Test stream by index
	options = NewOptionsBuilder().
		Input("test.mp4").
		SelectStreamByIndex(2).
		Build()
	
	assert.Equal(t, "2", options.SelectStreams)
}

func TestReadIntervals(t *testing.T) {
	// Test read interval
	options := NewOptionsBuilder().
		Input("test.mp4").
		ReadInterval("10", "30").
		Build()
	
	assert.Equal(t, "10+30", options.ReadIntervals)
	
	// Test read percentage
	options = NewOptionsBuilder().
		Input("test.mp4").
		ReadPercentage(10, 20).
		Build()
	
	assert.Equal(t, "10%+20%", options.ReadIntervals)
}

func TestConvenienceMethods(t *testing.T) {
	// Test BasicInfo
	options := NewOptionsBuilder().
		Input("test.mp4").
		BasicInfo().
		Build()
	
	assert.Equal(t, OutputJSON, options.OutputFormat)
	assert.True(t, options.ShowFormat)
	assert.True(t, options.ShowStreams)
	
	// Test VideoInfo
	options = NewOptionsBuilder().
		Input("test.mp4").
		VideoInfo().
		Build()
	
	assert.Equal(t, OutputJSON, options.OutputFormat)
	assert.True(t, options.ShowFormat)
	assert.True(t, options.ShowStreams)
	assert.Equal(t, "v", options.SelectStreams)
	
	// Test QuickInfo
	options = NewOptionsBuilder().
		Input("test.mp4").
		QuickInfo().
		Build()
	
	assert.Equal(t, int64(1024*1024), options.ProbeSize)
	assert.Equal(t, int64(1000000), options.AnalyzeDuration)
}

func TestTimeouts(t *testing.T) {
	// Test timeout in seconds
	options := NewOptionsBuilder().
		Input("test.mp4").
		TimeoutSeconds(30).
		Build()
	
	assert.Equal(t, 30*time.Second, options.Timeout)
	
	// Test custom timeout
	options = NewOptionsBuilder().
		Input("test.mp4").
		Timeout(5*time.Minute).
		Build()
	
	assert.Equal(t, 5*time.Minute, options.Timeout)
}

func TestInputOptions(t *testing.T) {
	// Test single input option
	options := NewOptionsBuilder().
		Input("test.mp4").
		InputOption("protocol_whitelist", "file,http,https").
		Build()
	
	assert.Equal(t, "file,http,https", options.InputOptions["protocol_whitelist"])
	
	// Test multiple input options
	inputOpts := map[string]string{
		"protocol_whitelist": "file,http,https",
		"timeout":           "5000000",
	}
	
	options = NewOptionsBuilder().
		Input("test.mp4").
		InputOptions(inputOpts).
		Build()
	
	assert.Equal(t, inputOpts, options.InputOptions)
}

// Integration tests (only run if ffprobe is available)
func TestFFprobeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	logger := zerolog.New(os.Stdout).Level(zerolog.WarnLevel)
	ffprobe := NewFFprobe("", logger)
	
	ctx := context.Background()
	
	// Test binary availability
	err := ffprobe.CheckBinary(ctx)
	if err != nil {
		t.Skip("FFprobe binary not available, skipping integration tests")
	}
	
	// Test version retrieval
	version, err := ffprobe.GetVersion(ctx)
	require.NoError(t, err)
	assert.Contains(t, version, "ffprobe version")
	
	t.Logf("FFprobe version: %s", version)
}