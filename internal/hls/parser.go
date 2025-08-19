package hls

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// HLSParser handles parsing of HLS manifests
type HLSParser struct {
	logger zerolog.Logger
}

// NewHLSParser creates a new HLS parser
func NewHLSParser(logger zerolog.Logger) *HLSParser {
	return &HLSParser{
		logger: logger,
	}
}

// ParseManifest parses an HLS manifest from a reader
func (p *HLSParser) ParseManifest(reader io.Reader, baseURL string) (*HLSAnalysis, error) {
	scanner := bufio.NewScanner(reader)
	var lines []string

	// Read all lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading manifest: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty manifest")
	}

	// Check if it's a valid M3U8 file
	if !strings.HasPrefix(lines[0], "#EXTM3U") {
		return nil, fmt.Errorf("invalid M3U8 format: missing #EXTM3U header")
	}

	// Determine manifest type
	manifestType := p.determineManifestType(lines)

	analysis := &HLSAnalysis{
		ID:           uuid.New(),
		ManifestURL:  baseURL,
		ManifestType: manifestType,
		Status:       HLSStatusProcessing,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	var err error
	switch manifestType {
	case ManifestTypeMaster:
		analysis.MasterPlaylist, err = p.parseMasterPlaylist(lines, baseURL)
	case ManifestTypeMedia:
		analysis.MediaPlaylist, err = p.parseMediaPlaylist(lines, baseURL)
	default:
		return nil, fmt.Errorf("unknown manifest type")
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing manifest: %w", err)
	}

	return analysis, nil
}

// determineManifestType determines if this is a master or media playlist
func (p *HLSParser) determineManifestType(lines []string) HLSManifestType {
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			return ManifestTypeMaster
		}
		if strings.HasPrefix(line, "#EXT-X-TARGETDURATION") {
			return ManifestTypeMedia
		}
	}
	return ManifestTypeMedia // Default to media if unclear
}

// parseMasterPlaylist parses a master playlist
func (p *HLSParser) parseMasterPlaylist(lines []string, baseURL string) (*HLSMasterPlaylist, error) {
	playlist := &HLSMasterPlaylist{
		Version:  3, // Default HLS version
		Variants: make([]*HLSVariant, 0),
	}

	i := 0
	for i < len(lines) {
		line := lines[i]

		switch {
		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			version, err := p.parseIntValue(line)
			if err == nil {
				playlist.Version = version
			}

		case strings.HasPrefix(line, "#EXT-X-STREAM-INF:"):
			variant, consumed, err := p.parseVariant(lines, i, baseURL)
			if err != nil {
				p.logger.Warn().Err(err).Msg("Failed to parse variant")
			} else {
				playlist.Variants = append(playlist.Variants, variant)
			}
			i += consumed - 1 // -1 because loop will increment

		case strings.HasPrefix(line, "#EXT-X-MEDIA:"):
			// Parse media renditions (audio, video, subtitles)
			mediaType, rendition := p.parseMediaRendition(line)
			switch mediaType {
			case "AUDIO":
				if audioRendition, ok := rendition.(*HLSAudioRendition); ok {
					playlist.AudioRenditions = append(playlist.AudioRenditions, audioRendition)
				}
			case "VIDEO":
				if videoRendition, ok := rendition.(*HLSVideoRendition); ok {
					playlist.VideoRenditions = append(playlist.VideoRenditions, videoRendition)
				}
			case "SUBTITLES":
				if subtitleRendition, ok := rendition.(*HLSSubtitleRendition); ok {
					playlist.SubtitleRenditions = append(playlist.SubtitleRenditions, subtitleRendition)
				}
			}

		case strings.HasPrefix(line, "#EXT-X-I-FRAME-STREAM-INF:"):
			iframePlaylist := p.parseIFramePlaylist(line, baseURL)
			if iframePlaylist != nil {
				playlist.IFramePlaylists = append(playlist.IFramePlaylists, iframePlaylist)
			}

		case strings.HasPrefix(line, "#EXT-X-SESSION-DATA:"):
			sessionData := p.parseSessionData(line)
			if sessionData != nil {
				playlist.SessionData = append(playlist.SessionData, sessionData)
			}

		case strings.HasPrefix(line, "#EXT-X-SESSION-KEY:"):
			sessionKey := p.parseSessionKey(line)
			if sessionKey != nil {
				playlist.SessionKey = sessionKey
			}
		}

		i++
	}

	return playlist, nil
}

