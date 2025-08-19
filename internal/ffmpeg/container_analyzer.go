package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

// ContainerAnalyzer handles container format analysis and validation
type ContainerAnalyzer struct{}

// NewContainerAnalyzer creates a new container analyzer
func NewContainerAnalyzer() *ContainerAnalyzer {
	return &ContainerAnalyzer{}
}

// AnalyzeContainer analyzes container format from format information
func (ca *ContainerAnalyzer) AnalyzeContainer(format *FormatInfo) *ContainerAnalysis {
	if format == nil {
		return &ContainerAnalysis{
			Validation: &ContainerValidation{
				IsValid: false,
				Issues:  []string{"No format information available"},
			},
		}
	}

	analysis := &ContainerAnalysis{
		FormatName:     format.FormatName,
		FormatLongName: format.FormatLongName,
		FileName:       format.Filename,
		StreamCount:    format.NBStreams,
		ProgramCount:   format.NBPrograms,
		ProbeScore:     format.ProbeScore,
		Tags:           format.Tags,
	}

	// Parse duration and size
	if format.Duration != "" {
		if duration, err := strconv.ParseFloat(format.Duration, 64); err == nil {
			analysis.Duration = duration
		}
	}

	if format.Size != "" {
		if size, err := strconv.ParseInt(format.Size, 10, 64); err == nil {
			analysis.FileSize = size
		}
	}

	if format.BitRate != "" {
		if bitrate, err := strconv.ParseInt(format.BitRate, 10, 64); err == nil {
			analysis.OverallBitRate = bitrate
		}
	}

	// Determine container characteristics
	analysis.ContainerFamily = ca.getContainerFamily(format.FormatName)
	analysis.ContainerInfo = ca.getContainerInfo(analysis.ContainerFamily)
	analysis.IsStreamingFriendly = ca.isStreamingFriendly(analysis.ContainerFamily)
	analysis.SupportedCodecs = ca.getSupportedCodecs(analysis.ContainerFamily)
	analysis.Features = ca.getContainerFeatures(analysis.ContainerFamily)
	analysis.UseCases = ca.getContainerUseCases(analysis.ContainerFamily)

	// Validate container
	analysis.Validation = ca.validateContainer(analysis)

	return analysis
}

// getContainerFamily determines the container family
func (ca *ContainerAnalyzer) getContainerFamily(formatName string) string {
	// Handle common format name variations
	formatLower := strings.ToLower(formatName)

	// Split by comma and take first format
	if strings.Contains(formatLower, ",") {
		formatLower = strings.Split(formatLower, ",")[0]
		formatLower = strings.TrimSpace(formatLower)
	}

	containerMap := map[string]string{
		"mp4":          "MP4",
		"mov":          "QuickTime",
		"qt":           "QuickTime",
		"avi":          "AVI",
		"mkv":          "Matroska",
		"matroska":     "Matroska",
		"webm":         "WebM",
		"flv":          "FLV",
		"f4v":          "FLV",
		"ts":           "MPEG-TS",
		"mpegts":       "MPEG-TS",
		"m2ts":         "MPEG-TS",
		"mts":          "MPEG-TS",
		"ps":           "MPEG-PS",
		"mpegps":       "MPEG-PS",
		"wmv":          "ASF/WMV",
		"asf":          "ASF/WMV",
		"3gp":          "3GPP",
		"3g2":          "3GPP2",
		"ogg":          "Ogg",
		"ogv":          "Ogg",
		"mxf":          "MXF",
		"gxf":          "GXF",
		"rm":           "RealMedia",
		"rmvb":         "RealMedia",
		"nut":          "NUT",
		"yuv4mpegpipe": "Y4M",
		"rawvideo":     "Raw Video",
		"wav":          "WAV",
		"mp3":          "MP3",
		"aac":          "AAC",
		"flac":         "FLAC",
		"oga":          "Ogg Audio",
		"m4a":          "M4A",
		"wma":          "WMA",
	}

	if family, exists := containerMap[formatLower]; exists {
		return family
	}

	return formatName
}

