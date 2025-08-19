package hls

import (
	"time"

	"github.com/google/uuid"
)

// HLSAnalysis represents a complete HLS analysis result
type HLSAnalysis struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	AnalysisID         uuid.UUID              `json:"analysis_id" db:"analysis_id"`
	ManifestURL        string                 `json:"manifest_url" db:"manifest_url"`
	ManifestType       HLSManifestType        `json:"manifest_type" db:"manifest_type"`
	Manifest           *HLSManifest           `json:"manifest,omitempty"`
	MasterPlaylist     *HLSMasterPlaylist     `json:"master_playlist,omitempty" db:"master_playlist"`
	MediaPlaylist      *HLSMediaPlaylist      `json:"media_playlist,omitempty" db:"media_playlist"`
	Variants           []*HLSVariant          `json:"variants,omitempty"`
	Segments           []*HLSSegment          `json:"segments,omitempty"`
	QualityLadder      *HLSQualityLadder      `json:"quality_ladder,omitempty"`
	ValidationResults  *HLSValidationResults  `json:"validation_results,omitempty"`
	PerformanceMetrics *HLSPerformanceMetrics `json:"performance_metrics,omitempty"`
	ProcessingTime     time.Duration          `json:"processing_time" db:"processing_time"`
	Status             HLSAnalysisStatus      `json:"status" db:"status"`
	ErrorMessage       string                 `json:"error_message,omitempty" db:"error_message"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// HLSAnalysisStatus represents the status of HLS analysis
type HLSAnalysisStatus string

const (
	HLSStatusPending    HLSAnalysisStatus = "pending"
	HLSStatusProcessing HLSAnalysisStatus = "processing"
	HLSStatusCompleted  HLSAnalysisStatus = "completed"
	HLSStatusFailed     HLSAnalysisStatus = "failed"
	HLSStatusCancelled  HLSAnalysisStatus = "cancelled"
)

// HLSStatus is an alias for HLSAnalysisStatus for backward compatibility
type HLSStatus = HLSAnalysisStatus

// HLSManifestType represents the type of HLS manifest
type HLSManifestType string

const (
	ManifestTypeMaster HLSManifestType = "master"
	ManifestTypeMedia  HLSManifestType = "media"
)

// Type aliases for backward compatibility
type HLSManifest struct {
	Type                  HLSManifestType    `json:"type"`
	Version               int                `json:"version"`
	TargetDuration        float64            `json:"target_duration,omitempty"`
	MediaSequence         int                `json:"media_sequence,omitempty"`
	DiscontinuitySequence int                `json:"discontinuity_sequence,omitempty"`
	Variants              []*HLSVariant      `json:"variants,omitempty"`
	Segments              []*HLSSegment      `json:"segments,omitempty"`
	MasterPlaylist        *HLSMasterPlaylist `json:"master_playlist,omitempty"`
	MediaPlaylist         *HLSMediaPlaylist  `json:"media_playlist,omitempty"`
}

const (
	HLSTypeMaster = ManifestTypeMaster
	HLSTypeMedia  = ManifestTypeMedia
)

// HLSMasterPlaylist represents a master playlist (m3u8)
type HLSMasterPlaylist struct {
	Version            int                     `json:"version"`
	Variants           []*HLSVariant           `json:"variants"`
	AudioRenditions    []*HLSAudioRendition    `json:"audio_renditions,omitempty"`
	VideoRenditions    []*HLSVideoRendition    `json:"video_renditions,omitempty"`
	SubtitleRenditions []*HLSSubtitleRendition `json:"subtitle_renditions,omitempty"`
	IFramePlaylists    []*HLSIFramePlaylist    `json:"iframe_playlists,omitempty"`
	SessionData        []*HLSSessionData       `json:"session_data,omitempty"`
	SessionKey         *HLSSessionKey          `json:"session_key,omitempty"`
}

// HLSMediaPlaylist represents a media playlist
type HLSMediaPlaylist struct {
	Version             int           `json:"version"`
	TargetDuration      float64       `json:"target_duration"`
	MediaSequence       int           `json:"media_sequence"`
	Segments            []*HLSSegment `json:"segments"`
	EndList             bool          `json:"end_list"`
	PlaylistType        string        `json:"playlist_type,omitempty"`
	AllowCache          *bool         `json:"allow_cache,omitempty"`
	IFramesOnly         bool          `json:"iframes_only"`
	IndependentSegments bool          `json:"independent_segments"`
	TotalDuration       float64       `json:"total_duration"`
	Key                 *HLSKey       `json:"key,omitempty"`
}

// HLSVariant represents a variant stream in master playlist
type HLSVariant struct {
	ID               uuid.UUID         `json:"id" db:"id"`
	AnalysisID       uuid.UUID         `json:"analysis_id" db:"analysis_id"`
	URI              string            `json:"uri" db:"uri"`
	Bandwidth        int               `json:"bandwidth" db:"bandwidth"`
	AverageBandwidth int               `json:"average_bandwidth,omitempty" db:"average_bandwidth"`
	Resolution       *HLSResolution    `json:"resolution,omitempty"`
	FrameRate        *float64          `json:"frame_rate,omitempty" db:"frame_rate"`
	Codecs           []string          `json:"codecs,omitempty" db:"codecs"`
	Audio            string            `json:"audio,omitempty" db:"audio"`
	Video            string            `json:"video,omitempty" db:"video"`
	Subtitles        string            `json:"subtitles,omitempty" db:"subtitles"`
	ClosedCaptions   string            `json:"closed_captions,omitempty" db:"closed_captions"`
	HDCPLevel        string            `json:"hdcp_level,omitempty" db:"hdcp_level"`
	VideoRange       string            `json:"video_range,omitempty" db:"video_range"`
	StableVariantID  string            `json:"stable_variant_id,omitempty" db:"stable_variant_id"`
	MediaPlaylist    *HLSMediaPlaylist `json:"media_playlist,omitempty"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
}

