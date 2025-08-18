package graphql

import (
	"encoding/json"
	"time"
)

// GraphQL Type Definitions
// These structs represent the GraphQL schema types

// Video Analysis Types
type VideoAnalysis struct {
	ID              string             `json:"id"`
	FilePath        string             `json:"filePath"`
	FileName        string             `json:"fileName"`
	FileSize        int                `json:"fileSize"`
	Duration        float64            `json:"duration"`
	Bitrate         int                `json:"bitrate"`
	Format          *VideoFormat       `json:"format"`
	Streams         []*Stream          `json:"streams"`
	QualityMetrics  *QualityMetrics    `json:"qualityMetrics"`
	ContentAnalysis *ContentAnalysis   `json:"contentAnalysis"`
	HlsInfo         *HLSInfo           `json:"hlsInfo"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	Status          AnalysisStatus     `json:"status"`
	ProcessingTime  *float64           `json:"processingTime"`
	Metadata        json.RawMessage    `json:"metadata"`
}

type VideoFormat struct {
	Name      string          `json:"name"`
	LongName  string          `json:"longName"`
	Size      int             `json:"size"`
	Bitrate   int             `json:"bitrate"`
	Duration  float64         `json:"duration"`
	StartTime *float64        `json:"startTime"`
	Tags      json.RawMessage `json:"tags"`
}

type Stream struct {
	Index               int             `json:"index"`
	CodecName           string          `json:"codecName"`
	CodecLongName       string          `json:"codecLongName"`
	CodecType           StreamType      `json:"codecType"`
	CodecTag            string          `json:"codecTag"`
	Width               *int            `json:"width"`
	Height              *int            `json:"height"`
	SampleAspectRatio   *string         `json:"sampleAspectRatio"`
	DisplayAspectRatio  *string         `json:"displayAspectRatio"`
	PixFmt              *string         `json:"pixFmt"`
	Level               *int            `json:"level"`
	ColorRange          *string         `json:"colorRange"`
	ColorSpace          *string         `json:"colorSpace"`
	ColorTransfer       *string         `json:"colorTransfer"`
	ColorPrimaries      *string         `json:"colorPrimaries"`
	ChromaLocation      *string         `json:"chromaLocation"`
	FieldOrder          *string         `json:"fieldOrder"`
	Refs                *int            `json:"refs"`
	Profile             *string         `json:"profile"`
	Bitrate             *int            `json:"bitrate"`
	MaxBitrate          *int            `json:"maxBitrate"`
	BufferSize          *int            `json:"bufferSize"`
	Framerate           *float64        `json:"framerate"`
	AvgFramerate        *float64        `json:"avgFramerate"`
	TimeBase            *string         `json:"timeBase"`
	StartPts            *int            `json:"startPts"`
	StartTime           *float64        `json:"startTime"`
	Duration            *float64        `json:"duration"`
	BitDepth            *int            `json:"bitDepth"`
	Channels            *int            `json:"channels"`
	ChannelLayout       *string         `json:"channelLayout"`
	SampleRate          *int            `json:"sampleRate"`
	SampleFmt           *string         `json:"sampleFmt"`
	Tags                json.RawMessage `json:"tags"`
}

type QualityMetrics struct {
	ID              string          `json:"id"`
	VideoAnalysisID string          `json:"videoAnalysisId"`
	VmafScore       *float64        `json:"vmafScore"`
	Psnr            *float64        `json:"psnr"`
	Ssim            *float64        `json:"ssim"`
	Msssim          *float64        `json:"msssim"`
	Butteraugli     *float64        `json:"butteraugli"`
	Lpips           *float64        `json:"lpips"`
	Dssim           *float64        `json:"dssim"`
	Fsim            *float64        `json:"fsim"`
	Vsi             *float64        `json:"vsi"`
	Haarpsi         *float64        `json:"haarpsi"`
	Mdsi            *float64        `json:"mdsi"`
	Gmsd            *float64        `json:"gmsd"`
	Mse             *float64        `json:"mse"`
	PsnrHvs         *float64        `json:"psnrHvs"`
	PsnrHvsM        *float64        `json:"psnrHvsM"`
	Ciede2000       *float64        `json:"ciede2000"`
	Cambi           *float64        `json:"cambi"`
	Blockiness      *float64        `json:"blockiness"`
	Blur            *float64        `json:"blur"`
	Ta              *float64        `json:"ta"`
	Si              *float64        `json:"si"`
	Ti              *float64        `json:"ti"`
	Niqe            *float64        `json:"niqe"`
	Brisque         *float64        `json:"brisque"`
	Piqe            *float64        `json:"piqe"`
	Ilniqe          *float64        `json:"ilniqe"`
	CreatedAt       time.Time       `json:"createdAt"`
	Metadata        json.RawMessage `json:"metadata"`
}

type ContentAnalysis struct {
	ID               string             `json:"id"`
	VideoAnalysisID  string             `json:"videoAnalysisId"`
	BlackFrames      []*BlackFrame      `json:"blackFrames"`
	FreezeFrames     []*FreezeFrame     `json:"freezeFrames"`
	SilenceSegments  []*SilenceSegment  `json:"silenceSegments"`
	AudioClipping    []*AudioClipping   `json:"audioClipping"`
	SceneChanges     []*SceneChange     `json:"sceneChanges"`
	MotionAnalysis   *MotionAnalysis    `json:"motionAnalysis"`
	ColorAnalysis    *ColorAnalysis     `json:"colorAnalysis"`
	LoudnessAnalysis *LoudnessAnalysis  `json:"loudnessAnalysis"`
	CreatedAt        time.Time          `json:"createdAt"`
	Metadata         json.RawMessage    `json:"metadata"`
}

type BlackFrame struct {
	Timestamp  float64 `json:"timestamp"`
	Duration   float64 `json:"duration"`
	Percentage float64 `json:"percentage"`
}

type FreezeFrame struct {
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
	Duration  float64 `json:"duration"`
}

type SilenceSegment struct {
	StartTime        float64 `json:"startTime"`
	EndTime          float64 `json:"endTime"`
	Duration         float64 `json:"duration"`
	SilenceThreshold float64 `json:"silenceThreshold"`
}

type AudioClipping struct {
	Channel   int     `json:"channel"`
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
	Duration  float64 `json:"duration"`
	PeakLevel float64 `json:"peakLevel"`
}

type SceneChange struct {
	Timestamp   float64 `json:"timestamp"`
	Score       float64 `json:"score"`
	FrameNumber int     `json:"frameNumber"`
}

type MotionAnalysis struct {
	AverageMotion       float64   `json:"averageMotion"`
	MaxMotion           float64   `json:"maxMotion"`
	MinMotion           float64   `json:"minMotion"`
	MotionVariance      float64   `json:"motionVariance"`
	MotionDistribution  []float64 `json:"motionDistribution"`
}

type ColorAnalysis struct {
	AverageBrightness float64         `json:"averageBrightness"`
	Colorfulness      float64         `json:"colorfulness"`
	Contrast          float64         `json:"contrast"`
	Saturation        float64         `json:"saturation"`
	DominantColors    []string        `json:"dominantColors"`
	ColorHistogram    json.RawMessage `json:"colorHistogram"`
}

type LoudnessAnalysis struct {
	IntegratedLoudness     float64 `json:"integratedLoudness"`
	LoudnessRange          float64 `json:"loudnessRange"`
	MaxTruePeak            float64 `json:"maxTruePeak"`
	MaxMomentaryLoudness   float64 `json:"maxMomentaryLoudness"`
	MaxShortTermLoudness   float64 `json:"maxShortTermLoudness"`
	EbuR128Compliant       bool    `json:"ebuR128Compliant"`
}

type HLSInfo struct {
	ID              string        `json:"id"`
	VideoAnalysisID string        `json:"videoAnalysisId"`
	MasterPlaylist  string        `json:"masterPlaylist"`
	Variants        []*HLSVariant `json:"variants"`
	Segments        []*HLSSegment `json:"segments"`
	TotalDuration   float64       `json:"totalDuration"`
	SegmentCount    int           `json:"segmentCount"`
	CreatedAt       time.Time     `json:"createdAt"`
}

type HLSVariant struct {
	Bandwidth     int      `json:"bandwidth"`
	Resolution    string   `json:"resolution"`
	Codecs        string   `json:"codecs"`
	URI           string   `json:"uri"`
	FrameRate     *float64 `json:"frameRate"`
	AudioGroup    *string  `json:"audioGroup"`
	SubtitleGroup *string  `json:"subtitleGroup"`
}

type HLSSegment struct {
	URI          string  `json:"uri"`
	Duration     float64 `json:"duration"`
	Sequence     int     `json:"sequence"`
	ByteRange    *string `json:"byteRange"`
	Discontinuity bool   `json:"discontinuity"`
}

// User and API Key Types
type User struct {
	ID             string             `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	Roles          []string           `json:"roles"`
	TenantID       string             `json:"tenantId"`
	IsActive       bool               `json:"isActive"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	LastLoginAt    *time.Time         `json:"lastLoginAt"`
	APIKeys        []*APIKey          `json:"apiKeys"`
	VideoAnalyses  []*VideoAnalysis   `json:"videoAnalyses"`
	RateLimits     *UserRateLimit     `json:"rateLimits"`
}

type APIKey struct {
	ID          string             `json:"id"`
	UserID      string             `json:"userId"`
	TenantID    string             `json:"tenantId"`
	KeyPrefix   string             `json:"keyPrefix"`
	Name        string             `json:"name"`
	Permissions []string           `json:"permissions"`
	Status      APIKeyStatus       `json:"status"`
	CreatedAt   time.Time          `json:"createdAt"`
	ExpiresAt   time.Time          `json:"expiresAt"`
	LastUsedAt  *time.Time         `json:"lastUsedAt"`
	LastRotated time.Time          `json:"lastRotated"`
	RotationDue time.Time          `json:"rotationDue"`
	UsageCount  int                `json:"usageCount"`
	RateLimits  *APIKeyRateLimit   `json:"rateLimits"`
}

type APIKeyRateLimit struct {
	PerMinute int `json:"perMinute"`
	PerHour   int `json:"perHour"`
	PerDay    int `json:"perDay"`
}

type UserRateLimit struct {
	UserID             string  `json:"userId"`
	PerMinute          int     `json:"perMinute"`
	PerHour            int     `json:"perHour"`
	PerDay             int     `json:"perDay"`
	BurstMultiplier    float64 `json:"burstMultiplier"`
	MonthlyQuota       *int    `json:"monthlyQuota"`
	CurrentMonthUsage  int     `json:"currentMonthUsage"`
	IsActive           bool    `json:"isActive"`
}

type TenantRateLimit struct {
	TenantID          string  `json:"tenantId"`
	PerMinute         int     `json:"perMinute"`
	PerHour           int     `json:"perHour"`
	PerDay            int     `json:"perDay"`
	BurstMultiplier   float64 `json:"burstMultiplier"`
	MonthlyQuota      *int    `json:"monthlyQuota"`
	CurrentMonthUsage int     `json:"currentMonthUsage"`
	Tier              string  `json:"tier"`
	IsActive          bool    `json:"isActive"`
}

// Comparison Types
type VideoComparison struct {
	ID                string              `json:"id"`
	ReferenceVideoID  string              `json:"referenceVideoId"`
	TestVideoID       string              `json:"testVideoId"`
	ReferenceVideo    *VideoAnalysis      `json:"referenceVideo"`
	TestVideo         *VideoAnalysis      `json:"testVideo"`
	ComparisonMetrics *ComparisonMetrics  `json:"comparisonMetrics"`
	Status            ComparisonStatus    `json:"status"`
	CreatedAt         time.Time           `json:"createdAt"`
	CompletedAt       *time.Time          `json:"completedAt"`
	ProcessingTime    *float64            `json:"processingTime"`
	Metadata          json.RawMessage     `json:"metadata"`
}

type ComparisonMetrics struct {
	ID                   string    `json:"id"`
	ComparisonID         string    `json:"comparisonId"`
	VmafScore            float64   `json:"vmafScore"`
	Psnr                 float64   `json:"psnr"`
	Ssim                 float64   `json:"ssim"`
	TemporalConsistency  *float64  `json:"temporalConsistency"`
	SpatialConsistency   *float64  `json:"spatialConsistency"`
	PerceptualQuality    *float64  `json:"perceptualQuality"`
	OverallScore         float64   `json:"overallScore"`
	Recommendations      []string  `json:"recommendations"`
	CreatedAt            time.Time `json:"createdAt"`
}

// Report Types
type AnalysisReport struct {
	ID              string        `json:"id"`
	VideoAnalysisID string        `json:"videoAnalysisId"`
	ReportType      ReportType    `json:"reportType"`
	Format          ReportFormat  `json:"format"`
	Content         string        `json:"content"`
	FilePath        *string       `json:"filePath"`
	GeneratedAt     time.Time     `json:"generatedAt"`
	LlmGenerated    bool          `json:"llmGenerated"`
	LlmModel        *string       `json:"llmModel"`
	Metadata        json.RawMessage `json:"metadata"`
}

// Input Types
type VideoAnalysisInput struct {
	FilePath               string          `json:"filePath"`
	EnableContentAnalysis  bool            `json:"enableContentAnalysis"`
	EnableQualityMetrics   bool            `json:"enableQualityMetrics"`
	EnableHLSAnalysis      bool            `json:"enableHLSAnalysis"`
	CustomParameters       json.RawMessage `json:"customParameters"`
}

type ComparisonInput struct {
	ReferenceVideoID     string          `json:"referenceVideoId"`
	TestVideoID          string          `json:"testVideoId"`
	EnableAdvancedMetrics bool           `json:"enableAdvancedMetrics"`
	CustomParameters     json.RawMessage `json:"customParameters"`
}

type ReportGenerationInput struct {
	VideoAnalysisID string       `json:"videoAnalysisId"`
	ReportType      ReportType   `json:"reportType"`
	Format          ReportFormat `json:"format"`
	IncludeGraphs   bool         `json:"includeGraphs"`
	CustomTemplate  *string      `json:"customTemplate"`
}

type UserFilter struct {
	TenantID      *string    `json:"tenantId"`
	Roles         []string   `json:"roles"`
	IsActive      *bool      `json:"isActive"`
	CreatedAfter  *time.Time `json:"createdAfter"`
	CreatedBefore *time.Time `json:"createdBefore"`
}

type VideoAnalysisFilter struct {
	UserID              *string           `json:"userId"`
	TenantID            *string           `json:"tenantId"`
	Status              []AnalysisStatus  `json:"status"`
	CreatedAfter        *time.Time        `json:"createdAfter"`
	CreatedBefore       *time.Time        `json:"createdBefore"`
	MinDuration         *float64          `json:"minDuration"`
	MaxDuration         *float64          `json:"maxDuration"`
	Format              []string          `json:"format"`
	HasQualityMetrics   *bool             `json:"hasQualityMetrics"`
	HasContentAnalysis  *bool             `json:"hasContentAnalysis"`
}

type PaginationInput struct {
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
	SortBy    string    `json:"sortBy"`
	SortOrder SortOrder `json:"sortOrder"`
}

// Response Types
type PaginatedVideoAnalyses struct {
	Items       []*VideoAnalysis `json:"items"`
	TotalCount  int              `json:"totalCount"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
	HasNextPage bool             `json:"hasNextPage"`
	HasPrevPage bool             `json:"hasPrevPage"`
}

