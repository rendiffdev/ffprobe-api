package ffmpeg

import "time"

// PSE (Photosensitive Epilepsy) analysis constants and thresholds.
// These constants are based on international broadcast safety standards
// and medical research on photosensitive epilepsy triggers.

// Flash Analysis Thresholds (ITU-R BT.1702 standards)
const (
	// MaxSafeFlashRate is the maximum safe flash rate in flashes per second.
	// Based on ITU-R BT.1702 recommendation: 3 flashes per second maximum.
	MaxSafeFlashRate = 3.0
	
	// CriticalFlashRate is the rate above which content is considered dangerous.
	// Rates above 3-25 Hz are particularly dangerous for photosensitive individuals.
	CriticalFlashRate = 5.0
	
	// MaxSafeRedFlashRate is the maximum safe rate for red flashes specifically.
	// Red flashes are more dangerous and have a lower threshold.
	MaxSafeRedFlashRate = 2.0
	
	// FlashDurationThreshold is the minimum duration to consider a flash event.
	// Measured in seconds - very brief changes are less likely to trigger seizures.
	FlashDurationThreshold = 0.033 // ~1 frame at 30fps
	
	// MinFlashIntensity is the minimum intensity change to register as a flash.
	// Measured as relative luminance change (0-1 scale).
	MinFlashIntensity = 0.1
	
	// DangerousFlashIntensity is the intensity above which flashes become dangerous.
	DangerousFlashIntensity = 0.8
)

// Luminance Analysis Thresholds
const (
	// MaxSafeLuminanceChange is the maximum safe luminance change in cd/m².
	// Based on ITC guidelines for broadcast safety.
	MaxSafeLuminanceChange = 160.0
	
	// CriticalLuminanceChange is the luminance change that requires warning.
	CriticalLuminanceChange = 320.0
	
	// LuminanceChangeRateThreshold is the maximum safe rate of luminance change.
	// Measured in cd/m² per second.
	LuminanceChangeRateThreshold = 80.0
	
	// MinLuminanceForAnalysis is the minimum luminance level to consider.
	// Very dark content has different risk characteristics.
	MinLuminanceForAnalysis = 5.0
)

// Pattern Analysis Thresholds (Spatial frequency analysis)
const (
	// MaxSafePatternFrequency is the maximum safe spatial frequency in cycles per degree.
	// Patterns above this frequency can trigger seizures in susceptible individuals.
	MaxSafePatternFrequency = 5.0
	
	// CriticalPatternFrequency is the frequency above which patterns are dangerous.
	CriticalPatternFrequency = 8.0
	
	// MinPatternContrast is the minimum contrast to consider a pattern dangerous.
	// Low contrast patterns are generally safer.
	MinPatternContrast = 0.3
	
	// DangerousPatternContrast is the contrast above which patterns become risky.
	DangerousPatternContrast = 0.7
	
	// PatternAreaThreshold is the minimum screen area for pattern analysis.
	// Small patterns covering minimal screen area are less dangerous.
	PatternAreaThreshold = 0.1 // 10% of screen
	
	// CriticalPatternArea is the screen coverage above which patterns are dangerous.
	CriticalPatternArea = 0.25 // 25% of screen
)

// Temporal Analysis Constants
const (
	// AnalysisWindowSize is the time window for temporal analysis in seconds.
	// ITU-R BT.1702 specifies analysis in 1-second windows.
	AnalysisWindowSize = 1.0
	
	// SamplingRate is the minimum sampling rate for PSE analysis in Hz.
	// Higher sampling rates provide more accurate detection.
	MinSamplingRate = 25.0
	
	// RecommendedSamplingRate is the recommended sampling rate for accurate analysis.
	RecommendedSamplingRate = 50.0
	
	// MaxAnalysisDuration is the maximum duration to analyze in seconds.
	// Very long content may need segmented analysis.
	MaxAnalysisDuration = 3600.0 // 1 hour
	
	// MinAnalysisDuration is the minimum duration needed for reliable analysis.
	MinAnalysisDuration = 0.5 // 0.5 seconds
)

// Risk Level Thresholds
const (
	// Risk scores are on a 0-100 scale
	SafeRiskThreshold     = 20.0  // Below this is considered safe
	LowRiskThreshold      = 40.0  // Low risk but monitor
	MediumRiskThreshold   = 60.0  // Medium risk - warning advised
	HighRiskThreshold     = 80.0  // High risk - warning required
	CriticalRiskThreshold = 95.0  // Critical risk - content review needed
)

// Spatial Analysis Constants
const (
	// CentralVisionRadius is the radius of central vision area as fraction of screen.
	CentralVisionRadius = 0.1 // 10% of screen radius
	
	// PeripheralVisionThreshold is the boundary for peripheral vision analysis.
	PeripheralVisionThreshold = 0.4 // 40% from center
	
	// ViewingDistanceAssumption is the assumed viewing distance in screen heights.
	// Used for calculating visual angles and spatial frequencies.
	ViewingDistanceAssumption = 3.0 // 3x screen height (typical TV viewing)
	
	// MaxCriticalViewingAngle is the maximum viewing angle for critical analysis.
	// Content within this angle has higher risk impact.
	MaxCriticalViewingAngle = 10.0 // degrees
)