// HLSSegment represents a media segment
type HLSSegment struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	VariantID       uuid.UUID     `json:"variant_id" db:"variant_id"`
	URI             string        `json:"uri" db:"uri"`
	Duration        float64       `json:"duration" db:"duration"`
	Title           string        `json:"title,omitempty" db:"title"`
	ByteRange       *HLSByteRange `json:"byte_range,omitempty"`
	Discontinuity   bool          `json:"discontinuity" db:"discontinuity"`
	Key             *HLSKey       `json:"key,omitempty"`
	Map             *HLSMap       `json:"map,omitempty"`
	ProgramDateTime *time.Time    `json:"program_date_time,omitempty" db:"program_date_time"`
	DateRange       *HLSDateRange `json:"date_range,omitempty"`
	Gap             bool          `json:"gap" db:"gap"`
	Sequence        int           `json:"sequence" db:"sequence"`
	FileSize        int64         `json:"file_size,omitempty" db:"file_size"`
	Bitrate         int           `json:"bitrate,omitempty" db:"bitrate"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
}

// HLSResolution represents video resolution
type HLSResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// HLSByteRange represents byte range for segment
type HLSByteRange struct {
	Length int `json:"length"`
	Start  int `json:"start"`
}

// HLSKey represents encryption key information
type HLSKey struct {
	Method            string `json:"method"`
	URI               string `json:"uri,omitempty"`
	IV                string `json:"iv,omitempty"`
	KeyFormat         string `json:"key_format,omitempty"`
	KeyFormatVersions string `json:"key_format_versions,omitempty"`
}

// HLSMap represents initialization section
type HLSMap struct {
	URI       string        `json:"uri"`
	ByteRange *HLSByteRange `json:"byte_range,omitempty"`
}

// HLSDateRange represents date range information
type HLSDateRange struct {
	ID              string            `json:"id"`
	Class           string            `json:"class,omitempty"`
	StartDate       *time.Time        `json:"start_date,omitempty"`
	EndDate         *time.Time        `json:"end_date,omitempty"`
	Duration        *float64          `json:"duration,omitempty"`
	PlannedDuration *float64          `json:"planned_duration,omitempty"`
	Attributes      map[string]string `json:"attributes,omitempty"`
}

// HLSAudioRendition represents audio rendition in master playlist
type HLSAudioRendition struct {
	Type            string `json:"type"`
	GroupID         string `json:"group_id"`
	Name            string `json:"name"`
	Default         bool   `json:"default"`
	AutoSelect      bool   `json:"auto_select"`
	Forced          bool   `json:"forced"`
	URI             string `json:"uri,omitempty"`
	Language        string `json:"language,omitempty"`
	Channels        string `json:"channels,omitempty"`
	Characteristics string `json:"characteristics,omitempty"`
}

// HLSVideoRendition represents video rendition in master playlist
type HLSVideoRendition struct {
	Type            string `json:"type"`
	GroupID         string `json:"group_id"`
	Name            string `json:"name"`
	Default         bool   `json:"default"`
	AutoSelect      bool   `json:"auto_select"`
	URI             string `json:"uri,omitempty"`
	Language        string `json:"language,omitempty"`
	Characteristics string `json:"characteristics,omitempty"`
}

// HLSSubtitleRendition represents subtitle rendition in master playlist
type HLSSubtitleRendition struct {
	Type            string `json:"type"`
	GroupID         string `json:"group_id"`
	Name            string `json:"name"`
	Default         bool   `json:"default"`
	AutoSelect      bool   `json:"auto_select"`
	Forced          bool   `json:"forced"`
	URI             string `json:"uri,omitempty"`
	Language        string `json:"language,omitempty"`
	Characteristics string `json:"characteristics,omitempty"`
}

// HLSIFramePlaylist represents I-frame playlist
type HLSIFramePlaylist struct {
	URI        string         `json:"uri"`
	Bandwidth  int            `json:"bandwidth"`
	Resolution *HLSResolution `json:"resolution,omitempty"`
	Codecs     []string       `json:"codecs,omitempty"`
	VideoRange string         `json:"video_range,omitempty"`
	HDCPLevel  string         `json:"hdcp_level,omitempty"`
}

// HLSSessionData represents session data
type HLSSessionData struct {
	DataID   string `json:"data_id"`
	Value    string `json:"value,omitempty"`
	URI      string `json:"uri,omitempty"`
	Language string `json:"language,omitempty"`
}

// HLSSessionKey represents session key
type HLSSessionKey struct {
	Method            string `json:"method"`
	URI               string `json:"uri,omitempty"`
	IV                string `json:"iv,omitempty"`
	KeyFormat         string `json:"key_format,omitempty"`
	KeyFormatVersions string `json:"key_format_versions,omitempty"`
}

// HLSQualityLadder represents the quality ladder analysis
type HLSQualityLadder struct {
	VariantCount        int                 `json:"variant_count"`
	BitrateRange        *HLSBitrateRange    `json:"bitrate_range"`
	ResolutionRange     *HLSResolutionRange `json:"resolution_range"`
	FrameRateRange      *HLSFrameRateRange  `json:"frame_rate_range"`
	CodecDistribution   map[string]int      `json:"codec_distribution"`
	BitrateDistribution []*HLSBitratePoint  `json:"bitrate_distribution"`
	QualityGaps         []*HLSQualityGap    `json:"quality_gaps"`
	Recommendations     []string            `json:"recommendations"`
}

// HLSBitrateRange represents bitrate range
type HLSBitrateRange struct {
	Min     int     `json:"min"`
	Max     int     `json:"max"`
	Average float64 `json:"average"`
}

// HLSResolutionRange represents resolution range
type HLSResolutionRange struct {
	MinWidth  int `json:"min_width"`
	MaxWidth  int `json:"max_width"`
	MinHeight int `json:"min_height"`
	MaxHeight int `json:"max_height"`
}

// HLSFrameRateRange represents frame rate range
type HLSFrameRateRange struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
}

// HLSBitratePoint represents a bitrate point in distribution
type HLSBitratePoint struct {
	Bitrate    int            `json:"bitrate"`
	Resolution *HLSResolution `json:"resolution"`
	FrameRate  *float64       `json:"frame_rate"`
	Codecs     []string       `json:"codecs"`
}

// HLSQualityGap represents a detected quality gap
type HLSQualityGap struct {
	Type           string           `json:"type"`
	Severity       string           `json:"severity"`
	Description    string           `json:"description"`
	LowerVariant   *HLSBitratePoint `json:"lower_variant"`
	UpperVariant   *HLSBitratePoint `json:"upper_variant"`
	GapSize        float64          `json:"gap_size"`
	Recommendation string           `json:"recommendation"`
}

// HLSValidationResults represents HLS validation results
type HLSValidationResults struct {
	IsValid    bool                    `json:"is_valid"`
	Errors     []*HLSValidationError   `json:"errors,omitempty"`
	Warnings   []*HLSValidationWarning `json:"warnings,omitempty"`
	Compliance *HLSComplianceCheck     `json:"compliance,omitempty"`
	Summary    string                  `json:"summary"`
}

// HLSValidationError represents validation error
type HLSValidationError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	LineNumber int    `json:"line_number,omitempty"`
	FieldName  string `json:"field_name,omitempty"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// HLSValidationWarning represents validation warning