// parseMediaPlaylist parses a media playlist
func (p *HLSParser) parseMediaPlaylist(lines []string, baseURL string) (*HLSMediaPlaylist, error) {
	playlist := &HLSMediaPlaylist{
		Version:  3, // Default HLS version
		Segments: make([]*HLSSegment, 0),
	}

	var currentKey *HLSKey
	var currentMap *HLSMap
	sequence := 0

	i := 0
	for i < len(lines) {
		line := lines[i]

		switch {
		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			version, err := p.parseIntValue(line)
			if err == nil {
				playlist.Version = version
			}

		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			duration, err := p.parseFloatValue(line)
			if err == nil {
				playlist.TargetDuration = duration
			}

		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			sequence, err := p.parseIntValue(line)
			if err == nil {
				playlist.MediaSequence = sequence
			}

		case strings.HasPrefix(line, "#EXT-X-ENDLIST"):
			playlist.EndList = true

		case strings.HasPrefix(line, "#EXT-X-PLAYLIST-TYPE:"):
			playlist.PlaylistType = p.parseStringValue(line)

		case strings.HasPrefix(line, "#EXT-X-ALLOW-CACHE:"):
			allowCache := strings.ToUpper(p.parseStringValue(line)) == "YES"
			playlist.AllowCache = &allowCache

		case strings.HasPrefix(line, "#EXT-X-I-FRAMES-ONLY"):
			playlist.IFramesOnly = true

		case strings.HasPrefix(line, "#EXT-X-INDEPENDENT-SEGMENTS"):
			playlist.IndependentSegments = true

		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			currentKey = p.parseKey(line)
			playlist.Key = currentKey

		case strings.HasPrefix(line, "#EXT-X-MAP:"):
			currentMap = p.parseMap(line, baseURL)

		case strings.HasPrefix(line, "#EXTINF:"):
			segment, consumed, err := p.parseSegment(lines, i, baseURL, currentKey, currentMap, sequence)
			if err != nil {
				p.logger.Warn().Err(err).Msg("Failed to parse segment")
			} else {
				playlist.Segments = append(playlist.Segments, segment)
				playlist.TotalDuration += segment.Duration
				sequence++
			}
			i += consumed - 1 // -1 because loop will increment
		}

		i++
	}

	return playlist, nil
}

// parseVariant parses a variant stream
func (p *HLSParser) parseVariant(lines []string, startIndex int, baseURL string) (*HLSVariant, int, error) {
	if startIndex >= len(lines) {
		return nil, 0, fmt.Errorf("invalid variant start index")
	}

	streamInfLine := lines[startIndex]
	if startIndex+1 >= len(lines) {
		return nil, 0, fmt.Errorf("missing variant URI")
	}

	uri := lines[startIndex+1]
	if strings.HasPrefix(uri, "#") {
		return nil, 0, fmt.Errorf("invalid variant URI")
	}

	variant := &HLSVariant{
		ID:        uuid.New(),
		URI:       p.resolveURL(uri, baseURL),
		CreatedAt: time.Now(),
	}

	// Parse EXT-X-STREAM-INF attributes
	attributes := p.parseAttributes(streamInfLine)

	if bandwidth, ok := attributes["BANDWIDTH"]; ok {
		if bw, err := strconv.Atoi(bandwidth); err == nil {
			variant.Bandwidth = bw
		}
	}

	if avgBandwidth, ok := attributes["AVERAGE-BANDWIDTH"]; ok {
		if avgBw, err := strconv.Atoi(avgBandwidth); err == nil {
			variant.AverageBandwidth = avgBw
		}
	}

	if resolution, ok := attributes["RESOLUTION"]; ok {
		if res := p.parseResolution(resolution); res != nil {
			variant.Resolution = res
		}
	}

	if frameRate, ok := attributes["FRAME-RATE"]; ok {
		if fr, err := strconv.ParseFloat(frameRate, 64); err == nil {
			variant.FrameRate = &fr
		}
	}

	if codecs, ok := attributes["CODECS"]; ok {
		variant.Codecs = p.parseCodecs(codecs)
	}

	if audio, ok := attributes["AUDIO"]; ok {
		variant.Audio = audio
	}

	if video, ok := attributes["VIDEO"]; ok {
		variant.Video = video
	}

	if subtitles, ok := attributes["SUBTITLES"]; ok {
		variant.Subtitles = subtitles
	}

	if cc, ok := attributes["CLOSED-CAPTIONS"]; ok {
		variant.ClosedCaptions = cc
	}

	if hdcp, ok := attributes["HDCP-LEVEL"]; ok {
		variant.HDCPLevel = hdcp
	}

	if videoRange, ok := attributes["VIDEO-RANGE"]; ok {
		variant.VideoRange = videoRange
	}

	if stableID, ok := attributes["STABLE-VARIANT-ID"]; ok {
		variant.StableVariantID = stableID
	}

	return variant, 2, nil // Consumed 2 lines (EXT-X-STREAM-INF + URI)
}

