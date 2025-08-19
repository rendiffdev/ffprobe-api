package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

// CodecAnalyzer handles codec profile and level analysis
type CodecAnalyzer struct{}

// NewCodecAnalyzer creates a new codec analyzer
func NewCodecAnalyzer() *CodecAnalyzer {
	return &CodecAnalyzer{}
}

// AnalyzeCodecs analyzes codec profiles and levels from stream information
func (ca *CodecAnalyzer) AnalyzeCodecs(streams []StreamInfo) *CodecAnalysis {
	analysis := &CodecAnalysis{
		VideoCodecs:  make(map[int]*VideoCodecInfo),
		AudioCodecs:  make(map[int]*AudioCodecInfo),
		CodecSummary: make(map[string]int),
	}

	for _, stream := range streams {
		// Count codec usage
		if stream.CodecName != "" {
			analysis.CodecSummary[stream.CodecName]++
		}

		switch strings.ToLower(stream.CodecType) {
		case "video":
			videoCodec := ca.analyzeVideoCodec(stream)
			if videoCodec != nil {
				analysis.VideoCodecs[stream.Index] = videoCodec
				if !ca.containsCodec(analysis.SupportedVideoCodecs, videoCodec.CodecFamily) {
					analysis.SupportedVideoCodecs = append(analysis.SupportedVideoCodecs, videoCodec.CodecFamily)
				}
			}
		case "audio":
			audioCodec := ca.analyzeAudioCodec(stream)
			if audioCodec != nil {
				analysis.AudioCodecs[stream.Index] = audioCodec
				if !ca.containsCodec(analysis.SupportedAudioCodecs, audioCodec.CodecFamily) {
					analysis.SupportedAudioCodecs = append(analysis.SupportedAudioCodecs, audioCodec.CodecFamily)
				}
			}
		}
	}

	// Determine overall characteristics
	analysis.HasModernCodecs = ca.hasModernCodecs(analysis.VideoCodecs, analysis.AudioCodecs)
	analysis.HasLegacyCodecs = ca.hasLegacyCodecs(analysis.VideoCodecs, analysis.AudioCodecs)
	analysis.IsStreamingOptimized = ca.isStreamingOptimized(analysis.VideoCodecs)
	analysis.Validation = ca.validateCodecs(analysis)

	return analysis
}

// analyzeVideoCodec extracts video codec information
func (ca *CodecAnalyzer) analyzeVideoCodec(stream StreamInfo) *VideoCodecInfo {
	codec := &VideoCodecInfo{
		CodecName:     stream.CodecName,
		CodecLongName: stream.CodecLongName,
		Profile:       stream.Profile,
		Level:         stream.Level,
	}

	// Determine codec family
	codec.CodecFamily = ca.getVideoCodecFamily(stream.CodecName)

	// Parse profile information
	codec.ProfileInfo = ca.parseVideoProfile(stream.CodecName, stream.Profile)

	// Parse level information
	codec.LevelInfo = ca.parseVideoLevel(stream.CodecName, stream.Level)

	// Determine codec generation and features
	codec.Generation = ca.getCodecGeneration(codec.CodecFamily)
	codec.Features = ca.getCodecFeatures(codec.CodecFamily, codec.Profile)

	// Check hardware support
	codec.HardwareSupport = ca.getHardwareSupport(codec.CodecFamily, codec.Profile)

	// Validate profile/level combination
	codec.IsValid = ca.validateVideoProfileLevel(codec)

	return codec
}

// analyzeAudioCodec extracts audio codec information
func (ca *CodecAnalyzer) analyzeAudioCodec(stream StreamInfo) *AudioCodecInfo {
	codec := &AudioCodecInfo{
		CodecName:     stream.CodecName,
		CodecLongName: stream.CodecLongName,
		Profile:       stream.Profile,
		SampleFormat:  stream.SampleFmt,
		SampleRate:    stream.SampleRate,
		Channels:      stream.Channels,
		ChannelLayout: stream.ChannelLayout,
	}

	// Determine codec family
	codec.CodecFamily = ca.getAudioCodecFamily(stream.CodecName)

	// Parse profile information
	codec.ProfileInfo = ca.parseAudioProfile(stream.CodecName, stream.Profile)

	// Determine codec characteristics
	codec.IsLossless = ca.isLosslessAudio(stream.CodecName)
	codec.IsSurround = ca.isSurroundAudio(stream.Channels, stream.ChannelLayout)

	// Check hardware support
	codec.HardwareSupport = ca.getAudioHardwareSupport(codec.CodecFamily)

	// Validate configuration
	codec.IsValid = ca.validateAudioConfig(codec)

	return codec
}