type PaginatedUsers struct {
	Items       []*User `json:"items"`
	TotalCount  int     `json:"totalCount"`
	Page        int     `json:"page"`
	Limit       int     `json:"limit"`
	HasNextPage bool    `json:"hasNextPage"`
	HasPrevPage bool    `json:"hasPrevPage"`
}

type PaginatedComparisons struct {
	Items       []*VideoComparison `json:"items"`
	TotalCount  int                `json:"totalCount"`
	Page        int                `json:"page"`
	Limit       int                `json:"limit"`
	HasNextPage bool               `json:"hasNextPage"`
	HasPrevPage bool               `json:"hasPrevPage"`
}

// Analytics Types
type AnalyticsOverview struct {
	TotalAnalyses          int               `json:"totalAnalyses"`
	AnalysesThisMonth      int               `json:"analysesThisMonth"`
	AverageProcessingTime  float64           `json:"averageProcessingTime"`
	TotalUsersActive       int               `json:"totalUsersActive"`
	PopularFormats         []*FormatUsage    `json:"popularFormats"`
	QualityDistribution    *QualityDistribution `json:"qualityDistribution"`
	UsageByTenant          []*TenantUsage    `json:"usageByTenant"`
}

type FormatUsage struct {
	Format     string  `json:"format"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type QualityDistribution struct {
	Excellent int `json:"excellent"` // VMAF > 90
	Good      int `json:"good"`      // VMAF 70-90
	Fair      int `json:"fair"`      // VMAF 50-70
	Poor      int `json:"poor"`      // VMAF < 50
}

type TenantUsage struct {
	TenantID           string `json:"tenantId"`
	AnalysisCount      int    `json:"analysisCount"`
	StorageUsed        int    `json:"storageUsed"`
	APICallsThisMonth  int    `json:"apiCallsThisMonth"`
}

// Enums
type AnalysisStatus string

const (
	AnalysisStatusPending    AnalysisStatus = "PENDING"
	AnalysisStatusProcessing AnalysisStatus = "PROCESSING"
	AnalysisStatusCompleted  AnalysisStatus = "COMPLETED"
	AnalysisStatusFailed     AnalysisStatus = "FAILED"
	AnalysisStatusCancelled  AnalysisStatus = "CANCELLED"
)

type StreamType string

const (
	StreamTypeVideo    StreamType = "VIDEO"
	StreamTypeAudio    StreamType = "AUDIO"
	StreamTypeSubtitle StreamType = "SUBTITLE"
	StreamTypeData     StreamType = "DATA"
)

type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "ACTIVE"
	APIKeyStatusRotating APIKeyStatus = "ROTATING"
	APIKeyStatusExpired  APIKeyStatus = "EXPIRED"
	APIKeyStatusRevoked  APIKeyStatus = "REVOKED"
)

type ComparisonStatus string

const (
	ComparisonStatusPending    ComparisonStatus = "PENDING"
	ComparisonStatusProcessing ComparisonStatus = "PROCESSING"
	ComparisonStatusCompleted  ComparisonStatus = "COMPLETED"
	ComparisonStatusFailed     ComparisonStatus = "FAILED"
)

type ReportType string

const (
	ReportTypeBasic           ReportType = "BASIC"
	ReportTypeDetailed        ReportType = "DETAILED"
	ReportTypeQualityAssessment ReportType = "QUALITY_ASSESSMENT"
	ReportTypeComplianceCheck ReportType = "COMPLIANCE_CHECK"
	ReportTypeComparisonReport ReportType = "COMPARISON_REPORT"
)

type ReportFormat string

const (
	ReportFormatJSON ReportFormat = "JSON"
	ReportFormatHTML ReportFormat = "HTML"
	ReportFormatPDF  ReportFormat = "PDF"
	ReportFormatCSV  ReportFormat = "CSV"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// Interface definitions for resolver pattern
type QueryResolver interface {
	VideoAnalysis(ctx context.Context, id string) (*VideoAnalysis, error)
	VideoAnalyses(ctx context.Context, filter *VideoAnalysisFilter, pagination *PaginationInput) (*PaginatedVideoAnalyses, error)
	Me(ctx context.Context) (*User, error)
	User(ctx context.Context, id string) (*User, error)
	Users(ctx context.Context, filter *UserFilter, pagination *PaginationInput) (*PaginatedUsers, error)
	MyAPIKeys(ctx context.Context) ([]*APIKey, error)
	APIKey(ctx context.Context, id string) (*APIKey, error)
	MyRateLimits(ctx context.Context) (*UserRateLimit, error)
	TenantRateLimits(ctx context.Context, tenantID string) (*TenantRateLimit, error)
	VideoComparison(ctx context.Context, id string) (*VideoComparison, error)
	VideoComparisons(ctx context.Context, filter *VideoAnalysisFilter, pagination *PaginationInput) (*PaginatedComparisons, error)
	AnalysisReport(ctx context.Context, id string) (*AnalysisReport, error)
	AnalysisReports(ctx context.Context, videoAnalysisID string) ([]*AnalysisReport, error)
	AnalyticsOverview(ctx context.Context, tenantID *string, startDate *time.Time, endDate *time.Time) (*AnalyticsOverview, error)
	SearchVideoAnalyses(ctx context.Context, query string, limit *int) ([]*VideoAnalysis, error)
}

type MutationResolver interface {
	CreateVideoAnalysis(ctx context.Context, input VideoAnalysisInput) (*VideoAnalysis, error)
	RetryVideoAnalysis(ctx context.Context, id string) (*VideoAnalysis, error)
	CancelVideoAnalysis(ctx context.Context, id string) (*VideoAnalysis, error)
	DeleteVideoAnalysis(ctx context.Context, id string) (bool, error)
	CreateVideoComparison(ctx context.Context, input ComparisonInput) (*VideoComparison, error)
	CancelVideoComparison(ctx context.Context, id string) (*VideoComparison, error)
	DeleteVideoComparison(ctx context.Context, id string) (bool, error)
	GenerateAnalysisReport(ctx context.Context, input ReportGenerationInput) (*AnalysisReport, error)
	DeleteAnalysisReport(ctx context.Context, id string) (bool, error)
	CreateAPIKey(ctx context.Context, name string, permissions []string) (*APIKey, error)
	RotateAPIKey(ctx context.Context, id string) (*APIKey, error)
	RevokeAPIKey(ctx context.Context, id string) (bool, error)
	UpdateAPIKeyRateLimits(ctx context.Context, id string, perMinute int, perHour int, perDay int) (*APIKey, error)
	RotateJwtSecret(ctx context.Context) (bool, error)
	UpdateUserRateLimits(ctx context.Context, userID string, perMinute int, perHour int, perDay int) (*UserRateLimit, error)
	UpdateTenantRateLimits(ctx context.Context, tenantID string, perMinute int, perHour int, perDay int) (*TenantRateLimit, error)
	CleanupExpiredCredentials(ctx context.Context) (bool, error)
}

type SubscriptionResolver interface {
	VideoAnalysisProgress(ctx context.Context, id string) (<-chan *VideoAnalysis, error)
	VideoComparisonProgress(ctx context.Context, id string) (<-chan *VideoComparison, error)
	UserNotifications(ctx context.Context) (<-chan json.RawMessage, error)
	SystemStatus(ctx context.Context) (<-chan json.RawMessage, error)
}

type VideoAnalysisResolver interface {
	Format(ctx context.Context, obj *VideoAnalysis) (*VideoFormat, error)
	Streams(ctx context.Context, obj *VideoAnalysis) ([]*Stream, error)
	QualityMetrics(ctx context.Context, obj *VideoAnalysis) (*QualityMetrics, error)
	ContentAnalysis(ctx context.Context, obj *VideoAnalysis) (*ContentAnalysis, error)
	HlsInfo(ctx context.Context, obj *VideoAnalysis) (*HLSInfo, error)
}

type UserResolver interface {
	APIKeys(ctx context.Context, obj *User) ([]*APIKey, error)
	VideoAnalyses(ctx context.Context, obj *User) ([]*VideoAnalysis, error)
	RateLimits(ctx context.Context, obj *User) (*UserRateLimit, error)
}

type APIKeyResolver interface {
	RateLimits(ctx context.Context, obj *APIKey) (*APIKeyRateLimit, error)
}

type VideoComparisonResolver interface {
	ReferenceVideo(ctx context.Context, obj *VideoComparison) (*VideoAnalysis, error)
	TestVideo(ctx context.Context, obj *VideoComparison) (*VideoAnalysis, error)
	ComparisonMetrics(ctx context.Context, obj *VideoComparison) (*ComparisonMetrics, error)
}