// getContainerInfo returns detailed container information
func (ca *ContainerAnalyzer) getContainerInfo(containerFamily string) *ContainerInfo {
	info := &ContainerInfo{}

	switch containerFamily {
	case "MP4":
		info.Description = "MPEG-4 Part 14 container format"
		info.MimeType = "video/mp4"
		info.Extensions = []string{".mp4", ".m4v", ".m4a"}
		info.StandardizedBy = "ISO/IEC 14496-14"
		info.YearIntroduced = 2001
		info.IsOpenStandard = true

	case "QuickTime":
		info.Description = "Apple QuickTime movie format"
		info.MimeType = "video/quicktime"
		info.Extensions = []string{".mov", ".qt"}
		info.StandardizedBy = "Apple Inc."
		info.YearIntroduced = 1991
		info.IsOpenStandard = false

	case "Matroska":
		info.Description = "Open-source multimedia container"
		info.MimeType = "video/x-matroska"
		info.Extensions = []string{".mkv", ".mka", ".mks"}
		info.StandardizedBy = "Matroska.org"
		info.YearIntroduced = 2002
		info.IsOpenStandard = true

	case "WebM":
		info.Description = "Google WebM format for web"
		info.MimeType = "video/webm"
		info.Extensions = []string{".webm"}
		info.StandardizedBy = "Google"
		info.YearIntroduced = 2010
		info.IsOpenStandard = true

	case "AVI":
		info.Description = "Audio Video Interleave"
		info.MimeType = "video/x-msvideo"
		info.Extensions = []string{".avi"}
		info.StandardizedBy = "Microsoft"
		info.YearIntroduced = 1992
		info.IsOpenStandard = false

	case "MPEG-TS":
		info.Description = "MPEG Transport Stream"
		info.MimeType = "video/mp2t"
		info.Extensions = []string{".ts", ".m2ts", ".mts"}
		info.StandardizedBy = "ISO/IEC 13818-1"
		info.YearIntroduced = 1995
		info.IsOpenStandard = true

	case "FLV":
		info.Description = "Flash Video format"
		info.MimeType = "video/x-flv"
		info.Extensions = []string{".flv", ".f4v"}
		info.StandardizedBy = "Adobe"
		info.YearIntroduced = 2003
		info.IsOpenStandard = false

	default:
		info.Description = fmt.Sprintf("%s container format", containerFamily)
		info.Extensions = []string{}
		info.IsOpenStandard = true
	}

	return info
}

// isStreamingFriendly determines if container is streaming-friendly
func (ca *ContainerAnalyzer) isStreamingFriendly(containerFamily string) bool {
	streamingFriendly := map[string]bool{
		"MP4":       true,
		"WebM":      true,
		"MPEG-TS":   true,
		"FLV":       true,
		"Matroska":  true,
		"QuickTime": false, // Requires moov atom at beginning
		"AVI":       false, // Index at end
		"MXF":       false, // Professional format, not for streaming
	}

	if friendly, exists := streamingFriendly[containerFamily]; exists {
		return friendly
	}

	return false
}

// getSupportedCodecs returns commonly supported codecs for container
func (ca *ContainerAnalyzer) getSupportedCodecs(containerFamily string) *SupportedCodecs {
	codecs := &SupportedCodecs{}

	switch containerFamily {
	case "MP4":
		codecs.Video = []string{"H.264", "H.265", "AV1", "VP9"}
		codecs.Audio = []string{"AAC", "MP3", "AC-3", "E-AC-3"}
		codecs.Subtitle = []string{"mov_text", "TTML", "WebVTT"}

	case "WebM":
		codecs.Video = []string{"VP8", "VP9", "AV1"}
		codecs.Audio = []string{"Vorbis", "Opus"}
		codecs.Subtitle = []string{"WebVTT"}

	case "Matroska":
		codecs.Video = []string{"H.264", "H.265", "VP8", "VP9", "AV1", "MPEG-2", "MPEG-4"}
		codecs.Audio = []string{"AAC", "MP3", "AC-3", "DTS", "FLAC", "Vorbis", "Opus"}
		codecs.Subtitle = []string{"ASS", "SSA", "SRT", "VobSub", "PGS"}

	case "MPEG-TS":
		codecs.Video = []string{"H.264", "H.265", "MPEG-2"}
		codecs.Audio = []string{"AAC", "MP3", "AC-3", "E-AC-3"}
		codecs.Subtitle = []string{"DVB subtitles", "Teletext"}

	case "AVI":
		codecs.Video = []string{"H.264", "MPEG-4", "DivX", "XviD"}
		codecs.Audio = []string{"MP3", "AC-3", "PCM"}
		codecs.Subtitle = []string{"SRT (external)"}

	case "QuickTime":
		codecs.Video = []string{"H.264", "H.265", "ProRes", "DNxHD"}
		codecs.Audio = []string{"AAC", "PCM", "ALAC"}
		codecs.Subtitle = []string{"mov_text", "CEA-608"}

	case "FLV":
		codecs.Video = []string{"H.264", "VP6", "Sorenson Spark"}
		codecs.Audio = []string{"AAC", "MP3", "Speex"}
		codecs.Subtitle = []string{}

	default:
		codecs.Video = []string{}
		codecs.Audio = []string{}
		codecs.Subtitle = []string{}
	}

	return codecs
}

