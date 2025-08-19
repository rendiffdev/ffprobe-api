package hls

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ValidateHLSRequest validates an HLS analysis request
func ValidateHLSRequest(request *HLSAnalysisRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate manifest URL
	if err := validateManifestURL(request.ManifestURL); err != nil {
		return fmt.Errorf("invalid manifest URL: %w", err)
	}

	// Validate segment limits
	if request.MaxSegments < 0 {
		return fmt.Errorf("max segments cannot be negative")
	}

	if request.MaxSegments > 1000 {
		return fmt.Errorf("max segments cannot exceed 1000")
	}

	// Validate timeout (timeout is in seconds, convert to duration for comparison)
	if request.Timeout > 0 && time.Duration(request.Timeout)*time.Second > 10*time.Minute {
		return fmt.Errorf("timeout cannot exceed 10 minutes")
	}

	return nil
}

// validateManifestURL validates HLS manifest URL
func validateManifestURL(manifestURL string) error {
	if strings.TrimSpace(manifestURL) == "" {
		return fmt.Errorf("manifest URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(manifestURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	validSchemes := []string{"http", "https"}
	validScheme := false
	for _, scheme := range validSchemes {
		if parsedURL.Scheme == scheme {
			validScheme = true
			break
		}
	}

	if !validScheme {
		return fmt.Errorf("unsupported URL scheme: %s (only http/https allowed)", parsedURL.Scheme)
	}

	// Check for valid HLS file extensions
	validExtensions := []string{".m3u8", ".m3u"}
	validExt := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(parsedURL.Path), ext) {
			validExt = true
			break
		}
	}

	if !validExt {
		return fmt.Errorf("URL does not appear to be an HLS manifest (missing .m3u8 extension)")
	}

	// Block localhost and private IPs for security
	host := strings.ToLower(parsedURL.Hostname())
	blockedHosts := []string{"localhost", "127.0.0.1", "0.0.0.0", "::1"}
	for _, blocked := range blockedHosts {
		if host == blocked {
			return fmt.Errorf("blocked host: %s", host)
		}
	}

	// Check for private IP ranges
	if isPrivateIP(host) {
		return fmt.Errorf("private IP addresses not allowed: %s", host)
	}

	return nil
}

// isPrivateIP checks if a host is a private IP
func isPrivateIP(host string) bool {
	privatePatterns := []string{
		`^10\.`,                         // 10.0.0.0/8
		`^172\.(1[6-9]|2[0-9]|3[01])\.`, // 172.16.0.0/12
		`^192\.168\.`,                   // 192.168.0.0/16
		`^169\.254\.`,                   // 169.254.0.0/16 (link-local)
		`^fc00:`,                        // IPv6 private
		`^fe80:`,                        // IPv6 link-local
	}

	for _, pattern := range privatePatterns {
		if matched, _ := regexp.MatchString(pattern, host); matched {
			return true
		}
	}

	return false
}

// ValidateHLSManifest validates the structure and content of an HLS manifest
func ValidateHLSManifest(manifest *HLSManifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest cannot be nil")
	}

	// Validate version
	if manifest.Version < 1 || manifest.Version > 10 {
		return fmt.Errorf("invalid HLS version: %d (must be 1-10)", manifest.Version)
	}

	// Validate manifest type
	validTypes := []HLSManifestType{HLSTypeMaster, HLSTypeMedia}
	validType := false
	for _, valid := range validTypes {
		if manifest.Type == valid {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid manifest type: %s", manifest.Type)
	}

	// Validate target duration for media playlists
	if manifest.Type == HLSTypeMedia {
		if manifest.TargetDuration <= 0 {
			return fmt.Errorf("target duration must be positive for media playlists")
		}

		if manifest.TargetDuration > 3600 { // 1 hour max
			return fmt.Errorf("target duration too large: %f seconds (max 3600)", manifest.TargetDuration)
		}
	}

	// Validate media sequence
	if manifest.MediaSequence < 0 {
		return fmt.Errorf("media sequence cannot be negative")
	}

	// Validate discontinuity sequence
	if manifest.DiscontinuitySequence < 0 {
		return fmt.Errorf("discontinuity sequence cannot be negative")
	}

	// Validate variants for master playlists
	if manifest.Type == HLSTypeMaster {
		if len(manifest.Variants) == 0 {
			return fmt.Errorf("master playlist must have at least one variant")
		}

		for i, variant := range manifest.Variants {
			if err := validateVariant(variant); err != nil {
				return fmt.Errorf("invalid variant %d: %w", i, err)
			}
		}
	}

	// Validate segments for media playlists
	if manifest.Type == HLSTypeMedia {
		for i, segment := range manifest.Segments {
			if err := validateSegment(segment); err != nil {
				return fmt.Errorf("invalid segment %d: %w", i, err)
			}
		}
	}

	return nil
}

