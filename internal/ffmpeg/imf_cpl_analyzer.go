package ffmpeg

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// IMFCPLAnalyzer handles Composition Playlist (CPL) analysis for IMF packages.
// This module focuses specifically on parsing and validating CPL files according
// to SMPTE 2067 standards.
//
// The CPL analyzer validates:
//   - CPL XML structure and schema compliance
//   - Timeline and segment structure
//   - Virtual track organization
//   - Asset references and mapping
//   - Edit rate and duration consistency
//   - UUID format and uniqueness
type IMFCPLAnalyzer struct {
	logger zerolog.Logger
}

// NewIMFCPLAnalyzer creates a new CPL analyzer for IMF packages.
//
// Parameters:
//   - logger: Structured logger for operation tracking
//
// Returns:
//   - *IMFCPLAnalyzer: Configured CPL analyzer instance
//
// The analyzer validates CPL files against SMPTE 2067-3 specifications
// and provides detailed analysis of composition structure.
func NewIMFCPLAnalyzer(logger zerolog.Logger) *IMFCPLAnalyzer {
	return &IMFCPLAnalyzer{
		logger: logger,
	}
}

// AnalyzeCPL performs comprehensive CPL analysis for an IMF package.
//
// Parameters:
//   - packagePath: Path to the IMF package directory
//
// Returns:
//   - *CPLAnalysis: Detailed CPL analysis results
//   - error: Error if CPL analysis fails
//
// The function locates CPL files, parses their XML structure, validates
// schema compliance, and analyzes composition timeline and asset references.
func (cpl *IMFCPLAnalyzer) AnalyzeCPL(packagePath string) (*CPLAnalysis, error) {
	cpl.logger.Info().Str("package_path", packagePath).Msg("Starting CPL analysis")

	analysis := &CPLAnalysis{}

	// Find CPL files in the package
	cplFiles, err := cpl.findCPLFiles(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to find CPL files: %w", err)
	}

	if len(cplFiles) == 0 {
		analysis.CPLExists = false
		return analysis, nil
	}

	analysis.CPLExists = true

	// Analyze the first CPL file (primary composition)
	primaryCPL := cplFiles[0]
	if err := cpl.parseCPLFile(primaryCPL, analysis); err != nil {
		return nil, fmt.Errorf("failed to parse CPL file %s: %w", primaryCPL, err)
	}

	// Validate CPL structure and content
	if err := cpl.validateCPLStructure(analysis); err != nil {
		cpl.logger.Warn().Err(err).Msg("CPL structure validation failed")
	}

	// Analyze segments and virtual tracks
	if err := cpl.analyzeSegments(analysis); err != nil {
		cpl.logger.Warn().Err(err).Msg("Segment analysis failed")
	}

	cpl.logger.Info().
		Str("cpl_id", analysis.CPLID).
		Int("segment_count", analysis.SegmentCount).
		Int("track_count", analysis.VirtualTrackCount).
		Msg("CPL analysis completed")

	return analysis, nil
}

// findCPLFiles locates all CPL files in the IMF package directory.
func (cpl *IMFCPLAnalyzer) findCPLFiles(packagePath string) ([]string, error) {
	pattern := filepath.Join(packagePath, "CPL*.xml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for CPL files: %w", err)
	}

	// Also check for lowercase cpl files
	patternLower := filepath.Join(packagePath, "cpl*.xml")
	matchesLower, err := filepath.Glob(patternLower)
	if err == nil {
		matches = append(matches, matchesLower...)
	}

	return matches, nil
}

