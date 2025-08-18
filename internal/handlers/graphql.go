package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/rendiffdev/ffprobe-api/internal/cache"
	"github.com/rs/zerolog"
	gql "github.com/rendiffdev/ffprobe-api/internal/graphql"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// GraphQLHandler manages GraphQL endpoint
type GraphQLHandler struct {
	server *handler.Server
	logger zerolog.Logger
}

// GraphQLConfig holds configuration for GraphQL server
type GraphQLConfig struct {
	EnablePlayground     bool
	EnableIntrospection  bool
	EnableQueryComplexity bool
	MaxQueryComplexity   int
	MaxQueryDepth        int
	QueryCacheSize       int
	EnableTracing        bool
	EnableAPQ            bool // Automatic Persisted Queries
}

// NewGraphQLHandler creates a new GraphQL handler
func NewGraphQLHandler(
	db *sqlx.DB,
	redisClient interface{},
	logger zerolog.Logger,
	analysisService *services.AnalysisService,
	comparisonService *services.ComparisonService,
	reportService *services.ReportService,
	rotationService *services.SecretRotationService,
	userService *services.UserService,
	storageService *services.StorageService,
	config GraphQLConfig,
) *GraphQLHandler {
	// Create resolver
	resolver := gql.NewResolver(
		db, redisClient, logger,
		analysisService, comparisonService, reportService,
		rotationService, userService, storageService,
	)

	// Create GraphQL server
	srv := handler.New(NewExecutableSchema(Config{Resolvers: resolver}))

	// Configure transports
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Configure CORS for WebSocket connections
				return true // In production, implement proper origin checking
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Set defaults if not configured
	if config.MaxQueryComplexity == 0 {
		config.MaxQueryComplexity = 1000
	}
	if config.MaxQueryDepth == 0 {
		config.MaxQueryDepth = 15
	}
	if config.QueryCacheSize == 0 {
		config.QueryCacheSize = 1000
	}

	// Configure query complexity analysis
	if config.EnableQueryComplexity {
		srv.Use(extension.FixedComplexityLimit(config.MaxQueryComplexity))
	}

	// Configure introspection
	if config.EnableIntrospection {
		srv.Use(extension.Introspection{})
	}

	// Configure automatic persisted queries
	if config.EnableAPQ {
		srv.Use(extension.AutoPersistedQuery{
			Cache: lru.New(100),
		})
	}

	// Configure query caching
	srv.SetQueryCache(lru.New(config.QueryCacheSize))

	// Add error presenter
	srv.SetErrorPresenter(func(ctx context.Context, e error) *graphql.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)
		
		// Log errors for debugging
		logger.Error().
			Err(e).
			Str("path", err.Path.String()).
			Interface("locations", err.Locations).
			Msg("GraphQL error")

		// Don't expose internal errors in production
		if err.Message == "internal system error" {
			err.Message = "An internal error occurred"
		}

		return err
	})

	// Add recovery handler
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		logger.Error().
			Interface("panic", err).
			Msg("GraphQL panic recovered")
		return graphql.ErrorOnPath(ctx, "Internal server error")
	})

	return &GraphQLHandler{
		server: srv,
		logger: logger,
	}
}

// GraphQLMiddleware adds authentication and context to GraphQL requests
func (h *GraphQLHandler) GraphQLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract authentication information
		userID := c.GetString("user_id")
		tenantID := c.GetString("tenant_id")
		roles := c.GetStringSlice("roles")
		apiKeyID := c.GetString("api_key_id")

		// Create context with authentication info
		ctx := c.Request.Context()
		if userID != "" {
			ctx = context.WithValue(ctx, "user_id", userID)
		}
		if tenantID != "" {
			ctx = context.WithValue(ctx, "tenant_id", tenantID)
		}
		if len(roles) > 0 {
			ctx = context.WithValue(ctx, "roles", roles)
		}
		if apiKeyID != "" {
			ctx = context.WithValue(ctx, "api_key_id", apiKeyID)
		}

		// Add request ID for tracing
		if requestID := c.GetString("request_id"); requestID != "" {
			ctx = context.WithValue(ctx, "request_id", requestID)
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// GraphQLEndpoint handles GraphQL queries and mutations
func (h *GraphQLHandler) GraphQLEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.server.ServeHTTP(c.Writer, c.Request)
	}
}