// getVideoCodecFamily determines the codec family
func (ca *CodecAnalyzer) getVideoCodecFamily(codecName string) string {
	codecMap := map[string]string{
		"h264":       "H.264/AVC",
		"avc":        "H.264/AVC",
		"h265":       "H.265/HEVC",
		"hevc":       "H.265/HEVC",
		"av1":        "AV1",
		"vp8":        "VP8",
		"vp9":        "VP9",
		"mpeg2video": "MPEG-2",
		"mpeg4":      "MPEG-4",
		"xvid":       "MPEG-4",
		"divx":       "MPEG-4",
		"prores":     "ProRes",
		"dnxhd":      "DNxHD",
		"dnxhr":      "DNxHR",
		"jpeg2000":   "JPEG 2000",
		"mjpeg":      "MJPEG",
		"rawvideo":   "Raw Video",
	}

	codecLower := strings.ToLower(codecName)
	if family, exists := codecMap[codecLower]; exists {
		return family
	}

	return codecName
}

// parseVideoProfile parses video codec profile information
func (ca *CodecAnalyzer) parseVideoProfile(codecName, profile string) *ProfileInfo {
	if profile == "" {
		return nil
	}

	profileInfo := &ProfileInfo{
		Name: profile,
	}

	codecLower := strings.ToLower(codecName)
	profileLower := strings.ToLower(profile)

	switch codecLower {
	case "h264", "avc":
		profileInfo.Description = ca.getH264ProfileDescription(profileLower)
		profileInfo.Capabilities = ca.getH264ProfileCapabilities(profileLower)
	case "h265", "hevc":
		profileInfo.Description = ca.getH265ProfileDescription(profileLower)
		profileInfo.Capabilities = ca.getH265ProfileCapabilities(profileLower)
	case "av1":
		profileInfo.Description = ca.getAV1ProfileDescription(profileLower)
		profileInfo.Capabilities = ca.getAV1ProfileCapabilities(profileLower)
	case "vp9":
		profileInfo.Description = ca.getVP9ProfileDescription(profileLower)
		profileInfo.Capabilities = ca.getVP9ProfileCapabilities(profileLower)
	}

	return profileInfo
}

// parseVideoLevel parses video codec level information
func (ca *CodecAnalyzer) parseVideoLevel(codecName string, level int) *LevelInfo {
	if level == 0 {
		return nil
	}

	levelInfo := &LevelInfo{
		Level: level,
	}

	codecLower := strings.ToLower(codecName)

	switch codecLower {
	case "h264", "avc":
		levelInfo.Description = ca.getH264LevelDescription(level)
		levelInfo.MaxResolution = ca.getH264LevelMaxResolution(level)
		levelInfo.MaxFrameRate = ca.getH264LevelMaxFrameRate(level)
	case "h265", "hevc":
		levelInfo.Description = ca.getH265LevelDescription(level)
		levelInfo.MaxResolution = ca.getH265LevelMaxResolution(level)
		levelInfo.MaxFrameRate = ca.getH265LevelMaxFrameRate(level)
	}

	return levelInfo
}

// getH264ProfileDescription returns H.264 profile description
func (ca *CodecAnalyzer) getH264ProfileDescription(profile string) string {
	profiles := map[string]string{
		"baseline":             "Baseline Profile - Basic features, mobile/web optimized",
		"constrained baseline": "Constrained Baseline - Baseline + additional constraints",
		"main":                 "Main Profile - Interlaced content support",
		"extended":             "Extended Profile - Error resilience features",
		"high":                 "High Profile - 8x8 DCT, quantization scaling",
		"high 10":              "High 10 Profile - 10-bit depth support",
		"high 422":             "High 422 Profile - 4:2:2 chroma sampling",
		"high 444":             "High 444 Profile - 4:4:4 chroma sampling",
	}

	if desc, exists := profiles[profile]; exists {
		return desc
	}
	return "Unknown H.264 Profile"
}

