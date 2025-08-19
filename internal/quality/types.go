package quality

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// QualityMetricType represents the type of quality metric
type QualityMetricType string

const (
	MetricVMAF   QualityMetricType = "vmaf"
	MetricPSNR   QualityMetricType = "psnr"
	MetricSSIM   QualityMetricType = "ssim"
	MetricMSE    QualityMetricType = "mse"
	MetricMSSSIM QualityMetricType = "ms_ssim" // Multi-Scale SSIM
	MetricLPIPS  QualityMetricType = "lpips"   // Learned Perceptual Image Patch Similarity
)

// QualityAnalysis represents a complete quality analysis
type QualityAnalysis struct {
	ID             uuid.UUID             `json:"id" db:"id"`
	AnalysisID     uuid.UUID             `json:"analysis_id" db:"analysis_id"`
	ReferenceFile  string                `json:"reference_file" db:"reference_file"`
	DistortedFile  string                `json:"distorted_file" db:"distorted_file"`
	MetricType     QualityMetricType     `json:"metric_type" db:"metric_type"`
	OverallScore   float64               `json:"overall_score" db:"overall_score"`
	MinScore       float64               `json:"min_score" db:"min_score"`
	MaxScore       float64               `json:"max_score" db:"max_score"`
	MeanScore      float64               `json:"mean_score" db:"mean_score"`
	MedianScore    float64               `json:"median_score" db:"median_score"`
	StdDevScore    float64               `json:"std_dev_score" db:"std_dev_score"`
	Percentile1    float64               `json:"percentile_1" db:"percentile_1"`
	Percentile5    float64               `json:"percentile_5" db:"percentile_5"`
	Percentile10   float64               `json:"percentile_10" db:"percentile_10"`
	Percentile25   float64               `json:"percentile_25" db:"percentile_25"`
	Percentile75   float64               `json:"percentile_75" db:"percentile_75"`
	Percentile90   float64               `json:"percentile_90" db:"percentile_90"`
	Percentile95   float64               `json:"percentile_95" db:"percentile_95"`
	Percentile99   float64               `json:"percentile_99" db:"percentile_99"`
	FrameCount     int                   `json:"frame_count" db:"frame_count"`
	Duration       float64               `json:"duration" db:"duration"`
	Width          int                   `json:"width" db:"width"`
	Height         int                   `json:"height" db:"height"`
	FrameRate      float64               `json:"frame_rate" db:"frame_rate"`
	BitRate        int64                 `json:"bit_rate" db:"bit_rate"`
	Configuration  json.RawMessage       `json:"configuration" db:"configuration"`
	ProcessingTime time.Duration         `json:"processing_time" db:"processing_time"`
	Status         QualityAnalysisStatus `json:"status" db:"status"`
	ErrorMessage   string                `json:"error_message,omitempty" db:"error_message"`
	CreatedAt      time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at" db:"updated_at"`
	CompletedAt    *time.Time            `json:"completed_at,omitempty" db:"completed_at"`
}

// QualityAnalysisStatus represents the status of a quality analysis
type QualityAnalysisStatus string

const (
	QualityStatusPending    QualityAnalysisStatus = "pending"
	QualityStatusProcessing QualityAnalysisStatus = "processing"
	QualityStatusCompleted  QualityAnalysisStatus = "completed"
	QualityStatusFailed     QualityAnalysisStatus = "failed"
	QualityStatusCancelled  QualityAnalysisStatus = "cancelled"
)

// QualityStatus is an alias for QualityAnalysisStatus for backward compatibility
type QualityStatus = QualityAnalysisStatus

