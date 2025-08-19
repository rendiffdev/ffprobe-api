package ffmpeg

// PSE (Photosensitive Epilepsy) analysis types and structures.
// This file contains all data structures used for photosensitive epilepsy risk analysis
// in compliance with international broadcast safety standards.

// PSEAnalysis contains comprehensive photosensitive epilepsy risk analysis results.
// This is the main container for all PSE-related analysis data, providing detailed
// assessment of potential seizure triggers in video content.
//
// The analysis follows international standards:
//   - ITU-R BT.1702: Guidelines for photosensitive epilepsy
//   - ITC Guidance: UK broadcast standards
//   - Ofcom Guidelines: UK regulatory requirements
//   - EBU Tech 3253: European broadcast safety standards
type PSEAnalysis struct {
	// Overall risk assessment
	PSERiskLevel            string                   `json:"pse_risk_level"`           // safe, low, medium, high, critical
	OverallRiskScore        float64                  `json:"overall_risk_score"`       // 0-100 comprehensive risk score
	MaxRiskTimestamp        float64                  `json:"max_risk_timestamp"`       // timestamp of highest risk
	RiskReason              string                   `json:"risk_reason"`              // human-readable explanation
	SafeForBroadcast        bool                     `json:"safe_for_broadcast"`       // ITU-R BT.1702 compliance
	RequiresWarning         bool                     `json:"requires_warning"`         // whether content needs PSE warning
	
	// Compliance with broadcast standards
	BroadcastCompliance     *BroadcastPSECompliance  `json:"broadcast_compliance,omitempty"`
	
	// Detailed analysis categories
	FlashAnalysis           *FlashAnalysis           `json:"flash_analysis,omitempty"`
	RedFlashAnalysis        *RedFlashAnalysis        `json:"red_flash_analysis,omitempty"`
	PatternAnalysis         *PatternAnalysis         `json:"pattern_analysis,omitempty"`
	LuminanceAnalysis       *LuminanceAnalysis       `json:"luminance_analysis,omitempty"`
	TemporalAnalysis        *TemporalPSEAnalysis     `json:"temporal_analysis,omitempty"`
	SpatialAnalysis         *SpatialPSEAnalysis      `json:"spatial_analysis,omitempty"`
	SceneAnalysis           *SceneAnalysis           `json:"scene_analysis,omitempty"`
	
	// Analysis metadata
	AnalysisDuration        float64                  `json:"analysis_duration"`        // seconds analyzed
	SamplingRate            float64                  `json:"sampling_rate"`            // frames per second analyzed
	StandardsVersion        string                   `json:"standards_version"`        // PSE standards version used
	AnalysisMethod          string                   `json:"analysis_method"`          // analysis algorithm version
	Confidence              float64                  `json:"confidence"`               // 0-1 confidence in results
}

// BroadcastPSECompliance contains compliance status for various broadcast standards.
// This structure tracks adherence to international and regional broadcast safety requirements.
type BroadcastPSECompliance struct {
	ITU_R_BT1702_Compliant  bool                     `json:"itu_r_bt1702_compliant"`   // ITU-R BT.1702 compliance
	ITC_Compliant           bool                     `json:"itc_compliant"`            // ITC guidelines compliance
	Ofcom_Compliant         bool                     `json:"ofcom_compliant"`          // Ofcom guidelines compliance
	EBU_Tech3253_Compliant  bool                     `json:"ebu_tech3253_compliant"`   // EBU Tech 3253 compliance
	FCC_Compliant           bool                     `json:"fcc_compliant"`            // FCC guidelines compliance
	ComplianceNotes         []string                 `json:"compliance_notes,omitempty"` // specific compliance issues
	LastUpdated             string                   `json:"last_updated"`             // standards version date
}

// FlashAnalysis analyzes general flash patterns that may trigger photosensitive seizures.
// This covers all types of flashing content including strobing effects and rapid scene changes.
type FlashAnalysis struct {
	TotalFlashes            int                      `json:"total_flashes"`            // total flash events detected
	FlashRate               float64                  `json:"flash_rate"`               // flashes per second (peak)
	AverageFlashRate        float64                  `json:"average_flash_rate"`       // average across content
	MaxFlashRate            float64                  `json:"max_flash_rate"`           // maximum rate in any second
	FlashDurations          []FlashDuration          `json:"flash_durations,omitempty"`
	FlashFrequencyBands     map[string]int           `json:"flash_frequency_bands"`    // frequency distribution
	DangerousFlashPeriods   []TimePeriod             `json:"dangerous_flash_periods,omitempty"`
	FlashIntensity          *FlashIntensity          `json:"flash_intensity,omitempty"`
	ExceedsFlashThreshold   bool                     `json:"exceeds_flash_threshold"`  // ITU-R threshold exceeded
}