// getH264ProfileCapabilities returns H.264 profile capabilities
func (ca *CodecAnalyzer) getH264ProfileCapabilities(profile string) []string {
	capabilities := map[string][]string{
		"baseline": {"4:2:0 chroma", "8-bit depth", "Progressive"},
		"main":     {"4:2:0 chroma", "8-bit depth", "Interlaced", "B-frames"},
		"high":     {"4:2:0 chroma", "8-bit depth", "8x8 DCT", "Quantization scaling"},
		"high 10":  {"4:2:0 chroma", "10-bit depth", "8x8 DCT", "Quantization scaling"},
		"high 422": {"4:2:2 chroma", "10-bit depth", "8x8 DCT", "Quantization scaling"},
		"high 444": {"4:4:4 chroma", "14-bit depth", "8x8 DCT", "Quantization scaling"},
	}

	if caps, exists := capabilities[profile]; exists {
		return caps
	}
	return []string{}
}

// getH264LevelDescription returns H.264 level description
func (ca *CodecAnalyzer) getH264LevelDescription(level int) string {
	levels := map[int]string{
		10: "Level 1.0 - QCIF/176x144",
		11: "Level 1.1 - CIF/352x288",
		12: "Level 1.2 - CIF/352x288",
		13: "Level 1.3 - CIF/352x288",
		20: "Level 2.0 - CIF/352x288",
		21: "Level 2.1 - HHR/352x480",
		22: "Level 2.2 - SD/720x480",
		30: "Level 3.0 - SD/720x480",
		31: "Level 3.1 - 720p/1280x720",
		32: "Level 3.2 - 720p/1280x720",
		40: "Level 4.0 - 1080p/1920x1080",
		41: "Level 4.1 - 1080p/1920x1080",
		42: "Level 4.2 - 1080p/1920x1080",
		50: "Level 5.0 - 4K/4096x2304",
		51: "Level 5.1 - 4K/4096x2304",
		52: "Level 5.2 - 4K/4096x2304",
		60: "Level 6.0 - 8K/8192x4320",
		61: "Level 6.1 - 8K/8192x4320",
		62: "Level 6.2 - 8K/8192x4320",
	}

	if desc, exists := levels[level]; exists {
		return desc
	}
	return fmt.Sprintf("Level %d", level)
}

// getCodecGeneration determines codec generation
func (ca *CodecAnalyzer) getCodecGeneration(codecFamily string) string {
	generations := map[string]string{
		"H.264/AVC":  "4th Generation",
		"H.265/HEVC": "5th Generation",
		"AV1":        "5th Generation",
		"VP9":        "5th Generation",
		"VP8":        "4th Generation",
		"MPEG-2":     "2nd Generation",
		"MPEG-4":     "3rd Generation",
	}

	if gen, exists := generations[codecFamily]; exists {
		return gen
	}
	return "Unknown Generation"
}

// getCodecFeatures returns codec features
func (ca *CodecAnalyzer) getCodecFeatures(codecFamily, profile string) []string {
	features := []string{}

	switch codecFamily {
	case "H.264/AVC":
		features = append(features, "DCT Transform", "Variable Block Size", "Intra Prediction")
		if strings.Contains(strings.ToLower(profile), "high") {
			features = append(features, "8x8 DCT", "Quantization Scaling")
		}
	case "H.265/HEVC":
		features = append(features, "CTU/CTB", "Advanced Motion Vectors", "SAO Filtering", "Tiles")
	case "AV1":
		features = append(features, "Film Grain", "CDEF Filtering", "Loop Restoration", "Superblocks")
	case "VP9":
		features = append(features, "Tiles", "Frame Parallel Decoding", "Lossless Mode")
	}

	return features
}