// QualityFrameMetric represents per-frame quality metrics
type QualityFrameMetric struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	QualityID      uuid.UUID       `json:"quality_id" db:"quality_id"`
	FrameNumber    int             `json:"frame_number" db:"frame_number"`
	Timestamp      float64         `json:"timestamp" db:"timestamp"`
	Score          float64         `json:"score" db:"score"`
	ComponentY     float64         `json:"component_y,omitempty" db:"component_y"`
	ComponentU     float64         `json:"component_u,omitempty" db:"component_u"`
	ComponentV     float64         `json:"component_v,omitempty" db:"component_v"`
	AdditionalData json.RawMessage `json:"additional_data,omitempty" db:"additional_data"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// VMAFConfiguration represents VMAF-specific configuration
type VMAFConfiguration struct {
	Model           string   `json:"model"`             // VMAF model version
	CustomModelPath string   `json:"custom_model_path"` // Path to custom VMAF model
	CustomModelID   string   `json:"custom_model_id"`   // ID of custom model from database
	SubSampling     int      `json:"sub_sampling"`      // Frame subsampling rate
	PoolingMethod   string   `json:"pooling_method"`    // mean, harmonic_mean, min
	NThreads        int      `json:"n_threads"`         // Number of threads
	Features        []string `json:"features"`          // Additional features to compute
	OutputFormat    string   `json:"output_format"`     // json, xml, csv
	LogLevel        string   `json:"log_level"`         // info, warning, error
	EnableTransform bool     `json:"enable_transform"`  // Enable score transformation
	PhoneModel      bool     `json:"phone_model"`       // Use phone model
}

// PSNRConfiguration represents PSNR-specific configuration
type PSNRConfiguration struct {
	ComponentMask string `json:"component_mask"` // Which components to analyze (Y, U, V)
	Stats         bool   `json:"stats"`          // Output additional statistics
	OutputFormat  string `json:"output_format"`  // json, csv
}

// SSIMConfiguration represents SSIM-specific configuration
type SSIMConfiguration struct {
	WindowSize   int     `json:"window_size"`   // Sliding window size
	K1           float64 `json:"k1"`            // Algorithm parameter
	K2           float64 `json:"k2"`            // Algorithm parameter
	Stats        bool    `json:"stats"`         // Output additional statistics
	OutputFormat string  `json:"output_format"` // json, csv
}

// QualityComparisonRequest represents a request to compare video quality
type QualityComparisonRequest struct {
	ReferenceFile string              `json:"reference_file" binding:"required"`
	DistortedFile string              `json:"distorted_file" binding:"required"`
	Metrics       []QualityMetricType `json:"metrics" binding:"required"`
	Configuration QualityConfig       `json:"configuration,omitempty"`
	Async         bool                `json:"async,omitempty"`
	FrameLevel    bool                `json:"frame_level,omitempty"` // Include per-frame metrics
	SaveFrames    bool                `json:"save_frames,omitempty"` // Save frame-level data to DB
}

// QualityConfig contains configuration for all quality metrics
type QualityConfig struct {
	VMAF    *VMAFConfiguration `json:"vmaf,omitempty"`
	PSNR    *PSNRConfiguration `json:"psnr,omitempty"`
	SSIM    *SSIMConfiguration `json:"ssim,omitempty"`
	Timeout *int               `json:"timeout,omitempty"` // Analysis timeout in seconds
}

// QualityResult represents the result of a quality analysis
type QualityResult struct {
	ID             uuid.UUID             `json:"id"`
	Status         QualityAnalysisStatus `json:"status"`
	Analysis       []*QualityAnalysis    `json:"analysis"`
	FrameMetrics   []*QualityFrameMetric `json:"frame_metrics,omitempty"`
	Summary        *QualitySummary       `json:"summary"`
	Visualization  *QualityVisualization `json:"visualization,omitempty"`
	ProcessingTime time.Duration         `json:"processing_time"`
	Message        string                `json:"message,omitempty"`
	Error          string                `json:"error,omitempty"`
}

// QualitySummary provides a human-readable summary of quality metrics
type QualitySummary struct {
	ReferenceFile      string                               `json:"reference_file"`
	DistortedFile      string                               `json:"distorted_file"`
	OverallRating      QualityRating                        `json:"overall_rating"`
	MetricSummaries    map[QualityMetricType]*MetricSummary `json:"metric_summaries"`
	Recommendations    []string                             `json:"recommendations"`
	QualityIssues      []QualityIssue                       `json:"quality_issues"`
	ComparisonInsights string                               `json:"comparison_insights"`
}

// MetricSummary provides summary for a specific metric
type MetricSummary struct {
	MetricType     QualityMetricType `json:"metric_type"`
	Score          float64           `json:"score"`
	Rating         QualityRating     `json:"rating"`
	Description    string            `json:"description"`
	Interpretation string            `json:"interpretation"`
	FrameStats     *FrameStatistics  `json:"frame_stats,omitempty"`
}

// FrameStatistics provides frame-level statistics
type FrameStatistics struct {
	TotalFrames   int     `json:"total_frames"`
	HighQuality   int     `json:"high_quality_frames"`   // Frames above threshold
	MediumQuality int     `json:"medium_quality_frames"` // Frames in medium range
	LowQuality    int     `json:"low_quality_frames"`    // Frames below threshold
	WorstFrame    int     `json:"worst_frame_number"`
	WorstScore    float64 `json:"worst_frame_score"`
	BestFrame     int     `json:"best_frame_number"`
	BestScore     float64 `json:"best_frame_score"`
}

// QualityRating represents overall quality rating
type QualityRating string

const (
	RatingExcellent QualityRating = "excellent" // 90-100
	RatingGood      QualityRating = "good"      // 80-89
	RatingFair      QualityRating = "fair"      // 70-79
	RatingPoor      QualityRating = "poor"      // 50-69
	RatingBad       QualityRating = "bad"       // <50
)

// QualityIssue represents a detected quality issue
type QualityIssue struct {
	QualityID   uuid.UUID   `json:"quality_id,omitempty"`
	Type        string      `json:"type"`     // "blocking", "blurring", "noise", etc.
	Severity    string      `json:"severity"` // "high", "medium", "low"
	Description string      `json:"description"`
	FrameRange  *FrameRange `json:"frame_range,omitempty"`
	Timestamp   *TimeRange  `json:"timestamp,omitempty"`
	Score       float64     `json:"score"`
}

// FrameRange represents a range of frames
type FrameRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start float64 `json:"start"` // seconds
	End   float64 `json:"end"`   // seconds
}

// QualityVisualization contains data for quality visualization
type QualityVisualization struct {
	ChartData     json.RawMessage `json:"chart_data"`     // Chart.js compatible data
	HeatmapData   json.RawMessage `json:"heatmap_data"`   // Heatmap visualization data
	HistogramData json.RawMessage `json:"histogram_data"` // Score distribution
	TimelineData  json.RawMessage `json:"timeline_data"`  // Timeline visualization
}

// QualityThresholds defines quality rating thresholds for different metrics
type QualityThresholds struct {
	VMAF VMAFThresholds `json:"vmaf"`
	PSNR PSNRThresholds `json:"psnr"`
	SSIM SSIMThresholds `json:"ssim"`
}

// VMAFThresholds defines VMAF-specific quality thresholds
type VMAFThresholds struct {
	Excellent float64 `json:"excellent"` // >= 95
	Good      float64 `json:"good"`      // >= 85
	Fair      float64 `json:"fair"`      // >= 75
	Poor      float64 `json:"poor"`      // >= 60
	// < Poor = Bad
}

// PSNRThresholds defines PSNR-specific quality thresholds (dB)
type PSNRThresholds struct {
	Excellent float64 `json:"excellent"` // >= 40 dB
	Good      float64 `json:"good"`      // >= 35 dB
	Fair      float64 `json:"fair"`      // >= 30 dB
	Poor      float64 `json:"poor"`      // >= 25 dB
	// < Poor = Bad
}

// SSIMThresholds defines SSIM-specific quality thresholds
type SSIMThresholds struct {
	Excellent float64 `json:"excellent"` // >= 0.95
	Good      float64 `json:"good"`      // >= 0.90
	Fair      float64 `json:"fair"`      // >= 0.85
	Poor      float64 `json:"poor"`      // >= 0.80
	// < Poor = Bad
}

// DefaultQualityThresholds returns default quality thresholds
func DefaultQualityThresholds() QualityThresholds {
	return QualityThresholds{
		VMAF: VMAFThresholds{
			Excellent: 95.0,
			Good:      85.0,
			Fair:      75.0,
			Poor:      60.0,
		},
		PSNR: PSNRThresholds{
			Excellent: 40.0,
			Good:      35.0,
			Fair:      30.0,
			Poor:      25.0,
		},
		SSIM: SSIMThresholds{
			Excellent: 0.95,
			Good:      0.90,
			Fair:      0.85,
			Poor:      0.80,
		},
	}
}

// GetRating returns the quality rating based on score and metric type
func (qt QualityThresholds) GetRating(metricType QualityMetricType, score float64) QualityRating {
	var thresholds interface{}

	switch metricType {
	case MetricVMAF:
		thresholds = qt.VMAF
	case MetricPSNR:
		thresholds = qt.PSNR
	case MetricSSIM:
		thresholds = qt.SSIM
	default:
		return RatingFair // Default for unknown metrics
	}

	// Use reflection or type switching to get thresholds
	switch t := thresholds.(type) {
	case VMAFThresholds:
		if score >= t.Excellent {
			return RatingExcellent
		} else if score >= t.Good {
			return RatingGood
		} else if score >= t.Fair {
			return RatingFair
		} else if score >= t.Poor {
			return RatingPoor
		}
		return RatingBad
	case PSNRThresholds:
		if score >= t.Excellent {
			return RatingExcellent
		} else if score >= t.Good {
			return RatingGood
		} else if score >= t.Fair {
			return RatingFair
		} else if score >= t.Poor {
			return RatingPoor
		}
		return RatingBad
	case SSIMThresholds:
		if score >= t.Excellent {
			return RatingExcellent
		} else if score >= t.Good {
			return RatingGood
		} else if score >= t.Fair {
			return RatingFair
		} else if score >= t.Poor {
			return RatingPoor
		}
		return RatingBad
	}

	return RatingFair
}

// QualityProgressUpdate represents progress updates for quality analysis
type QualityProgressUpdate struct {
	ID              uuid.UUID             `json:"id"`
	Status          QualityAnalysisStatus `json:"status"`
	Progress        float64               `json:"progress"` // 0.0 to 1.0
	CurrentMetric   QualityMetricType     `json:"current_metric"`
	ProcessedFrames int                   `json:"processed_frames"`
	TotalFrames     int                   `json:"total_frames"`
	EstimatedTime   time.Duration         `json:"estimated_time"`
	Message         string                `json:"message"`
}

// BatchQualityRequest represents a batch quality analysis request
type BatchQualityRequest struct {
	Comparisons []QualityComparisonRequest `json:"comparisons" binding:"required"`
	Async       bool                       `json:"async,omitempty"`
	Parallel    int                        `json:"parallel,omitempty"` // Number of parallel analyses
}

// BatchQualityResult represents results from batch quality analysis
type BatchQualityResult struct {
	BatchID   uuid.UUID            `json:"batch_id"`
	Status    string               `json:"status"`
	Total     int                  `json:"total"`
	Completed int                  `json:"completed"`
	Failed    int                  `json:"failed"`
	Results   []*QualityResult     `json:"results"`
	Summary   *BatchQualitySummary `json:"summary,omitempty"`
}

// BatchQualitySummary provides summary of batch analysis
type BatchQualitySummary struct {
	OverallRating   QualityRating                 `json:"overall_rating"`
	AverageScores   map[QualityMetricType]float64 `json:"average_scores"`
	BestPerforming  string                        `json:"best_performing"`
	WorstPerforming string                        `json:"worst_performing"`
	Recommendations []string                      `json:"recommendations"`
	ProcessingTime  time.Duration                 `json:"processing_time"`
}