// RedFlashAnalysis analyzes red flash patterns specifically.
// Red flashes are particularly dangerous for photosensitive individuals and have
// specific thresholds in broadcast safety standards.
type RedFlashAnalysis struct {
	RedFlashCount           int                      `json:"red_flash_count"`          // total red flash events
	RedFlashRate            float64                  `json:"red_flash_rate"`           // red flashes per second
	MaxRedFlashRate         float64                  `json:"max_red_flash_rate"`       // peak red flash rate
	RedFlashDurations       []FlashDuration          `json:"red_flash_durations,omitempty"`
	RedSaturationLevels     []float64                `json:"red_saturation_levels,omitempty"` // red saturation values
	DangerousRedPeriods     []TimePeriod             `json:"dangerous_red_periods,omitempty"`
	ExceedsRedThreshold     bool                     `json:"exceeds_red_threshold"`    // specific red flash limits
}

// PatternAnalysis analyzes spatial patterns that may trigger seizures.
// Certain visual patterns (stripes, checkerboards, spirals) can be dangerous
// for photosensitive individuals even without flashing.
type PatternAnalysis struct {
	HasStripedPatterns      bool                    `json:"has_striped_patterns"`
	HasCheckerboardPatterns bool                    `json:"has_checkerboard_patterns"`
	HasSpiralPatterns       bool                    `json:"has_spiral_patterns"`
	HasRadialPatterns       bool                    `json:"has_radial_patterns"`
	PatternFrequency        float64                 `json:"pattern_frequency"`        // cycles per degree
	PatternContrast         float64                 `json:"pattern_contrast"`         // 0-1
	PatternInstances        []PatternInstance       `json:"pattern_instances,omitempty"`
	ExceedsPatternThreshold bool                    `json:"exceeds_pattern_threshold"`
	HighRiskPatterns        []HighRiskPattern       `json:"high_risk_patterns,omitempty"`
}

// LuminanceAnalysis analyzes luminance changes that may trigger seizures.
// Rapid changes in brightness can be dangerous even without color flashing.
type LuminanceAnalysis struct {
	LuminanceFlashes        int                     `json:"luminance_flashes"`
	MaxLuminanceChange      float64                 `json:"max_luminance_change"`     // cd/m²
	LuminanceChangeRate     float64                 `json:"luminance_change_rate"`    // changes per second
	LuminanceTransitions    []LuminanceTransition   `json:"luminance_transitions,omitempty"`
	BrightnessVariation     *BrightnessVariation    `json:"brightness_variation,omitempty"`
	ContrastAnalysis        *ContrastAnalysis       `json:"contrast_analysis,omitempty"`
}

// TemporalPSEAnalysis analyzes temporal aspects of PSE risk.
// This examines how risk factors change over time and identifies critical periods.
type TemporalPSEAnalysis struct {
	AnalysisDuration        float64                 `json:"analysis_duration"`        // seconds
	SamplingRate            float64                 `json:"sampling_rate"`            // samples per second
	TemporalWindows         []TemporalWindow        `json:"temporal_windows,omitempty"`
	FrequencyAnalysis       *FrequencyAnalysis      `json:"frequency_analysis,omitempty"`
	RhythmAnalysis          *RhythmAnalysis         `json:"rhythm_analysis,omitempty"`
	CriticalTimeWindows     []CriticalTimeWindow    `json:"critical_time_windows,omitempty"`
}