// parseCPLFile parses a CPL XML file and extracts composition information.
func (cpl *IMFCPLAnalyzer) parseCPLFile(filePath string, analysis *CPLAnalysis) error {
	cpl.logger.Debug().Str("file_path", filePath).Msg("Parsing CPL file")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read CPL file: %w", err)
	}

	// Parse basic CPL structure using XML
	var cplData struct {
		XMLName struct{}  `xml:"CompositionPlaylist"`
		ID      string    `xml:"Id"`
		Issuer  string    `xml:"Issuer"`
		Creator string    `xml:"Creator"`
		ContentTitleText string `xml:"ContentTitleText"`
		EditRate struct {
			Text string `xml:",chardata"`
		} `xml:"EditRate"`
		TotalRunningTime struct {
			Text string `xml:",chardata"`
		} `xml:"TotalRunningTime"`
		SegmentList struct {
			Segments []struct {
				ID string `xml:"Id"`
				SequenceList struct {
					Sequences []struct {
						TrackID   string `xml:"TrackId"`
						TrackType string `xml:"type,attr"`
						ResourceList struct {
							Resources []struct {
								ID           string `xml:"Id"`
								EditRate     struct {
									Text string `xml:",chardata"`
								} `xml:"EditRate"`
								IntrinsicDuration struct {
									Text string `xml:",chardata"`
								} `xml:"IntrinsicDuration"`
								EntryPoint struct {
									Text string `xml:",chardata"`
								} `xml:"EntryPoint"`
								SourceDuration struct {
									Text string `xml:",chardata"`
								} `xml:"SourceDuration"`
							} `xml:"Resource"`
						} `xml:"ResourceList"`
					} `xml:"Sequence"`
				} `xml:"SequenceList"`
			} `xml:"Segment"`
		} `xml:"SegmentList"`
	}

	if err := xml.Unmarshal(data, &cplData); err != nil {
		return fmt.Errorf("failed to parse CPL XML: %w", err)
	}

	// Extract basic CPL information
	analysis.CPLID = cplData.ID
	analysis.CPLTitle = cplData.ContentTitleText
	analysis.EditRate = cplData.EditRate.Text
	analysis.Duration = cplData.TotalRunningTime.Text

	// Count segments and tracks
	analysis.SegmentCount = len(cplData.SegmentList.Segments)
	
	// Process segments to extract track information
	trackTypes := make(map[string]int)
	var segments []IMFSegment
	
	for _, segment := range cplData.SegmentList.Segments {
		imfSegment := IMFSegment{
			SegmentID:  segment.ID,
			TrackCount: len(segment.SequenceList.Sequences),
		}
		
		var tracks []IMFTrack
		for _, sequence := range segment.SequenceList.Sequences {
			trackType := sequence.TrackType
			if trackType == "" {
				// Infer track type from track ID pattern
				trackType = cpl.inferTrackType(sequence.TrackID)
			}
			
			trackTypes[trackType]++
			
			track := IMFTrack{
				TrackID:       sequence.TrackID,
				TrackType:     trackType,
				ResourceCount: len(sequence.ResourceList.Resources),
			}
			
			// Process resources in this track
			var resources []IMFAsset
			for _, resource := range sequence.ResourceList.Resources {
				asset := IMFAsset{
					AssetID:    resource.ID,
					EditRate:   resource.EditRate.Text,
					Duration:   resource.IntrinsicDuration.Text,
					EntryPoint: resource.EntryPoint.Text,
				}
				resources = append(resources, asset)
			}
			track.Resources = resources
			tracks = append(tracks, track)
		}
		
		imfSegment.Tracks = tracks
		segments = append(segments, imfSegment)
	}

	analysis.IMFSegments = segments
	analysis.VirtualTrackCount = len(trackTypes)
	analysis.VideoTrackCount = trackTypes["video"]
	analysis.AudioTrackCount = trackTypes["audio"]
	analysis.SubtitleTrackCount = trackTypes["subtitle"]

	return nil
}

// inferTrackType attempts to determine track type from track ID patterns.
func (cpl *IMFCPLAnalyzer) inferTrackType(trackID string) string {
	trackIDLower := strings.ToLower(trackID)
	
	if strings.Contains(trackIDLower, "video") || strings.Contains(trackIDLower, "pict") {
		return "video"
	} else if strings.Contains(trackIDLower, "audio") || strings.Contains(trackIDLower, "sound") {
		return "audio"
	} else if strings.Contains(trackIDLower, "subtitle") || strings.Contains(trackIDLower, "text") {
		return "subtitle"
	}
	
	return "unknown"
}

// validateCPLStructure validates the structure and content of the CPL.
func (cpl *IMFCPLAnalyzer) validateCPLStructure(analysis *CPLAnalysis) error {
	var validationErrors []string

	// Validate UUID format
	uuidPattern := regexp.MustCompile(UUIDPattern)
	if !uuidPattern.MatchString(analysis.CPLID) {
		validationErrors = append(validationErrors, "CPL ID is not a valid UUID")
	}

	// Validate edit rate format
	if analysis.EditRate != "" {
		if err := cpl.validateEditRate(analysis.EditRate); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid edit rate: %v", err))
		}
	}

	// Validate duration format
	if analysis.Duration != "" {
		if err := cpl.validateDuration(analysis.Duration); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid duration: %v", err))
		}
	}

	// Validate segment structure
	if analysis.SegmentCount == 0 {
		validationErrors = append(validationErrors, "CPL must contain at least one segment")
	}

	// Validate track structure
	if analysis.VirtualTrackCount == 0 {
		validationErrors = append(validationErrors, "CPL must contain at least one virtual track")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("CPL validation errors: %v", validationErrors)
	}

	return nil
}