type HLSValidationWarning struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	LineNumber int    `json:"line_number,omitempty"`
	FieldName  string `json:"field_name,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// HLSComplianceCheck represents compliance check results
type HLSComplianceCheck struct {
	HLSVersion       string                `json:"hls_version"`
	RFC8216Compliant bool                  `json:"rfc8216_compliant"`
	AppleCompliant   bool                  `json:"apple_compliant"`
	AndroidCompliant bool                  `json:"android_compliant"`
	WebCompliant     bool                  `json:"web_compliant"`
	Issues           []*HLSComplianceIssue `json:"issues,omitempty"`
}

// HLSComplianceIssue represents compliance issue
type HLSComplianceIssue struct {
	Platform    string `json:"platform"`
	Issue       string `json:"issue"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Fix         string `json:"fix,omitempty"`
}

// HLSPerformanceMetrics represents performance metrics
type HLSPerformanceMetrics struct {
	SegmentCount            int                  `json:"segment_count"`
	TotalDuration           float64              `json:"total_duration"`
	AverageSegmentDuration  float64              `json:"average_segment_duration"`
	TargetDuration          float64              `json:"target_duration"`
	SegmentDurationVariance float64              `json:"segment_duration_variance"`
	BitrateVariance         float64              `json:"bitrate_variance"`
	StartupMetrics          *HLSStartupMetrics   `json:"startup_metrics,omitempty"`
	BufferingMetrics        *HLSBufferingMetrics `json:"buffering_metrics,omitempty"`
	BandwidthMetrics        *HLSBandwidthMetrics `json:"bandwidth_metrics,omitempty"`
	QualityMetrics          *HLSQualityMetrics   `json:"quality_metrics,omitempty"`
}