// Red Flash Specific Constants
const (
	// RedSaturationThreshold is the minimum red saturation to consider red flash.
	RedSaturationThreshold = 0.6
	
	// RedLuminanceWeight is the weighting factor for red in luminance calculations.
	// Red has higher seizure risk so gets additional weighting.
	RedLuminanceWeight = 1.5
	
	// RedContrastMultiplier is the multiplier for red contrast in risk calculations.
	RedContrastMultiplier = 1.3
)

// Analysis Configuration
const (
	// DefaultAnalysisTimeout is the default timeout for PSE analysis operations.
	DefaultAnalysisTimeout = 5 * time.Minute
	
	// MaxConcurrentAnalysis is the maximum number of concurrent PSE analyses.
	// PSE analysis is CPU intensive and should be limited.
	MaxConcurrentAnalysis = 2
	
	// PSEMemoryLimitPerAnalysis is the estimated memory limit per PSE analysis in MB.
	PSEMemoryLimitPerAnalysis = 512 // 512MB per PSE analysis
)

// Compliance Standards Versions
const (
	// ITU_R_BT1702_Version is the version of ITU-R BT.1702 standard implemented.
	ITU_R_BT1702_Version = "2012"
	
	// ITC_GuidelinesVersion is the version of ITC guidelines implemented.
	ITC_GuidelinesVersion = "2001"
	
	// EBU_Tech3253_Version is the version of EBU Tech 3253 implemented.
	EBU_Tech3253_Version = "2010"
	
	// StandardsLastUpdated is when these thresholds were last updated.
	StandardsLastUpdated = "2024-01-01"
)

// Risk Level Classifications
var (
	// RiskLevelNames maps risk scores to human-readable names
	RiskLevelNames = map[float64]string{
		0:   "safe",
		20:  "low",
		40:  "medium", 
		60:  "high",
		80:  "critical",
		100: "dangerous",
	}
	
	// BroadcastComplianceThresholds defines thresholds for different standards
	BroadcastComplianceThresholds = map[string]float64{
		"ITU-R BT.1702": SafeRiskThreshold,
		"ITC":           LowRiskThreshold,
		"Ofcom":         LowRiskThreshold,
		"EBU Tech 3253": MediumRiskThreshold,
		"FCC":           LowRiskThreshold,
	}
	
	// SceneTypeRiskProfiles maps scene types to base risk multipliers
	SceneTypeRiskProfiles = map[string]float64{
		"action":      1.8,  // Action scenes have higher risk
		"concert":     2.0,  // Concerts with stage lighting are very risky
		"sports":      1.3,  // Sports can have rapid camera movement
		"animation":   1.5,  // Animation can have rapid color changes
		"documentary": 1.0,  // Documentaries are typically lower risk
		"news":        1.0,  // News content is typically safe
		"nature":      0.8,  // Nature content is typically low risk
		"drama":       1.1,  // Drama has moderate risk
		"comedy":      1.0,  // Comedy is typically safe
		"horror":      1.6,  // Horror can have sudden flashes
		"sci-fi":      1.4,  // Sci-fi often has strobing effects
		"unknown":     1.2,  // Unknown content gets moderate multiplier
	}
)

// Analysis Quality Settings
const (
	// FastAnalysisSkipFrames is frame skip count for fast analysis mode
	FastAnalysisSkipFrames = 2 // Analyze every 3rd frame
	
	// StandardAnalysisSkipFrames is frame skip count for standard analysis
	StandardAnalysisSkipFrames = 1 // Analyze every 2nd frame
	
	// PrecisionAnalysisSkipFrames is frame skip count for precision analysis
	PrecisionAnalysisSkipFrames = 0 // Analyze every frame
	
	// MinFramesForReliableAnalysis is minimum frames needed for reliable results
	MinFramesForReliableAnalysis = 25 // ~1 second at 25fps
)

// Warning and Compliance Messages
var (
	// StandardWarningMessages for different risk levels
	StandardWarningMessages = map[string]string{
		"safe":     "Content appears safe for all viewers including those with photosensitive epilepsy.",
		"low":      "Content has low risk. May be suitable with standard PSE warnings.",
		"medium":   "Content has medium PSE risk. PSE warning recommended before broadcast.",
		"high":     "Content has high PSE risk. PSE warning required. Consider content review.",
		"critical": "Content has critical PSE risk. Requires content modification or strong warnings.",
		"dangerous": "Content exceeds safe limits. Requires immediate attention and modification.",
	}
	
	// ComplianceRecommendations for different standards
	ComplianceRecommendations = map[string]string{
		"ITU-R BT.1702": "Follow ITU-R BT.1702 guidelines for international broadcast compliance.",
		"ITC":           "Comply with ITC guidelines for UK terrestrial broadcast.",
		"Ofcom":         "Meet Ofcom requirements for UK broadcast licensing.",
		"EBU Tech 3253": "Follow EBU Tech 3253 for European broadcast compliance.",
		"FCC":           "Adhere to FCC guidelines for US broadcast standards.",
	}
)