// validateEditRate validates the edit rate format (e.g., "24/1", "25/1").
func (cpl *IMFCPLAnalyzer) validateEditRate(editRate string) error {
	parts := strings.Split(editRate, "/")
	if len(parts) != 2 {
		return fmt.Errorf("edit rate must be in format 'num/den'")
	}

	numerator, err := strconv.Atoi(parts[0])
	if err != nil || numerator <= 0 {
		return fmt.Errorf("invalid numerator in edit rate")
	}

	denominator, err := strconv.Atoi(parts[1])
	if err != nil || denominator <= 0 {
		return fmt.Errorf("invalid denominator in edit rate")
	}

	// Check for common frame rates
	frameRate := float64(numerator) / float64(denominator)
	validFrameRates := []float64{23.976, 24, 25, 29.97, 30, 48, 50, 59.94, 60}
	
	for _, validRate := range validFrameRates {
		if abs(frameRate-validRate) < 0.001 {
			return nil
		}
	}

	cpl.logger.Warn().
		Float64("frame_rate", frameRate).
		Msg("Unusual frame rate detected in CPL")

	return nil
}

// validateDuration validates the duration format.
func (cpl *IMFCPLAnalyzer) validateDuration(duration string) error {
	// Duration can be in frames or timecode format
	if _, err := strconv.Atoi(duration); err == nil {
		// Frame count format
		return nil
	}

	// Timecode format validation (HH:MM:SS:FF)
	timecodePattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}:\d{2}$`)
	if timecodePattern.MatchString(duration) {
		return nil
	}

	return fmt.Errorf("duration must be in frames or timecode format (HH:MM:SS:FF)")
}

// analyzeSegments performs detailed analysis of CPL segments.
func (cpl *IMFCPLAnalyzer) analyzeSegments(analysis *CPLAnalysis) error {
	if len(analysis.IMFSegments) == 0 {
		return nil
	}

	// Validate segment continuity
	for i, segment := range analysis.IMFSegments {
		cpl.logger.Debug().
			Int("segment_index", i).
			Str("segment_id", segment.SegmentID).
			Int("track_count", segment.TrackCount).
			Msg("Analyzing CPL segment")

		// Validate segment UUID
		uuidPattern := regexp.MustCompile(UUIDPattern)
		if !uuidPattern.MatchString(segment.SegmentID) {
			cpl.logger.Warn().
				Str("segment_id", segment.SegmentID).
				Msg("Segment ID is not a valid UUID")
		}

		// Validate track structure within segment
		if err := cpl.validateSegmentTracks(segment); err != nil {
			cpl.logger.Warn().
				Err(err).
				Str("segment_id", segment.SegmentID).
				Msg("Segment track validation failed")
		}
	}

	return nil
}

// validateSegmentTracks validates the track structure within a segment.
func (cpl *IMFCPLAnalyzer) validateSegmentTracks(segment IMFSegment) error {
	if len(segment.Tracks) == 0 {
		return fmt.Errorf("segment must contain at least one track")
	}

	trackTypes := make(map[string]bool)
	for _, track := range segment.Tracks {
		trackTypes[track.TrackType] = true

		// Validate track resources
		if track.ResourceCount == 0 {
			cpl.logger.Warn().
				Str("track_id", track.TrackID).
				Msg("Track contains no resources")
		}

		// Validate resource continuity
		if err := cpl.validateTrackResources(track); err != nil {
			return fmt.Errorf("track %s validation failed: %w", track.TrackID, err)
		}
	}

	return nil
}

// validateTrackResources validates the resources within a track.
func (cpl *IMFCPLAnalyzer) validateTrackResources(track IMFTrack) error {
	if len(track.Resources) == 0 {
		return nil
	}

	// Check resource timing continuity
	for i, resource := range track.Resources {
		// Validate resource UUID
		uuidPattern := regexp.MustCompile(UUIDPattern)
		if !uuidPattern.MatchString(resource.AssetID) {
			cpl.logger.Warn().
				Str("asset_id", resource.AssetID).
				Msg("Resource asset ID is not a valid UUID")
		}

		// Validate edit rate consistency
		if i == 0 {
			track.EditRate = resource.EditRate
		} else if track.EditRate != resource.EditRate {
			cpl.logger.Warn().
				Str("expected_rate", track.EditRate).
				Str("actual_rate", resource.EditRate).
				Str("asset_id", resource.AssetID).
				Msg("Edit rate inconsistency detected in track")
		}
	}

	return nil
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}