// HLSStartupMetrics represents startup performance metrics
type HLSStartupMetrics struct {
	ManifestLoadTime  float64 `json:"manifest_load_time"`
	FirstSegmentTime  float64 `json:"first_segment_time"`
	PlaybackStartTime float64 `json:"playback_start_time"`
	TimeToFirstFrame  float64 `json:"time_to_first_frame"`
}

// HLSBufferingMetrics represents buffering metrics
type HLSBufferingMetrics struct {
	BufferingRatio      float64 `json:"buffering_ratio"`
	BufferingEvents     int     `json:"buffering_events"`
	AverageBufferDepth  float64 `json:"average_buffer_depth"`
	BufferUnderruns     int     `json:"buffer_underruns"`
	RebufferingDuration float64 `json:"rebuffering_duration"`
}

// HLSBandwidthMetrics represents bandwidth metrics
type HLSBandwidthMetrics struct {
	RequiredBandwidth    int     `json:"required_bandwidth"`
	AverageBandwidth     float64 `json:"average_bandwidth"`
	PeakBandwidth        int     `json:"peak_bandwidth"`
	BandwidthUtilization float64 `json:"bandwidth_utilization"`
	AdaptationEvents     int     `json:"adaptation_events"`
}

// HLSQualityMetrics represents quality metrics
type HLSQualityMetrics struct {
	AverageQuality   float64 `json:"average_quality"`
	QualitySwitches  int     `json:"quality_switches"`
	QualityStability float64 `json:"quality_stability"`
	UpshiftEvents    int     `json:"upshift_events"`
	DownshiftEvents  int     `json:"downshift_events"`
}

// HLSAnalysisRequest represents an HLS analysis request
type HLSAnalysisRequest struct {
	ManifestURL         string   `json:"manifest_url" binding:"required"`
	AnalyzeSegments     bool     `json:"analyze_segments,omitempty"`
	AnalyzeQuality      bool     `json:"analyze_quality,omitempty"`
	ValidateCompliance  bool     `json:"validate_compliance,omitempty"`
	PerformanceAnalysis bool     `json:"performance_analysis,omitempty"`
	IncludeMetrics      []string `json:"include_metrics,omitempty"`
	MaxSegments         int      `json:"max_segments,omitempty"`
	Timeout             int      `json:"timeout,omitempty"`
	Async               bool     `json:"async,omitempty"`
}

// HLSAnalysisResult represents the result of HLS analysis
type HLSAnalysisResult struct {
	ID             uuid.UUID         `json:"id"`
	Status         HLSAnalysisStatus `json:"status"`
	Analysis       *HLSAnalysis      `json:"analysis,omitempty"`
	ProcessingTime time.Duration     `json:"processing_time"`
	Message        string            `json:"message,omitempty"`
	Error          string            `json:"error,omitempty"`
}

// HLSBatchAnalysisRequest represents batch HLS analysis request
type HLSBatchAnalysisRequest struct {
	Requests []HLSAnalysisRequest `json:"requests" binding:"required"`
	Async    bool                 `json:"async,omitempty"`
	Parallel int                  `json:"parallel,omitempty"`
}

// HLSBatchAnalysisResult represents batch HLS analysis result
type HLSBatchAnalysisResult struct {
	BatchID        uuid.UUID            `json:"batch_id"`
	Status         string               `json:"status"`
	Total          int                  `json:"total"`
	Completed      int                  `json:"completed"`
	Failed         int                  `json:"failed"`
	Results        []*HLSAnalysisResult `json:"results"`
	Summary        *HLSBatchSummary     `json:"summary,omitempty"`
	ProcessingTime time.Duration        `json:"processing_time"`
}

// HLSBatchSummary represents batch analysis summary
type HLSBatchSummary struct {
	TotalStreams    int           `json:"total_streams"`
	ValidStreams    int           `json:"valid_streams"`
	InvalidStreams  int           `json:"invalid_streams"`
	AverageVariants float64       `json:"average_variants"`
	AverageSegments float64       `json:"average_segments"`
	AverageDuration float64       `json:"average_duration"`
	CommonIssues    []string      `json:"common_issues"`
	Recommendations []string      `json:"recommendations"`
	ProcessingTime  time.Duration `json:"processing_time"`
}
