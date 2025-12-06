package ffmpeg

import (
	"testing"
	"time"
)

func TestNewOptionsBuilder(t *testing.T) {
	builder := NewOptionsBuilder()

	if builder == nil {
		t.Fatal("NewOptionsBuilder returned nil")
	}

	if builder.options == nil {
		t.Fatal("builder.options is nil")
	}

	// Default should have HideBanner true
	if !builder.options.HideBanner {
		t.Error("expected HideBanner to be true by default")
	}
}

func TestOptionsBuilder_Input(t *testing.T) {
	builder := NewOptionsBuilder().Input("test.mp4")

	opts := builder.Build()
	if opts.Input != "test.mp4" {
		t.Errorf("expected Input to be 'test.mp4', got %q", opts.Input)
	}
}

func TestOptionsBuilder_Format(t *testing.T) {
	builder := NewOptionsBuilder().Format("mp4")

	opts := builder.Build()
	if opts.Format != "mp4" {
		t.Errorf("expected Format to be 'mp4', got %q", opts.Format)
	}
}

func TestOptionsBuilder_OutputFormats(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*OptionsBuilder) *OptionsBuilder
		expected OutputFormat
	}{
		{
			name:     "JSON format",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.JSON() },
			expected: OutputJSON,
		},
		{
			name:     "XML format",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.XML() },
			expected: OutputXML,
		},
		{
			name:     "CSV format",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.CSV() },
			expected: OutputCSV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOptionsBuilder()
			tt.method(builder)
			opts := builder.Build()

			if opts.OutputFormat != tt.expected {
				t.Errorf("expected OutputFormat %v, got %v", tt.expected, opts.OutputFormat)
			}
		})
	}
}

func TestOptionsBuilder_ShowMethods(t *testing.T) {
	builder := NewOptionsBuilder().
		ShowFormat().
		ShowStreams().
		ShowPackets().
		ShowFrames().
		ShowChapters().
		ShowPrograms().
		ShowError().
		ShowData().
		ShowPrivateData().
		ShowDataHash()

	opts := builder.Build()

	if !opts.ShowFormat {
		t.Error("expected ShowFormat to be true")
	}
	if !opts.ShowStreams {
		t.Error("expected ShowStreams to be true")
	}
	if !opts.ShowPackets {
		t.Error("expected ShowPackets to be true")
	}
	if !opts.ShowFrames {
		t.Error("expected ShowFrames to be true")
	}
	if !opts.ShowChapters {
		t.Error("expected ShowChapters to be true")
	}
	if !opts.ShowPrograms {
		t.Error("expected ShowPrograms to be true")
	}
	if !opts.ShowError {
		t.Error("expected ShowError to be true")
	}
	if !opts.ShowData {
		t.Error("expected ShowData to be true")
	}
	if !opts.ShowPrivateData {
		t.Error("expected ShowPrivateData to be true")
	}
	if !opts.ShowDataHash {
		t.Error("expected ShowDataHash to be true")
	}
}

func TestOptionsBuilder_ShowAll(t *testing.T) {
	builder := NewOptionsBuilder().ShowAll()
	opts := builder.Build()

	if !opts.ShowFormat {
		t.Error("ShowAll should enable ShowFormat")
	}
	if !opts.ShowStreams {
		t.Error("ShowAll should enable ShowStreams")
	}
	if !opts.ShowChapters {
		t.Error("ShowAll should enable ShowChapters")
	}
	if !opts.ShowPrograms {
		t.Error("ShowAll should enable ShowPrograms")
	}
}

func TestOptionsBuilder_HashAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*OptionsBuilder) *OptionsBuilder
		expected string
	}{
		{
			name:     "MD5 hash",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.MD5Hash() },
			expected: "md5",
		},
		{
			name:     "CRC32 hash",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.CRC32Hash() },
			expected: "crc32",
		},
		{
			name:     "custom hash",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.HashAlgorithm("sha256") },
			expected: "sha256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOptionsBuilder()
			tt.method(builder)
			opts := builder.Build()

			if opts.HashAlgorithm != tt.expected {
				t.Errorf("expected HashAlgorithm %q, got %q", tt.expected, opts.HashAlgorithm)
			}
			if !opts.ShowDataHash {
				t.Error("HashAlgorithm should auto-enable ShowDataHash")
			}
		})
	}
}