// parseSegment parses a media segment
func (p *HLSParser) parseSegment(lines []string, startIndex int, baseURL string, currentKey *HLSKey, currentMap *HLSMap, sequence int) (*HLSSegment, int, error) {
	if startIndex >= len(lines) {
		return nil, 0, fmt.Errorf("invalid segment start index")
	}

	extinfLine := lines[startIndex]
	if startIndex+1 >= len(lines) {
		return nil, 0, fmt.Errorf("missing segment URI")
	}

	uri := lines[startIndex+1]
	if strings.HasPrefix(uri, "#") {
		return nil, 0, fmt.Errorf("invalid segment URI")
	}

	segment := &HLSSegment{
		ID:        uuid.New(),
		URI:       p.resolveURL(uri, baseURL),
		Key:       currentKey,
		Map:       currentMap,
		Sequence:  sequence,
		CreatedAt: time.Now(),
	}

	// Parse EXTINF
	if err := p.parseExtinf(extinfLine, segment); err != nil {
		return nil, 0, fmt.Errorf("error parsing EXTINF: %w", err)
	}

	consumed := 2

	// Look for additional segment tags before this segment
	for i := startIndex - 1; i >= 0; i-- {
		line := lines[i]
		if strings.HasPrefix(line, "#EXT-X-DISCONTINUITY") {
			segment.Discontinuity = true
			break
		}
		if strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:") {
			if pdt := p.parseProgramDateTime(line); pdt != nil {
				segment.ProgramDateTime = pdt
			}
			break
		}
		if strings.HasPrefix(line, "#EXT-X-BYTERANGE:") {
			if br := p.parseByteRange(line); br != nil {
				segment.ByteRange = br
			}
			break
		}
		if strings.HasPrefix(line, "#EXT-X-GAP") {
			segment.Gap = true
			break
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			break // Found previous segment
		}
	}

	return segment, consumed, nil
}

// parseExtinf parses the EXTINF line
func (p *HLSParser) parseExtinf(line string, segment *HLSSegment) error {
	// Format: #EXTINF:duration,[title]
	content := strings.TrimPrefix(line, "#EXTINF:")
	parts := strings.SplitN(content, ",", 2)

	if len(parts) < 1 {
		return fmt.Errorf("invalid EXTINF format")
	}

	duration, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	segment.Duration = duration

	if len(parts) > 1 {
		segment.Title = parts[1]
	}

	return nil
}

// parseAttributes parses attribute string into map
func (p *HLSParser) parseAttributes(line string) map[string]string {
	attributes := make(map[string]string)

	// Find the attributes part (after the colon)
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		return attributes
	}

	attributesStr := line[colonIndex+1:]

	// Regular expression to match key=value pairs
	re := regexp.MustCompile(`([A-Z-]+)=([^,]+)`)
	matches := re.FindAllStringSubmatch(attributesStr, -1)

	for _, match := range matches {
		if len(match) == 3 {
			key := match[1]
			value := strings.Trim(match[2], `"`)
			attributes[key] = value
		}
	}

	return attributes
}

// parseResolution parses resolution string
func (p *HLSParser) parseResolution(resolution string) *HLSResolution {
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return nil
	}

	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return nil
	}

	return &HLSResolution{
		Width:  width,
		Height: height,
	}
}