// validateVariant validates an HLS variant stream
func validateVariant(variant *HLSVariant) error {
	if variant == nil {
		return fmt.Errorf("variant cannot be nil")
	}

	// Validate bandwidth
	if variant.Bandwidth <= 0 {
		return fmt.Errorf("bandwidth must be positive")
	}

	if variant.Bandwidth > 1000000000 { // 1 Gbps max
		return fmt.Errorf("bandwidth too large: %d bps (max 1 Gbps)", variant.Bandwidth)
	}

	// Validate URI
	if strings.TrimSpace(variant.URI) == "" {
		return fmt.Errorf("variant URI cannot be empty")
	}

	// Validate codecs if provided
	if len(variant.Codecs) > 0 {
		for _, codec := range variant.Codecs {
			if err := validateCodecs(codec); err != nil {
				return fmt.Errorf("invalid codec %s: %w", codec, err)
			}
		}
	}

	// Validate resolution if provided
	if variant.Resolution != nil {
		resolutionStr := fmt.Sprintf("%dx%d", variant.Resolution.Width, variant.Resolution.Height)
		if err := validateResolution(resolutionStr); err != nil {
			return fmt.Errorf("invalid resolution: %w", err)
		}
	}

	// Validate frame rate if provided
	if variant.FrameRate != nil && *variant.FrameRate > 0 {
		if *variant.FrameRate > 120 {
			return fmt.Errorf("frame rate too high: %f fps (max 120)", *variant.FrameRate)
		}
	}

	return nil
}

// validateSegment validates an HLS segment
func validateSegment(segment *HLSSegment) error {
	if segment == nil {
		return fmt.Errorf("segment cannot be nil")
	}

	// Validate duration
	if segment.Duration <= 0 {
		return fmt.Errorf("segment duration must be positive")
	}

	if segment.Duration > 3600 { // 1 hour max
		return fmt.Errorf("segment duration too large: %f seconds (max 3600)", segment.Duration)
	}

	// Validate URI
	if strings.TrimSpace(segment.URI) == "" {
		return fmt.Errorf("segment URI cannot be empty")
	}

	// Validate sequence number
	if segment.Sequence < 0 {
		return fmt.Errorf("sequence number cannot be negative")
	}

	return nil
}

// validateCodecs validates codec string format
func validateCodecs(codecs string) error {
	if strings.TrimSpace(codecs) == "" {
		return fmt.Errorf("codecs string cannot be empty")
	}

	// Basic validation for common codec formats
	validCodecPatterns := []string{
		`^avc1\.`, // H.264
		`^hev1\.`, // H.265
		`^mp4a\.`, // AAC
		`^opus`,   // Opus
		`^vp9`,    // VP9
		`^av01\.`, // AV1
	}

	codecParts := strings.Split(codecs, ",")
	for _, codec := range codecParts {
		codec = strings.TrimSpace(codec)
		validCodec := false

		for _, pattern := range validCodecPatterns {
			if matched, _ := regexp.MatchString(pattern, codec); matched {
				validCodec = true
				break
			}
		}

		if !validCodec {
			return fmt.Errorf("unsupported codec: %s", codec)
		}
	}

	return nil
}

// validateResolution validates resolution string format
func validateResolution(resolution string) error {
	if strings.TrimSpace(resolution) == "" {
		return fmt.Errorf("resolution cannot be empty")
	}

	// Check format: WIDTHxHEIGHT
	resolutionPattern := `^(\d+)x(\d+)$`
	matched, err := regexp.MatchString(resolutionPattern, resolution)
	if err != nil {
		return fmt.Errorf("error validating resolution: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid resolution format (expected WIDTHxHEIGHT): %s", resolution)
	}

	// Extract width and height
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return fmt.Errorf("invalid resolution format: %s", resolution)
	}

	// Validate reasonable dimensions
	width := parseInt(parts[0])
	height := parseInt(parts[1])

	if width <= 0 || height <= 0 {
		return fmt.Errorf("resolution dimensions must be positive: %dx%d", width, height)
	}

	if width > 7680 || height > 4320 { // 8K max
		return fmt.Errorf("resolution too large: %dx%d (max 7680x4320)", width, height)
	}

	return nil
}

// ValidateHLSAnalysisResult validates an HLS analysis result
func ValidateHLSAnalysisResult(result *HLSAnalysisResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	// Validate status
	validStatuses := []HLSStatus{
		HLSStatusPending,
		HLSStatusProcessing,
		HLSStatusCompleted,
		HLSStatusFailed,
	}

	validStatus := false
	for _, status := range validStatuses {
		if result.Status == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		return fmt.Errorf("invalid HLS status: %s", result.Status)
	}

	// Validate analysis if present
	if result.Analysis != nil {
		if err := ValidateHLSAnalysis(result.Analysis); err != nil {
			return fmt.Errorf("invalid HLS analysis: %w", err)
		}
	}

	return nil
}

// ValidateHLSAnalysis validates an HLS analysis
func ValidateHLSAnalysis(analysis *HLSAnalysis) error {
	if analysis == nil {
		return fmt.Errorf("analysis cannot be nil")
	}

	// Validate manifest
	if analysis.Manifest != nil {
		if err := ValidateHLSManifest(analysis.Manifest); err != nil {
			return fmt.Errorf("invalid manifest: %w", err)
		}
	}

	// Validate processing time
	if analysis.ProcessingTime < 0 {
		return fmt.Errorf("processing time cannot be negative")
	}

	return nil
}

// Helper function to parse integer
func parseInt(s string) int {
	val := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			val = val*10 + int(char-'0')
		}
	}
	return val
}