// getHardwareSupport returns hardware support information
func (ca *CodecAnalyzer) getHardwareSupport(codecFamily, profile string) []string {
	support := []string{}

	switch codecFamily {
	case "H.264/AVC":
		support = append(support, "Universal Hardware Support", "Mobile Devices", "Smart TVs", "Browsers")
	case "H.265/HEVC":
		support = append(support, "Modern Hardware Support", "4K Streaming", "Mobile (iOS/Android)")
		if strings.Contains(strings.ToLower(profile), "main") {
			support = append(support, "Broad Hardware Decode")
		}
	case "AV1":
		support = append(support, "Limited Hardware Support", "Growing Support", "Chrome/Firefox")
	case "VP9":
		support = append(support, "Google Ecosystem", "YouTube", "Chrome", "Android")
	}

	return support
}

// validateVideoProfileLevel validates profile/level combination
func (ca *CodecAnalyzer) validateVideoProfileLevel(codec *VideoCodecInfo) bool {
	// Basic validation - ensure profile and level are reasonable
	if codec.Profile == "" && codec.Level == 0 {
		return false // Both missing
	}

	// Codec-specific validation
	switch strings.ToLower(codec.CodecName) {
	case "h264", "avc":
		return ca.validateH264ProfileLevel(codec.Profile, codec.Level)
	case "h265", "hevc":
		return ca.validateH265ProfileLevel(codec.Profile, codec.Level)
	}

	return true // Assume valid for other codecs
}

// validateH264ProfileLevel validates H.264 profile/level combination
func (ca *CodecAnalyzer) validateH264ProfileLevel(profile string, level int) bool {
	// High profiles require level 3.0 or higher
	if strings.Contains(strings.ToLower(profile), "high") && level < 30 {
		return false
	}

	// Level range validation
	if level < 10 || level > 62 {
		return false
	}

	return true
}

// getAudioCodecFamily determines audio codec family
func (ca *CodecAnalyzer) getAudioCodecFamily(codecName string) string {
	codecMap := map[string]string{
		"aac":       "AAC",
		"mp3":       "MP3",
		"ac3":       "AC-3",
		"eac3":      "E-AC-3",
		"dts":       "DTS",
		"dtshd":     "DTS-HD",
		"truehd":    "TrueHD",
		"flac":      "FLAC",
		"alac":      "ALAC",
		"opus":      "Opus",
		"vorbis":    "Vorbis",
		"pcm":       "PCM",
		"pcm_s16le": "PCM",
		"pcm_s24le": "PCM",
		"pcm_s32le": "PCM",
	}

	codecLower := strings.ToLower(codecName)
	if family, exists := codecMap[codecLower]; exists {
		return family
	}

	// Handle PCM variants
	if strings.HasPrefix(codecLower, "pcm") {
		return "PCM"
	}

	return codecName
}

// parseAudioProfile parses audio codec profile
func (ca *CodecAnalyzer) parseAudioProfile(codecName, profile string) *ProfileInfo {
	if profile == "" {
		return nil
	}

	profileInfo := &ProfileInfo{
		Name: profile,
	}

	codecLower := strings.ToLower(codecName)

	switch codecLower {
	case "aac":
		profileInfo.Description = ca.getAACProfileDescription(profile)
	case "ac3", "eac3":
		profileInfo.Description = ca.getAC3ProfileDescription(profile)
	case "dts":
		profileInfo.Description = ca.getDTSProfileDescription(profile)
	}

	return profileInfo
}

// isLosslessAudio determines if audio codec is lossless
func (ca *CodecAnalyzer) isLosslessAudio(codecName string) bool {
	losslessCodecs := map[string]bool{
		"flac":      true,
		"alac":      true,
		"pcm":       true,
		"pcm_s16le": true,
		"pcm_s24le": true,
		"pcm_s32le": true,
		"truehd":    true,
		"dtshd":     true,
	}

	codecLower := strings.ToLower(codecName)
	if strings.HasPrefix(codecLower, "pcm") {
		return true
	}

	return losslessCodecs[codecLower]
}

// isSurroundAudio determines if audio has surround sound
func (ca *CodecAnalyzer) isSurroundAudio(channels int, channelLayout string) bool {
	if channels > 2 {
		return true
	}

	surroundLayouts := []string{
		"5.1", "7.1", "surround", "quad",
	}

	layoutLower := strings.ToLower(channelLayout)
	for _, layout := range surroundLayouts {
		if strings.Contains(layoutLower, layout) {
			return true
		}
	}

	return false
}