func TestOptionsBuilder_ErrorDetect(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*OptionsBuilder) *OptionsBuilder
		contains string
	}{
		{
			name:     "ErrorDetectAll",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.ErrorDetectAll() },
			contains: "crccheck",
		},
		{
			name:     "ErrorDetectBroadcast",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.ErrorDetectBroadcast() },
			contains: "crccheck",
		},
		{
			name:     "custom ErrorDetect",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.ErrorDetect("crccheck") },
			contains: "crccheck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOptionsBuilder()
			tt.method(builder)
			opts := builder.Build()

			if opts.ErrorDetect == "" {
				t.Error("expected ErrorDetect to be set")
			}
		})
	}
}

func TestOptionsBuilder_StreamSelection(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*OptionsBuilder) *OptionsBuilder
		expected string
	}{
		{
			name:     "video streams",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.SelectVideoStreams() },
			expected: "v",
		},
		{
			name:     "audio streams",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.SelectAudioStreams() },
			expected: "a",
		},
		{
			name:     "subtitle streams",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.SelectSubtitleStreams() },
			expected: "s",
		},
		{
			name:     "stream by index",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.SelectStreamByIndex(0) },
			expected: "0",
		},
		{
			name:     "custom stream specifier",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.SelectStreams("v:0") },
			expected: "v:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOptionsBuilder()
			tt.method(builder)
			opts := builder.Build()

			if opts.SelectStreams != tt.expected {
				t.Errorf("expected SelectStreams %q, got %q", tt.expected, opts.SelectStreams)
			}
		})
	}
}

func TestOptionsBuilder_ProbeSize(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		builder := NewOptionsBuilder().ProbeSize(1024)
		opts := builder.Build()

		if opts.ProbeSize != 1024 {
			t.Errorf("expected ProbeSize 1024, got %d", opts.ProbeSize)
		}
	})

	t.Run("megabytes", func(t *testing.T) {
		builder := NewOptionsBuilder().ProbeSizeMB(5)
		opts := builder.Build()

		expected := int64(5 * 1024 * 1024)
		if opts.ProbeSize != expected {
			t.Errorf("expected ProbeSize %d, got %d", expected, opts.ProbeSize)
		}
	})
}

func TestOptionsBuilder_AnalyzeDuration(t *testing.T) {
	t.Run("microseconds", func(t *testing.T) {
		builder := NewOptionsBuilder().AnalyzeDuration(1000000)
		opts := builder.Build()

		if opts.AnalyzeDuration != 1000000 {
			t.Errorf("expected AnalyzeDuration 1000000, got %d", opts.AnalyzeDuration)
		}
	})

	t.Run("seconds", func(t *testing.T) {
		builder := NewOptionsBuilder().AnalyzeDurationSeconds(5)
		opts := builder.Build()

		expected := int64(5000000)
		if opts.AnalyzeDuration != expected {
			t.Errorf("expected AnalyzeDuration %d, got %d", expected, opts.AnalyzeDuration)
		}
	})
}

func TestOptionsBuilder_LogLevel(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*OptionsBuilder) *OptionsBuilder
		expected LogLevel
	}{
		{
			name:     "quiet",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.Quiet() },
			expected: LogQuiet,
		},
		{
			name:     "verbose",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.Verbose() },
			expected: LogVerbose,
		},
		{
			name:     "debug",
			method:   func(b *OptionsBuilder) *OptionsBuilder { return b.Debug() },
			expected: LogDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOptionsBuilder()
			tt.method(builder)
			opts := builder.Build()

			if opts.LogLevel != tt.expected {
				t.Errorf("expected LogLevel %v, got %v", tt.expected, opts.LogLevel)
			}
		})
	}
}

func TestOptionsBuilder_Timeout(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		timeout := 30 * time.Second
		builder := NewOptionsBuilder().Timeout(timeout)
		opts := builder.Build()

		if opts.Timeout != timeout {
			t.Errorf("expected Timeout %v, got %v", timeout, opts.Timeout)
		}
	})

	t.Run("seconds", func(t *testing.T) {
		builder := NewOptionsBuilder().TimeoutSeconds(60)
		opts := builder.Build()

		expected := 60 * time.Second
		if opts.Timeout != expected {
			t.Errorf("expected Timeout %v, got %v", expected, opts.Timeout)
		}
	})
}