// getContainerFeatures returns container-specific features
func (ca *ContainerAnalyzer) getContainerFeatures(containerFamily string) []string {
	features := map[string][]string{
		"MP4": {
			"Fast start support", "Fragmented MP4", "Multiple tracks",
			"Metadata support", "DRM support", "Live streaming",
		},
		"WebM": {
			"Open source", "Royalty-free", "Web optimized",
			"Live streaming", "Cluster-based", "Seeking support",
		},
		"Matroska": {
			"Unlimited tracks", "Chapter support", "Attachment support",
			"Menu support", "Error recovery", "Seeking", "Subtitles",
		},
		"MPEG-TS": {
			"Broadcast standard", "Error resilience", "Multiplexing",
			"Live streaming", "Program streams", "PSI/SI tables",
		},
		"AVI": {
			"Simple structure", "Index support", "Multiple streams",
			"Legacy compatibility", "OpenDML extensions",
		},
		"QuickTime": {
			"Professional features", "Edit lists", "Timecode tracks",
			"Reference movies", "Alternate tracks", "User data",
		},
		"FLV": {
			"Flash compatibility", "Streaming optimized", "Cue points",
			"On2 VP6 support", "RTMP streaming",
		},
	}

	if containerFeatures, exists := features[containerFamily]; exists {
		return containerFeatures
	}

	return []string{}
}

// getContainerUseCases returns typical use cases for container
func (ca *ContainerAnalyzer) getContainerUseCases(containerFamily string) []string {
	useCases := map[string][]string{
		"MP4": {
			"Web streaming", "Mobile devices", "Digital distribution",
			"Social media", "OTT platforms", "Progressive download",
		},
		"WebM": {
			"Web browsers", "HTML5 video", "YouTube", "Open platforms",
			"Real-time communication", "Web conferencing",
		},
		"Matroska": {
			"High-quality archival", "Multi-language content", "Anime distribution",
			"Remuxing", "Fan subtitles", "Personal media libraries",
		},
		"MPEG-TS": {
			"Broadcast television", "Satellite transmission", "Cable TV",
			"IPTV", "Digital terrestrial", "Live streaming",
		},
		"AVI": {
			"Legacy systems", "Screen recording", "Editing workflows",
			"Intermediate formats", "Capture devices",
		},
		"QuickTime": {
			"Professional editing", "Post-production", "Apple ecosystem",
			"High-end video production", "Cinema workflows",
		},
		"FLV": {
			"Legacy Flash content", "RTMP streaming", "Flash players",
			"Online video platforms (legacy)", "Adobe ecosystem",
		},
	}

	if containerUseCases, exists := useCases[containerFamily]; exists {
		return containerUseCases
	}

	return []string{}
}

// validateContainer validates container format characteristics
func (ca *ContainerAnalyzer) validateContainer(analysis *ContainerAnalysis) *ContainerValidation {
	validation := &ContainerValidation{
		IsValid:         true,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Validate probe score
	if analysis.ProbeScore < 50 {
		validation.Issues = append(validation.Issues,
			fmt.Sprintf("Low probe score (%d) - container format detection may be unreliable", analysis.ProbeScore))
		if analysis.ProbeScore < 25 {
			validation.IsValid = false
		}
	}

	// Check for deprecated formats
	deprecatedFormats := map[string]bool{
		"FLV":       true,
		"RealMedia": true,
		"AVI":       true, // Not deprecated but showing age
	}

	if deprecatedFormats[analysis.ContainerFamily] {
		validation.Recommendations = append(validation.Recommendations,
			fmt.Sprintf("%s is a legacy format - consider converting to MP4 or WebM for better compatibility", analysis.ContainerFamily))
	}

	// Check file size consistency
	if analysis.FileSize > 0 && analysis.Duration > 0 && analysis.OverallBitRate > 0 {
		expectedSize := int64((analysis.Duration * float64(analysis.OverallBitRate)) / 8)
		sizeDiff := float64(analysis.FileSize-expectedSize) / float64(expectedSize)

		if sizeDiff > 0.1 || sizeDiff < -0.1 { // 10% tolerance
			validation.Issues = append(validation.Issues,
				"File size inconsistent with duration and bitrate - possible corruption or incomplete file")
		}
	}

	// Streaming recommendations
	if !analysis.IsStreamingFriendly {
		validation.Recommendations = append(validation.Recommendations,
			"Container format is not optimized for streaming - consider MP4 with fast-start for web delivery")
	}

	// Check for missing metadata
	if len(analysis.Tags) == 0 {
		validation.Recommendations = append(validation.Recommendations,
			"No metadata tags found - consider adding title, artist, and other descriptive information")
	}

	// Program count validation for transport streams
	if analysis.ContainerFamily == "MPEG-TS" && analysis.ProgramCount == 0 {
		validation.Issues = append(validation.Issues,
			"MPEG-TS container has no programs - this may indicate a malformed stream")
		validation.IsValid = false
	}

	// Stream count validation
	if analysis.StreamCount == 0 {
		validation.Issues = append(validation.Issues,
			"Container has no streams - this indicates an empty or corrupted file")
		validation.IsValid = false
	}

	// Duration validation
	if analysis.Duration <= 0 {
		validation.Issues = append(validation.Issues,
			"Container has no duration information - may indicate live stream or corrupted file")
	}

	return validation
}