// SpatialPSEAnalysis analyzes spatial aspects of PSE risk.
// This examines how dangerous patterns are distributed across the screen
// and their impact on different areas of vision.
type SpatialPSEAnalysis struct {
	ScreenCoverage          float64                 `json:"screen_coverage"`          // 0-1, portion of screen with risky content
	CentralVisionImpact     float64                 `json:"central_vision_impact"`    // 0-1, impact on central vision
	PeripheralVisionImpact  float64                 `json:"peripheral_vision_impact"` // 0-1, impact on peripheral vision
	SpatialExtent           *SpatialExtent          `json:"spatial_extent,omitempty"`
	ViewingAngleAnalysis    *ViewingAngleAnalysis   `json:"viewing_angle_analysis,omitempty"`
	RegionAnalysis          []RegionRiskAnalysis    `json:"region_analysis,omitempty"`
}

// SceneAnalysis analyzes scene types and their PSE risk.
// Different types of content (action scenes, concerts, etc.) have different risk profiles.
type SceneAnalysis struct {
	SceneTypes              []SceneType             `json:"scene_types,omitempty"`
	HighRiskScenes          []HighRiskScene         `json:"high_risk_scenes,omitempty"`
	SceneTransitions        []SceneTransition       `json:"scene_transitions,omitempty"`
	ContentClassification   *ContentClassification  `json:"content_classification,omitempty"`
	MotionAnalysis          *MotionAnalysis         `json:"motion_analysis,omitempty"`
}

// Supporting data structures for detailed analysis

// FlashDuration represents the duration and characteristics of a flash event.
type FlashDuration struct {
	StartTime    float64 `json:"start_time"`    // seconds
	EndTime      float64 `json:"end_time"`      // seconds
	Duration     float64 `json:"duration"`      // seconds
	Intensity    float64 `json:"intensity"`     // 0-1
	ScreenArea   float64 `json:"screen_area"`   // fraction of screen affected
	RiskLevel    string  `json:"risk_level"`    // low, medium, high
}

// TimePeriod represents a time range with associated risk characteristics.
type TimePeriod struct {
	StartTime    float64 `json:"start_time"`    // seconds
	EndTime      float64 `json:"end_time"`      // seconds
	RiskLevel    string  `json:"risk_level"`    // low, medium, high, critical
	Description  string  `json:"description"`   // human-readable description
	Confidence   float64 `json:"confidence"`    // 0-1 confidence in assessment
}

// FlashIntensity contains detailed intensity analysis for flash events.
type FlashIntensity struct {
	PeakIntensity       float64 `json:"peak_intensity"`       // maximum intensity level
	AverageIntensity    float64 `json:"average_intensity"`    // average across all flashes
	IntensityVariance   float64 `json:"intensity_variance"`   // variability measure
	IntensityDistribution map[string]int `json:"intensity_distribution"` // histogram
}

// PatternInstance represents a detected dangerous pattern.
type PatternInstance struct {
	PatternType     string  `json:"pattern_type"`     // stripe, checkerboard, spiral, etc.
	StartTime       float64 `json:"start_time"`       // seconds
	Duration        float64 `json:"duration"`         // seconds
	ScreenRegion    Rectangle `json:"screen_region"`  // spatial location
	Frequency       float64 `json:"frequency"`        // cycles per degree
	Contrast        float64 `json:"contrast"`         // 0-1
	RiskLevel       string  `json:"risk_level"`       // low, medium, high
}

// HighRiskPattern represents patterns that exceed safety thresholds.
type HighRiskPattern struct {
	PatternType     string  `json:"pattern_type"`     // type of dangerous pattern
	Timestamp       float64 `json:"timestamp"`        // when pattern occurs
	Duration        float64 `json:"duration"`         // how long pattern lasts
	SeverityScore   float64 `json:"severity_score"`   // 0-100 severity rating
	Description     string  `json:"description"`      // human-readable explanation
	Recommendation  string  `json:"recommendation"`   // suggested action
}

// Additional supporting structures (simplified definitions)
type LuminanceTransition struct {
	Timestamp       float64 `json:"timestamp"`
	FromLuminance   float64 `json:"from_luminance"`
	ToLuminance     float64 `json:"to_luminance"`
	ChangeRate      float64 `json:"change_rate"`
	RiskLevel       string  `json:"risk_level"`
}

type BrightnessVariation struct {
	StandardDeviation float64 `json:"standard_deviation"`
	PeakToPeak       float64 `json:"peak_to_peak"`
	AverageChange    float64 `json:"average_change"`
}

type ContrastAnalysis struct {
	MaxContrast      float64 `json:"max_contrast"`
	AverageContrast  float64 `json:"average_contrast"`
	ContrastChanges  int     `json:"contrast_changes"`
}

