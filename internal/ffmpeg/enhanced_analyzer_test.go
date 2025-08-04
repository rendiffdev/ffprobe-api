package ffmpeg

import (
	"testing"
)

func TestEnhancedAnalyzer_AnalyzeStreamCounts(t *testing.T) {
	analyzer := NewEnhancedAnalyzer()

	streams := []StreamInfo{
		{CodecType: "video", CodecName: "h264"},
		{CodecType: "audio", CodecName: "aac"},
		{CodecType: "audio", CodecName: "mp3"},
		{CodecType: "subtitle", CodecName: "mov_text"},
	}

	counts := analyzer.analyzeStreamCounts(streams)

	if counts.TotalStreams != 4 {
		t.Errorf("Expected 4 total streams, got %d", counts.TotalStreams)
	}
	if counts.VideoStreams != 1 {
		t.Errorf("Expected 1 video stream, got %d", counts.VideoStreams)
	}
	if counts.AudioStreams != 2 {
		t.Errorf("Expected 2 audio streams, got %d", counts.AudioStreams)
	}
	if counts.SubtitleStreams != 1 {
		t.Errorf("Expected 1 subtitle stream, got %d", counts.SubtitleStreams)
	}
}

func TestEnhancedAnalyzer_ExtractChromaSubsampling(t *testing.T) {
	analyzer := NewEnhancedAnalyzer()

	testCases := []struct {
		pixfmt   string
		expected string
	}{
		{"yuv420p", "4:2:0"},
		{"yuv422p", "4:2:2"},
		{"yuv444p", "4:4:4"},
		{"yuvj420p", "4:2:0"},
		{"unknown", ""},
	}

	for _, tc := range testCases {
		result := analyzer.extractChromaSubsampling(tc.pixfmt)
		if result != tc.expected {
			t.Errorf("For pixfmt %s, expected %s, got %s", tc.pixfmt, tc.expected, result)
		}
	}
}

func TestEnhancedAnalyzer_AnalyzeGOPStructure(t *testing.T) {
	analyzer := NewEnhancedAnalyzer()

	// Create test frames with keyframes at positions 0, 12, 24
	frames := []FrameInfo{
		{MediaType: "video", KeyFrame: 1}, // I-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 0}, // P-frame
		{MediaType: "video", KeyFrame: 1}, // I-frame
	}

	gop := analyzer.analyzeGOPStructure(frames)

	if gop.TotalFrameCount != 13 {
		t.Errorf("Expected 13 total frames, got %d", gop.TotalFrameCount)
	}
	if gop.KeyFrameCount != 2 {
		t.Errorf("Expected 2 key frames, got %d", gop.KeyFrameCount)
	}
	if gop.AverageGOPSize == nil || *gop.AverageGOPSize != 12.0 {
		t.Errorf("Expected average GOP size of 12.0, got %v", gop.AverageGOPSize)
	}
}

func TestEnhancedAnalyzer_AnalyzeResult(t *testing.T) {
	analyzer := NewEnhancedAnalyzer()

	result := &FFprobeResult{
		Streams: []StreamInfo{
			{
				CodecType:    "video",
				CodecName:    "h264",
				PixFmt:       "yuv420p",
				ColorSpace:   "bt709",
				BitRate:      "5000000",
			},
			{
				CodecType: "audio",
				CodecName: "aac",
				BitRate:   "128000",
			},
		},
		Frames: []FrameInfo{
			{MediaType: "video", KeyFrame: 1},
			{MediaType: "video", KeyFrame: 0},
			{MediaType: "video", KeyFrame: 0},
		},
	}

	err := analyzer.AnalyzeResult(result)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.EnhancedAnalysis == nil {
		t.Fatal("Expected enhanced analysis to be populated")
	}

	// Check stream counts
	if result.EnhancedAnalysis.StreamCounts == nil {
		t.Fatal("Expected stream counts to be populated")
	}
	if result.EnhancedAnalysis.StreamCounts.VideoStreams != 1 {
		t.Errorf("Expected 1 video stream, got %d", result.EnhancedAnalysis.StreamCounts.VideoStreams)
	}
	if result.EnhancedAnalysis.StreamCounts.AudioStreams != 1 {
		t.Errorf("Expected 1 audio stream, got %d", result.EnhancedAnalysis.StreamCounts.AudioStreams)
	}

	// Check video analysis
	if result.EnhancedAnalysis.VideoAnalysis == nil {
		t.Fatal("Expected video analysis to be populated")
	}
	if result.EnhancedAnalysis.VideoAnalysis.ChromaSubsampling == nil || *result.EnhancedAnalysis.VideoAnalysis.ChromaSubsampling != "4:2:0" {
		t.Errorf("Expected chroma subsampling 4:2:0, got %v", result.EnhancedAnalysis.VideoAnalysis.ChromaSubsampling)
	}

	// Check GOP analysis
	if result.EnhancedAnalysis.GOPAnalysis == nil {
		t.Fatal("Expected GOP analysis to be populated")
	}
	if result.EnhancedAnalysis.GOPAnalysis.KeyFrameCount != 1 {
		t.Errorf("Expected 1 key frame, got %d", result.EnhancedAnalysis.GOPAnalysis.KeyFrameCount)
	}
}

func TestEnhancedAnalyzer_FrameStatistics(t *testing.T) {
	analyzer := NewEnhancedAnalyzer()

	frames := []FrameInfo{
		{MediaType: "video", PictType: "I", PktSize: "10000"},
		{MediaType: "video", PictType: "P", PktSize: "5000"},
		{MediaType: "video", PictType: "B", PktSize: "3000"},
		{MediaType: "video", PictType: "P", PktSize: "5500"},
	}

	stats := analyzer.analyzeFrameStatistics(frames)

	if stats.TotalFrames != 4 {
		t.Errorf("Expected 4 total frames, got %d", stats.TotalFrames)
	}
	if stats.IFrames != 1 {
		t.Errorf("Expected 1 I-frame, got %d", stats.IFrames)
	}
	if stats.PFrames != 2 {
		t.Errorf("Expected 2 P-frames, got %d", stats.PFrames)
	}
	if stats.BFrames != 1 {
		t.Errorf("Expected 1 B-frame, got %d", stats.BFrames)
	}

	if stats.AverageFrameSize == nil {
		t.Error("Expected average frame size to be calculated")
	} else {
		expected := (10000.0 + 5000.0 + 3000.0 + 5500.0) / 4.0
		if *stats.AverageFrameSize != expected {
			t.Errorf("Expected average frame size %.2f, got %.2f", expected, *stats.AverageFrameSize)
		}
	}

	if stats.MaxFrameSize == nil || *stats.MaxFrameSize != 10000 {
		t.Errorf("Expected max frame size 10000, got %v", stats.MaxFrameSize)
	}
	if stats.MinFrameSize == nil || *stats.MinFrameSize != 3000 {
		t.Errorf("Expected min frame size 3000, got %v", stats.MinFrameSize)
	}
}