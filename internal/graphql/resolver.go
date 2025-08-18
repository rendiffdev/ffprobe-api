package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// Resolver is the root resolver struct
type Resolver struct {
	db                  *sqlx.DB
	redis               *redis.Client
	logger              zerolog.Logger
	analysisService     *services.AnalysisService
	comparisonService   *services.ComparisonService
	reportService       *services.ReportService
	rotationService     *services.SecretRotationService
	userService         *services.UserService
	storageService      *services.StorageService
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	db *sqlx.DB,
	redis *redis.Client,
	logger zerolog.Logger,
	analysisService *services.AnalysisService,
	comparisonService *services.ComparisonService,
	reportService *services.ReportService,
	rotationService *services.SecretRotationService,
	userService *services.UserService,
	storageService *services.StorageService,
) *Resolver {
	return &Resolver{
		db:                db,
		redis:             redis,
		logger:            logger,
		analysisService:   analysisService,
		comparisonService: comparisonService,
		reportService:     reportService,
		rotationService:   rotationService,
		userService:       userService,
		storageService:    storageService,
	}
}

// Query resolver
type queryResolver struct{ *Resolver }

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation resolver
type mutationResolver struct{ *Resolver }

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Subscription resolver
type subscriptionResolver struct{ *Resolver }

func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

// VideoAnalysis resolver
type videoAnalysisResolver struct{ *Resolver }

func (r *Resolver) VideoAnalysis() VideoAnalysisResolver { return &videoAnalysisResolver{r} }

// User resolver
type userResolver struct{ *Resolver }

func (r *Resolver) User() UserResolver { return &userResolver{r} }

// APIKey resolver
type apiKeyResolver struct{ *Resolver }

func (r *Resolver) APIKey() APIKeyResolver { return &apiKeyResolver{r} }

// VideoComparison resolver
type videoComparisonResolver struct{ *Resolver }

func (r *Resolver) VideoComparison() VideoComparisonResolver { return &videoComparisonResolver{r} }

// Query Implementations
func (r *queryResolver) VideoAnalysis(ctx context.Context, id string) (*VideoAnalysis, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	analysis, err := r.getVideoAnalysisFromDB(ctx, id, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("analysis_id", id).Msg("Failed to get video analysis")
		return nil, fmt.Errorf("failed to get video analysis: %w", err)
	}

	return analysis, nil
}

func (r *queryResolver) VideoAnalyses(ctx context.Context, filter *VideoAnalysisFilter, pagination *PaginationInput) (*PaginatedVideoAnalyses, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	// Set defaults
	if pagination == nil {
		pagination = &PaginationInput{
			Page:      1,
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: SortOrderDesc,
		}
	}

	analyses, totalCount, err := r.getVideoAnalysesFromDB(ctx, userID, filter, pagination)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get video analyses")
		return nil, fmt.Errorf("failed to get video analyses: %w", err)
	}

	hasNextPage := (pagination.Page * pagination.Limit) < totalCount
	hasPrevPage := pagination.Page > 1

	return &PaginatedVideoAnalyses{
		Items:       analyses,
		TotalCount:  totalCount,
		Page:        pagination.Page,
		Limit:       pagination.Limit,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}, nil
}

func (r *queryResolver) Me(ctx context.Context) (*User, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	user, err := r.getUserFromDB(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *queryResolver) MyAPIKeys(ctx context.Context) ([]*APIKey, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	keys, err := r.getAPIKeysFromDB(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get API keys")
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	return keys, nil
}

func (r *queryResolver) MyRateLimits(ctx context.Context) (*UserRateLimit, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	rateLimits, err := r.getUserRateLimitsFromDB(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rate limits")
		return nil, fmt.Errorf("failed to get rate limits: %w", err)
	}

	return rateLimits, nil
}

func (r *queryResolver) AnalyticsOverview(ctx context.Context, tenantID *string, startDate *time.Time, endDate *time.Time) (*AnalyticsOverview, error) {
	// Check admin role or tenant ownership
	if !hasAdminRole(ctx) && (tenantID == nil || *tenantID != getTenantIDFromContext(ctx)) {
		return nil, fmt.Errorf("insufficient permissions")
	}

	overview, err := r.getAnalyticsOverviewFromDB(ctx, tenantID, startDate, endDate)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get analytics overview")
		return nil, fmt.Errorf("failed to get analytics overview: %w", err)
	}

	return overview, nil
}

// Mutation Implementations
func (r *mutationResolver) CreateVideoAnalysis(ctx context.Context, input VideoAnalysisInput) (*VideoAnalysis, error) {
	userID := getUserIDFromContext(ctx)
	tenantID := getTenantIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	// Create analysis record
	analysisID, err := r.createVideoAnalysisInDB(ctx, userID, tenantID, input)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create video analysis")
		return nil, fmt.Errorf("failed to create video analysis: %w", err)
	}

	// Start async analysis
	go func() {
		if err := r.analysisService.ProcessVideo(context.Background(), analysisID, input.FilePath); err != nil {
			r.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Failed to process video")
		}
	}()

	// Return the created analysis
	return r.getVideoAnalysisFromDB(ctx, analysisID, userID)
}