func TestOptionsBuilder_MaxOutputSize(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		builder := NewOptionsBuilder().MaxOutputSize(1024)
		opts := builder.Build()

		if opts.MaxOutputSize != 1024 {
			t.Errorf("expected MaxOutputSize 1024, got %d", opts.MaxOutputSize)
		}
	})

	t.Run("megabytes", func(t *testing.T) {
		builder := NewOptionsBuilder().MaxOutputSizeMB(10)
		opts := builder.Build()

		expected := int64(10 * 1024 * 1024)
		if opts.MaxOutputSize != expected {
			t.Errorf("expected MaxOutputSize %d, got %d", expected, opts.MaxOutputSize)
		}
	})
}

func TestOptionsBuilder_CountMethods(t *testing.T) {
	builder := NewOptionsBuilder().CountFrames().CountPackets()
	opts := builder.Build()

	if !opts.CountFrames {
		t.Error("expected CountFrames to be true")
	}
	if !opts.CountPackets {
		t.Error("expected CountPackets to be true")
	}
}

func TestOptionsBuilder_HideBanner(t *testing.T) {
	t.Run("hide banner", func(t *testing.T) {
		builder := NewOptionsBuilder().HideBanner(true)
		opts := builder.Build()

		if !opts.HideBanner {
			t.Error("expected HideBanner to be true")
		}
	})

	t.Run("show banner", func(t *testing.T) {
		builder := NewOptionsBuilder().ShowBanner()
		opts := builder.Build()

		if opts.HideBanner {
			t.Error("expected HideBanner to be false after ShowBanner")
		}
	})
}

func TestOptionsBuilder_Report(t *testing.T) {
	builder := NewOptionsBuilder().Report()
	opts := builder.Build()

	if !opts.Report {
		t.Error("expected Report to be true")
	}
}

func TestOptionsBuilder_InputOptions(t *testing.T) {
	t.Run("single option", func(t *testing.T) {
		builder := NewOptionsBuilder().InputOption("key", "value")
		opts := builder.Build()

		if opts.InputOptions == nil {
			t.Fatal("InputOptions is nil")
		}
		if opts.InputOptions["key"] != "value" {
			t.Errorf("expected InputOptions[key] = 'value', got %q", opts.InputOptions["key"])
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		options := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		builder := NewOptionsBuilder().InputOptions(options)
		opts := builder.Build()

		if opts.InputOptions == nil {
			t.Fatal("InputOptions is nil")
		}
		if len(opts.InputOptions) != 2 {
			t.Errorf("expected 2 input options, got %d", len(opts.InputOptions))
		}
	})
}

func TestOptionsBuilder_ConvenienceMethods(t *testing.T) {
	tests := []struct {
		name   string
		method func(*OptionsBuilder) *OptionsBuilder
		check  func(*FFprobeOptions) bool
	}{
		{
			name:   "BasicInfo",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.BasicInfo() },
			check: func(o *FFprobeOptions) bool {
				return o.OutputFormat == OutputJSON && o.ShowFormat && o.ShowStreams
			},
		},
		{
			name:   "DetailedInfo",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.DetailedInfo() },
			check: func(o *FFprobeOptions) bool {
				return o.OutputFormat == OutputJSON && o.CountFrames
			},
		},
		{
			name:   "VideoInfo",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.VideoInfo() },
			check: func(o *FFprobeOptions) bool {
				return o.SelectStreams == "v"
			},
		},
		{
			name:   "AudioInfo",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.AudioInfo() },
			check: func(o *FFprobeOptions) bool {
				return o.SelectStreams == "a"
			},
		},
		{
			name:   "QuickInfo",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.QuickInfo() },
			check: func(o *FFprobeOptions) bool {
				return o.ProbeSize > 0 && o.AnalyzeDuration > 0
			},
		},
		{
			name:   "DeepAnalysis",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.DeepAnalysis() },
			check: func(o *FFprobeOptions) bool {
				return o.CountFrames && o.CountPackets
			},
		},
		{
			name:   "QualityControlAnalysis",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.QualityControlAnalysis() },
			check: func(o *FFprobeOptions) bool {
				return o.ShowError && o.ShowData && o.ShowDataHash
			},
		},
		{
			name:   "BroadcastQC",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.BroadcastQC() },
			check: func(o *FFprobeOptions) bool {
				return o.ErrorDetect != ""
			},
		},
		{
			name:   "StreamingQC",
			method: func(b *OptionsBuilder) *OptionsBuilder { return b.StreamingQC() },
			check: func(o *FFprobeOptions) bool {
				return o.HashAlgorithm == "md5"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOptionsBuilder()
			tt.method(builder)
			opts := builder.Build()

			if !tt.check(opts) {
				t.Errorf("%s did not configure options correctly", tt.name)
			}
		})
	}
}