// parseCodecs parses codecs string
func (p *HLSParser) parseCodecs(codecs string) []string {
	// Remove quotes if present
	codecs = strings.Trim(codecs, `"`)
	return strings.Split(codecs, ",")
}

// parseKey parses EXT-X-KEY tag
func (p *HLSParser) parseKey(line string) *HLSKey {
	attributes := p.parseAttributes(line)

	key := &HLSKey{}

	if method, ok := attributes["METHOD"]; ok {
		key.Method = method
	}

	if uri, ok := attributes["URI"]; ok {
		key.URI = strings.Trim(uri, `"`)
	}

	if iv, ok := attributes["IV"]; ok {
		key.IV = iv
	}

	if keyFormat, ok := attributes["KEYFORMAT"]; ok {
		key.KeyFormat = keyFormat
	}

	if keyFormatVersions, ok := attributes["KEYFORMATVERSIONS"]; ok {
		key.KeyFormatVersions = keyFormatVersions
	}

	return key
}

// parseMap parses EXT-X-MAP tag
func (p *HLSParser) parseMap(line string, baseURL string) *HLSMap {
	attributes := p.parseAttributes(line)

	mapInfo := &HLSMap{}

	if uri, ok := attributes["URI"]; ok {
		mapInfo.URI = p.resolveURL(strings.Trim(uri, `"`), baseURL)
	}

	if byteRange, ok := attributes["BYTERANGE"]; ok {
		if br := p.parseByteRangeString(byteRange); br != nil {
			mapInfo.ByteRange = br
		}
	}

	return mapInfo
}

// parseByteRange parses EXT-X-BYTERANGE tag
func (p *HLSParser) parseByteRange(line string) *HLSByteRange {
	content := strings.TrimPrefix(line, "#EXT-X-BYTERANGE:")
	return p.parseByteRangeString(content)
}

// parseByteRangeString parses byte range string
func (p *HLSParser) parseByteRangeString(byteRange string) *HLSByteRange {
	parts := strings.Split(byteRange, "@")
	if len(parts) < 1 {
		return nil
	}

	length, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}

	br := &HLSByteRange{Length: length}

	if len(parts) > 1 {
		start, err := strconv.Atoi(parts[1])
		if err == nil {
			br.Start = start
		}
	}

	return br
}

// parseProgramDateTime parses EXT-X-PROGRAM-DATE-TIME tag
func (p *HLSParser) parseProgramDateTime(line string) *time.Time {
	content := strings.TrimPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:")

	// Parse ISO 8601 format
	pdt, err := time.Parse(time.RFC3339, content)
	if err != nil {
		// Try alternative format
		pdt, err = time.Parse("2006-01-02T15:04:05.000Z", content)
		if err != nil {
			return nil
		}
	}

	return &pdt
}

// parseMediaRendition parses EXT-X-MEDIA tag
func (p *HLSParser) parseMediaRendition(line string) (string, interface{}) {
	attributes := p.parseAttributes(line)

	mediaType, ok := attributes["TYPE"]
	if !ok {
		return "", nil
	}

	switch mediaType {
	case "AUDIO":
		return mediaType, &HLSAudioRendition{
			Type:            mediaType,
			GroupID:         attributes["GROUP-ID"],
			Name:            attributes["NAME"],
			Default:         attributes["DEFAULT"] == "YES",
			AutoSelect:      attributes["AUTOSELECT"] == "YES",
			Forced:          attributes["FORCED"] == "YES",
			URI:             attributes["URI"],
			Language:        attributes["LANGUAGE"],
			Channels:        attributes["CHANNELS"],
			Characteristics: attributes["CHARACTERISTICS"],
		}
	case "VIDEO":
		return mediaType, &HLSVideoRendition{
			Type:            mediaType,
			GroupID:         attributes["GROUP-ID"],
			Name:            attributes["NAME"],
			Default:         attributes["DEFAULT"] == "YES",
			AutoSelect:      attributes["AUTOSELECT"] == "YES",
			URI:             attributes["URI"],
			Language:        attributes["LANGUAGE"],
			Characteristics: attributes["CHARACTERISTICS"],
		}
	case "SUBTITLES":
		return mediaType, &HLSSubtitleRendition{
			Type:            mediaType,
			GroupID:         attributes["GROUP-ID"],
			Name:            attributes["NAME"],
			Default:         attributes["DEFAULT"] == "YES",
			AutoSelect:      attributes["AUTOSELECT"] == "YES",
			Forced:          attributes["FORCED"] == "YES",
			URI:             attributes["URI"],
			Language:        attributes["LANGUAGE"],
			Characteristics: attributes["CHARACTERISTICS"],
		}
	}

	return "", nil
}