type TemporalWindow struct {
	StartTime        float64 `json:"start_time"`
	EndTime          float64 `json:"end_time"`
	RiskScore        float64 `json:"risk_score"`
	PrimaryRiskFactor string `json:"primary_risk_factor"`
}

type FrequencyAnalysis struct {
	DominantFrequency float64 `json:"dominant_frequency"`
	FrequencySpread   float64 `json:"frequency_spread"`
	HarmonicContent   float64 `json:"harmonic_content"`
}

type RhythmAnalysis struct {
	RegularRhythm     bool    `json:"regular_rhythm"`
	RhythmFrequency   float64 `json:"rhythm_frequency"`
	RhythmStability   float64 `json:"rhythm_stability"`
}

type CriticalTimeWindow struct {
	StartTime         float64 `json:"start_time"`
	EndTime           float64 `json:"end_time"`
	CriticalityLevel  string  `json:"criticality_level"`
	RiskFactors       []string `json:"risk_factors"`
	Recommendation    string  `json:"recommendation"`
}

type SpatialExtent struct {
	Width            float64 `json:"width"`            // fraction of screen width
	Height           float64 `json:"height"`           // fraction of screen height
	CenterX          float64 `json:"center_x"`         // center position
	CenterY          float64 `json:"center_y"`         // center position
}

type ViewingAngleAnalysis struct {
	MaxViewingAngle  float64 `json:"max_viewing_angle"`  // degrees
	CriticalAngle    float64 `json:"critical_angle"`     // degrees at which risk increases
	AngularExtent    float64 `json:"angular_extent"`     // degrees of visual field affected
}

type RegionRiskAnalysis struct {
	Region           Rectangle `json:"region"`           // screen region
	RiskLevel        string    `json:"risk_level"`       // risk assessment for this region
	RiskFactors      []string  `json:"risk_factors"`     // what makes this region risky
	Coverage         float64   `json:"coverage"`         // time coverage for this region
}

type SceneType struct {
	Type             string  `json:"type"`             // action, concert, nature, etc.
	StartTime        float64 `json:"start_time"`       // seconds
	EndTime          float64 `json:"end_time"`         // seconds
	RiskLevel        string  `json:"risk_level"`       // inherent risk for this scene type
	Confidence       float64 `json:"confidence"`       // 0-1 confidence in classification
}

type HighRiskScene struct {
	SceneType        string  `json:"scene_type"`       // type of scene
	StartTime        float64 `json:"start_time"`       // seconds
	Duration         float64 `json:"duration"`         // seconds
	RiskFactors      []string `json:"risk_factors"`    // what makes it risky
	SeverityScore    float64 `json:"severity_score"`   // 0-100
	Recommendation   string  `json:"recommendation"`   // suggested action
}

type SceneTransition struct {
	Timestamp        float64 `json:"timestamp"`        // when transition occurs
	FromScene        string  `json:"from_scene"`       // previous scene type
	ToScene          string  `json:"to_scene"`         // new scene type
	TransitionType   string  `json:"transition_type"`  // cut, fade, dissolve, etc.
	RiskIncrease     bool    `json:"risk_increase"`    // whether transition increases risk
}

type ContentClassification struct {
	Genre            string   `json:"genre"`            // movie, sports, news, etc.
	SubGenre         string   `json:"sub_genre"`        // action, documentary, etc.
	RiskProfile      string   `json:"risk_profile"`     // low, medium, high risk content type
	AgeRating        string   `json:"age_rating"`       // content age rating
	PSEWarningNeeded bool     `json:"pse_warning_needed"` // whether PSE warning is required
}

type MotionAnalysis struct {
	AverageMotion    float64  `json:"average_motion"`   // average motion intensity
	PeakMotion       float64  `json:"peak_motion"`      // peak motion intensity
	MotionVariability float64 `json:"motion_variability"` // how much motion varies
	FastMotionPeriods []TimePeriod `json:"fast_motion_periods,omitempty"` // periods of high motion
}

type Rectangle struct {
	X      float64 `json:"x"`      // left edge (0-1)
	Y      float64 `json:"y"`      // top edge (0-1)
	Width  float64 `json:"width"`  // width (0-1)
	Height float64 `json:"height"` // height (0-1)
}