// GraphQLPlaygroundHandler serves the GraphQL playground interface
func (h *GraphQLHandler) GraphQLPlaygroundHandler() gin.HandlerFunc {
	playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
	return func(c *gin.Context) {
		playgroundHandler.ServeHTTP(c.Writer, c.Request)
	}
}

// GraphQLSchemaHandler returns the GraphQL schema as SDL
func (h *GraphQLHandler) GraphQLSchemaHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		schema := h.server.Schema()
		c.Header("Content-Type", "application/graphql")
		c.String(http.StatusOK, schema.String())
	}
}

// Custom directives and field resolvers can be added here

// Example: Authentication directive
type AuthDirective struct{}

func (a AuthDirective) Validate(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
	userRoles, ok := ctx.Value("roles").([]string)
	if !ok || len(userRoles) == 0 {
		return nil, graphql.NewResponseError("Authentication required")
	}

	// Check if user has required roles
	if len(roles) > 0 {
		hasRole := false
		for _, requiredRole := range roles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		if !hasRole {
			return nil, graphql.NewResponseError("Insufficient permissions")
		}
	}

	return next(ctx)
}

// Rate limiting directive
type RateLimitDirective struct {
	cache  cache.Client
	logger zerolog.Logger
}

func (r RateLimitDirective) Validate(ctx context.Context, obj interface{}, next graphql.Resolver, maxRequests int, window time.Duration) (interface{}, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return next(ctx) // Skip rate limiting for unauthenticated requests
	}

	// Get operation name for rate limiting key
	reqCtx := graphql.GetOperationContext(ctx)
	operation := "unknown"
	if reqCtx.Operation != nil && reqCtx.Operation.Name != "" {
		operation = reqCtx.Operation.Name
	}

	key := fmt.Sprintf("graphql_rate_limit:%s:%s", userID, operation)
	
	// Check current count
	count, err := r.cache.Incr(ctx, key)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to increment rate limit counter")
		return next(ctx) // Continue on cache error
	}

	// Set expiration on first request
	if count == 1 {
		r.cache.Expire(ctx, key, window)
	}

	if count > int64(maxRequests) {
		return nil, graphql.NewResponseError(fmt.Sprintf("Rate limit exceeded: %d requests per %v", maxRequests, window))
	}

	return next(ctx)
}

// Complexity analysis for expensive operations
func VideoAnalysisComplexity(childComplexity int, filter *gql.VideoAnalysisFilter, pagination *gql.PaginationInput) int {
	complexity := childComplexity

	// Base complexity for pagination
	if pagination != nil && pagination.Limit > 0 {
		complexity += pagination.Limit * 2
	} else {
		complexity += 40 // Default limit of 20 * 2
	}

	// Add complexity for filters
	if filter != nil {
		if filter.HasQualityMetrics != nil && *filter.HasQualityMetrics {
			complexity += 50
		}
		if filter.HasContentAnalysis != nil && *filter.HasContentAnalysis {
			complexity += 30
		}
		if len(filter.Status) > 0 {
			complexity += len(filter.Status) * 2
		}
	}

	return complexity
}

// Middleware to add operation context for logging and metrics
func (h *GraphQLHandler) OperationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse GraphQL request to extract operation info
		var req struct {
			Query         string                 `json:"query"`
			Variables     map[string]interface{} `json:"variables"`
			OperationName string                 `json:"operationName"`
		}

		if c.Request.Method == "POST" {
			if err := c.ShouldBindJSON(&req); err == nil {
				// Log GraphQL operation
				h.logger.Info().
					Str("operation", req.OperationName).
					Str("user_id", c.GetString("user_id")).
					Str("tenant_id", c.GetString("tenant_id")).
					Msg("GraphQL operation started")

				// Add operation name to context for metrics
				c.Set("graphql_operation", req.OperationName)

				// Re-encode the body for downstream processing
				body, _ := json.Marshal(req)
				c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
			}
		}

		c.Next()
	}
}

// Helper function to create executable schema (placeholder)
// This would typically be generated by gqlgen
func NewExecutableSchema(cfg Config) graphql.ExecutableSchema {
	// This is a placeholder - in a real implementation,
	// this would be generated by gqlgen based on your schema
	return nil
}

// Config struct for schema configuration
type Config struct {
	Resolvers interface{}
}