// parseIFramePlaylist parses EXT-X-I-FRAME-STREAM-INF tag
func (p *HLSParser) parseIFramePlaylist(line string, baseURL string) *HLSIFramePlaylist {
	attributes := p.parseAttributes(line)

	iframe := &HLSIFramePlaylist{}

	if uri, ok := attributes["URI"]; ok {
		iframe.URI = p.resolveURL(strings.Trim(uri, `"`), baseURL)
	}

	if bandwidth, ok := attributes["BANDWIDTH"]; ok {
		if bw, err := strconv.Atoi(bandwidth); err == nil {
			iframe.Bandwidth = bw
		}
	}

	if resolution, ok := attributes["RESOLUTION"]; ok {
		if res := p.parseResolution(resolution); res != nil {
			iframe.Resolution = res
		}
	}

	if codecs, ok := attributes["CODECS"]; ok {
		iframe.Codecs = p.parseCodecs(codecs)
	}

	if videoRange, ok := attributes["VIDEO-RANGE"]; ok {
		iframe.VideoRange = videoRange
	}

	if hdcp, ok := attributes["HDCP-LEVEL"]; ok {
		iframe.HDCPLevel = hdcp
	}

	return iframe
}

// parseSessionData parses EXT-X-SESSION-DATA tag
func (p *HLSParser) parseSessionData(line string) *HLSSessionData {
	attributes := p.parseAttributes(line)

	sessionData := &HLSSessionData{}

	if dataID, ok := attributes["DATA-ID"]; ok {
		sessionData.DataID = dataID
	}

	if value, ok := attributes["VALUE"]; ok {
		sessionData.Value = value
	}

	if uri, ok := attributes["URI"]; ok {
		sessionData.URI = uri
	}

	if language, ok := attributes["LANGUAGE"]; ok {
		sessionData.Language = language
	}

	return sessionData
}

// parseSessionKey parses EXT-X-SESSION-KEY tag
func (p *HLSParser) parseSessionKey(line string) *HLSSessionKey {
	attributes := p.parseAttributes(line)

	sessionKey := &HLSSessionKey{}

	if method, ok := attributes["METHOD"]; ok {
		sessionKey.Method = method
	}

	if uri, ok := attributes["URI"]; ok {
		sessionKey.URI = strings.Trim(uri, `"`)
	}

	if iv, ok := attributes["IV"]; ok {
		sessionKey.IV = iv
	}

	if keyFormat, ok := attributes["KEYFORMAT"]; ok {
		sessionKey.KeyFormat = keyFormat
	}

	if keyFormatVersions, ok := attributes["KEYFORMATVERSIONS"]; ok {
		sessionKey.KeyFormatVersions = keyFormatVersions
	}

	return sessionKey
}

// resolveURL resolves relative URLs against base URL
func (p *HLSParser) resolveURL(uri, baseURL string) string {
	if baseURL == "" {
		return uri
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	resolved := base.ResolveReference(u)
	return resolved.String()
}

// parseIntValue parses integer value from tag
func (p *HLSParser) parseIntValue(line string) (int, error) {
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		return 0, fmt.Errorf("no colon found")
	}

	valueStr := line[colonIndex+1:]
	return strconv.Atoi(valueStr)
}

// parseFloatValue parses float value from tag
func (p *HLSParser) parseFloatValue(line string) (float64, error) {
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		return 0, fmt.Errorf("no colon found")
	}

	valueStr := line[colonIndex+1:]
	return strconv.ParseFloat(valueStr, 64)
}

// parseStringValue parses string value from tag
func (p *HLSParser) parseStringValue(line string) string {
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		return ""
	}

	return line[colonIndex+1:]
}