// validateCodecs validates overall codec configuration
func (ca *CodecAnalyzer) validateCodecs(analysis *CodecAnalysis) *CodecValidation {
	validation := &CodecValidation{
		IsValid:         true,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Check for invalid codecs
	for streamIndex, videoCodec := range analysis.VideoCodecs {
		if !videoCodec.IsValid {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Video stream %d has invalid codec configuration", streamIndex))
			validation.IsValid = false
		}
	}

	for streamIndex, audioCodec := range analysis.AudioCodecs {
		if !audioCodec.IsValid {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Audio stream %d has invalid codec configuration", streamIndex))
			validation.IsValid = false
		}
	}

	// Provide recommendations
	if analysis.HasLegacyCodecs {
		validation.Recommendations = append(validation.Recommendations,
			"Legacy codecs detected - consider upgrading to modern codecs for better compression")
	}

	if !analysis.HasModernCodecs {
		validation.Recommendations = append(validation.Recommendations,
			"No modern codecs detected - consider using H.265/HEVC or AV1 for improved efficiency")
	}

	if !analysis.IsStreamingOptimized {
		validation.Recommendations = append(validation.Recommendations,
			"Content may not be optimized for streaming - consider using streaming-friendly profiles")
	}

	return validation
}

// hasModernCodecs checks for modern codec usage
func (ca *CodecAnalyzer) hasModernCodecs(videoCodecs map[int]*VideoCodecInfo, audioCodecs map[int]*AudioCodecInfo) bool {
	modernVideoCodecs := map[string]bool{
		"H.265/HEVC": true,
		"AV1":        true,
		"VP9":        true,
	}

	for _, codec := range videoCodecs {
		if modernVideoCodecs[codec.CodecFamily] {
			return true
		}
	}

	return false
}

// hasLegacyCodecs checks for legacy codec usage
func (ca *CodecAnalyzer) hasLegacyCodecs(videoCodecs map[int]*VideoCodecInfo, audioCodecs map[int]*AudioCodecInfo) bool {
	legacyVideoCodecs := map[string]bool{
		"MPEG-2": true,
		"MPEG-4": true,
		"VP8":    true,
	}

	legacyAudioCodecs := map[string]bool{
		"MP3":  true,
		"AC-3": true,
	}

	for _, codec := range videoCodecs {
		if legacyVideoCodecs[codec.CodecFamily] {
			return true
		}
	}

	for _, codec := range audioCodecs {
		if legacyAudioCodecs[codec.CodecFamily] {
			return true
		}
	}

	return false
}

// isStreamingOptimized checks if content is optimized for streaming
func (ca *CodecAnalyzer) isStreamingOptimized(videoCodecs map[int]*VideoCodecInfo) bool {
	for _, codec := range videoCodecs {
		// Check for streaming-friendly profiles
		profileLower := strings.ToLower(codec.Profile)
		if strings.Contains(profileLower, "baseline") || strings.Contains(profileLower, "main") {
			return true
		}
	}
	return false
}

// Helper functions for additional profile/level parsing
func (ca *CodecAnalyzer) getH265ProfileDescription(profile string) string {
	profiles := map[string]string{
		"main":    "Main Profile - 8-bit 4:2:0",
		"main10":  "Main 10 Profile - 10-bit 4:2:0",
		"main422": "Main 422 Profile - 10-bit 4:2:2",
		"main444": "Main 444 Profile - 10-bit 4:4:4",
	}
	if desc, exists := profiles[profile]; exists {
		return desc
	}
	return "Unknown HEVC Profile"
}

func (ca *CodecAnalyzer) getH265ProfileCapabilities(profile string) []string {
	capabilities := map[string][]string{
		"main":    {"8-bit depth", "4:2:0 chroma"},
		"main10":  {"10-bit depth", "4:2:0 chroma"},
		"main422": {"10-bit depth", "4:2:2 chroma"},
		"main444": {"10-bit depth", "4:4:4 chroma"},
	}
	if caps, exists := capabilities[profile]; exists {
		return caps
	}
	return []string{}
}

func (ca *CodecAnalyzer) getAV1ProfileDescription(profile string) string {
	profiles := map[string]string{
		"main":         "Main Profile - 8/10-bit 4:2:0",
		"high":         "High Profile - 8/10-bit 4:4:4",
		"professional": "Professional Profile - 12-bit support",
	}
	if desc, exists := profiles[profile]; exists {
		return desc
	}
	return "Unknown AV1 Profile"
}

func (ca *CodecAnalyzer) getAV1ProfileCapabilities(profile string) []string {
	return []string{"Advanced compression", "Royalty-free", "Future-proof"}
}

func (ca *CodecAnalyzer) getVP9ProfileDescription(profile string) string {
	profiles := map[string]string{
		"profile 0": "Profile 0 - 8-bit 4:2:0",
		"profile 1": "Profile 1 - 8-bit 4:2:2/4:4:4",
		"profile 2": "Profile 2 - 10/12-bit 4:2:0",
		"profile 3": "Profile 3 - 10/12-bit 4:2:2/4:4:4",
	}
	if desc, exists := profiles[profile]; exists {
		return desc
	}
	return "Unknown VP9 Profile"
}

func (ca *CodecAnalyzer) getVP9ProfileCapabilities(profile string) []string {
	return []string{"Tiles", "Frame parallel", "Lossless"}
}

func (ca *CodecAnalyzer) getH264LevelMaxResolution(level int) string {
	resolutions := map[int]string{
		30: "720x480",
		31: "1280x720",
		40: "1920x1080",
		50: "4096x2304",
		60: "8192x4320",
	}
	if res, exists := resolutions[level]; exists {
		return res
	}
	return "Unknown"
}

func (ca *CodecAnalyzer) getH264LevelMaxFrameRate(level int) string {
	frameRates := map[int]string{
		30: "30 fps",
		31: "60 fps",
		40: "30 fps",
		50: "30 fps",
		60: "30 fps",
	}
	if fr, exists := frameRates[level]; exists {
		return fr
	}
	return "Unknown"
}

func (ca *CodecAnalyzer) getH265LevelDescription(level int) string {
	return fmt.Sprintf("HEVC Level %d", level)
}

func (ca *CodecAnalyzer) getH265LevelMaxResolution(level int) string {
	return "Varies by tier"
}

func (ca *CodecAnalyzer) getH265LevelMaxFrameRate(level int) string {
	return "Varies by tier"
}

func (ca *CodecAnalyzer) validateH265ProfileLevel(profile string, level int) bool {
	return level >= 30 && level <= 186 // HEVC level range
}

func (ca *CodecAnalyzer) getAACProfileDescription(profile string) string {
	profiles := map[string]string{
		"lc":   "Low Complexity - Most common AAC profile",
		"he":   "High Efficiency - AAC+ with SBR",
		"hev2": "High Efficiency v2 - AAC+ with SBR and PS",
	}
	if desc, exists := profiles[profile]; exists {
		return desc
	}
	return "Unknown AAC Profile"
}

func (ca *CodecAnalyzer) getAC3ProfileDescription(profile string) string {
	return "Dolby Digital Profile"
}

func (ca *CodecAnalyzer) getDTSProfileDescription(profile string) string {
	return "DTS Profile"
}

func (ca *CodecAnalyzer) getAudioHardwareSupport(codecFamily string) []string {
	support := map[string][]string{
		"AAC":    {"Universal Support", "Mobile", "Streaming"},
		"MP3":    {"Universal Support", "Legacy"},
		"AC-3":   {"Home Theater", "Surround Sound"},
		"E-AC-3": {"Advanced Surround", "Streaming"},
		"DTS":    {"Home Theater", "High Quality"},
		"TrueHD": {"High-End Audio", "Lossless"},
		"FLAC":   {"Audiophile", "Lossless"},
		"Opus":   {"Low Latency", "VoIP"},
	}
	if sup, exists := support[codecFamily]; exists {
		return sup
	}
	return []string{}
}

func (ca *CodecAnalyzer) validateAudioConfig(codec *AudioCodecInfo) bool {
	// Basic validation
	if codec.CodecName == "" {
		return false
	}

	// Sample rate validation
	if codec.SampleRate != "" {
		if rate, err := strconv.Atoi(codec.SampleRate); err == nil {
			if rate < 8000 || rate > 192000 {
				return false // Unreasonable sample rate
			}
		}
	}

	return true
}

func (ca *CodecAnalyzer) containsCodec(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