func TestOptionsBuilder_Chaining(t *testing.T) {
	// Test that all methods return the builder for chaining
	opts := NewOptionsBuilder().
		Input("test.mp4").
		Format("mp4").
		JSON().
		ShowFormat().
		ShowStreams().
		CountFrames().
		ProbeSizeMB(5).
		TimeoutSeconds(30).
		Quiet().
		Build()

	if opts.Input != "test.mp4" {
		t.Error("chained Input failed")
	}
	if opts.Format != "mp4" {
		t.Error("chained Format failed")
	}
	if opts.OutputFormat != OutputJSON {
		t.Error("chained JSON failed")
	}
	if !opts.ShowFormat {
		t.Error("chained ShowFormat failed")
	}
	if !opts.ShowStreams {
		t.Error("chained ShowStreams failed")
	}
	if !opts.CountFrames {
		t.Error("chained CountFrames failed")
	}
	if opts.ProbeSize != 5*1024*1024 {
		t.Error("chained ProbeSizeMB failed")
	}
	if opts.Timeout != 30*time.Second {
		t.Error("chained TimeoutSeconds failed")
	}
	if opts.LogLevel != LogQuiet {
		t.Error("chained Quiet failed")
	}
}

func TestOptionsBuilder_ReadIntervals(t *testing.T) {
	t.Run("intervals string", func(t *testing.T) {
		builder := NewOptionsBuilder().ReadIntervals("0%+10%")
		opts := builder.Build()

		if opts.ReadIntervals != "0%+10%" {
			t.Errorf("expected ReadIntervals '0%%+10%%', got %q", opts.ReadIntervals)
		}
	})

	t.Run("single interval", func(t *testing.T) {
		builder := NewOptionsBuilder().ReadInterval("00:00:00", "00:00:10")
		opts := builder.Build()

		if opts.ReadIntervals != "00:00:00+00:00:10" {
			t.Errorf("expected ReadIntervals '00:00:00+00:00:10', got %q", opts.ReadIntervals)
		}
	})

	t.Run("percentage", func(t *testing.T) {
		builder := NewOptionsBuilder().ReadPercentage(0, 10)
		opts := builder.Build()

		if opts.ReadIntervals != "0%+10%" {
			t.Errorf("expected ReadIntervals '0%%+10%%', got %q", opts.ReadIntervals)
		}
	})
}

func TestOptionsBuilder_ShowEntries(t *testing.T) {
	t.Run("custom entries", func(t *testing.T) {
		builder := NewOptionsBuilder().ShowEntries("stream=codec_name,width,height")
		opts := builder.Build()

		if opts.ShowEntries != "stream=codec_name,width,height" {
			t.Errorf("unexpected ShowEntries: %q", opts.ShowEntries)
		}
	})

	t.Run("stream entries", func(t *testing.T) {
		builder := NewOptionsBuilder().ShowStreamEntries("codec_name", "width", "height")
		opts := builder.Build()

		if opts.ShowEntries != "stream=codec_name,width,height" {
			t.Errorf("expected 'stream=codec_name,width,height', got %q", opts.ShowEntries)
		}
	})

	t.Run("format entries", func(t *testing.T) {
		builder := NewOptionsBuilder().ShowFormatEntries("duration", "size")
		opts := builder.Build()

		if opts.ShowEntries != "format=duration,size" {
			t.Errorf("expected 'format=duration,size', got %q", opts.ShowEntries)
		}
	})
}