func (r *mutationResolver) CreateAPIKey(ctx context.Context, name string, permissions []string) (*APIKey, error) {
	userID := getUserIDFromContext(ctx)
	tenantID := getTenantIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	// Create API key using rotation service
	apiKey, rawKey, err := r.rotationService.GenerateAPIKey(ctx, userID, tenantID, name, permissions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create API key")
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Convert to GraphQL type
	return &APIKey{
		ID:          apiKey.ID,
		UserID:      apiKey.UserID,
		TenantID:    apiKey.TenantID,
		KeyPrefix:   apiKey.KeyPrefix,
		Name:        apiKey.Name,
		Permissions: apiKey.Permissions,
		Status:      APIKeyStatus(apiKey.Status),
		CreatedAt:   apiKey.CreatedAt,
		ExpiresAt:   apiKey.ExpiresAt,
		LastUsedAt:  &apiKey.LastUsedAt,
		LastRotated: apiKey.LastRotated,
		RotationDue: apiKey.RotationDue,
		UsageCount:  int(apiKey.UsageCount),
		RateLimits: &APIKeyRateLimit{
			PerMinute: apiKey.RateLimitRPM,
			PerHour:   apiKey.RateLimitRPH,
			PerDay:    apiKey.RateLimitRPD,
		},
	}, nil
}

func (r *mutationResolver) RotateAPIKey(ctx context.Context, id string) (*APIKey, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	// Rotate the key
	newKey, _, err := r.rotationService.RotateAPIKey(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("key_id", id).Msg("Failed to rotate API key")
		return nil, fmt.Errorf("failed to rotate API key: %w", err)
	}

	// Convert to GraphQL type
	return &APIKey{
		ID:          newKey.ID,
		UserID:      newKey.UserID,
		TenantID:    newKey.TenantID,
		KeyPrefix:   newKey.KeyPrefix,
		Name:        newKey.Name,
		Permissions: newKey.Permissions,
		Status:      APIKeyStatus(newKey.Status),
		CreatedAt:   newKey.CreatedAt,
		ExpiresAt:   newKey.ExpiresAt,
		LastUsedAt:  &newKey.LastUsedAt,
		LastRotated: newKey.LastRotated,
		RotationDue: newKey.RotationDue,
		UsageCount:  int(newKey.UsageCount),
		RateLimits: &APIKeyRateLimit{
			PerMinute: newKey.RateLimitRPM,
			PerHour:   newKey.RateLimitRPH,
			PerDay:    newKey.RateLimitRPD,
		},
	}, nil
}

// Helper functions to extract context values
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func getTenantIDFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

func hasAdminRole(ctx context.Context) bool {
	if roles, ok := ctx.Value("roles").([]string); ok {
		for _, role := range roles {
			if role == "admin" {
				return true
			}
		}
	}
	return false
}

// Database helper functions (implement based on your existing services)
func (r *Resolver) getVideoAnalysisFromDB(ctx context.Context, id, userID string) (*VideoAnalysis, error) {
	// Implementation would query the database and return a VideoAnalysis
	// This is a placeholder - implement based on your existing database schema
	return nil, fmt.Errorf("not implemented")
}

func (r *Resolver) getVideoAnalysesFromDB(ctx context.Context, userID string, filter *VideoAnalysisFilter, pagination *PaginationInput) ([]*VideoAnalysis, int, error) {
	// Implementation would query the database with filters and pagination
	// This is a placeholder - implement based on your existing database schema
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *Resolver) getUserFromDB(ctx context.Context, userID string) (*User, error) {
	// Implementation would query the users table
	// This is a placeholder - implement based on your existing database schema
	return nil, fmt.Errorf("not implemented")
}

func (r *Resolver) getAPIKeysFromDB(ctx context.Context, userID string) ([]*APIKey, error) {
	// Implementation would query the api_keys table
	// This is a placeholder - implement based on your existing database schema
	return nil, fmt.Errorf("not implemented")
}

func (r *Resolver) getUserRateLimitsFromDB(ctx context.Context, userID string) (*UserRateLimit, error) {
	// Implementation would query the user_rate_limits table
	// This is a placeholder - implement based on your existing database schema
	return nil, fmt.Errorf("not implemented")
}

func (r *Resolver) getAnalyticsOverviewFromDB(ctx context.Context, tenantID *string, startDate, endDate *time.Time) (*AnalyticsOverview, error) {
	// Implementation would aggregate analytics data
	// This is a placeholder - implement based on your analytics requirements
	return nil, fmt.Errorf("not implemented")
}

func (r *Resolver) createVideoAnalysisInDB(ctx context.Context, userID, tenantID string, input VideoAnalysisInput) (string, error) {
	// Implementation would insert a new video analysis record
	// This is a placeholder - implement based on your existing database schema
	return "", fmt.Errorf("not implemented")
}