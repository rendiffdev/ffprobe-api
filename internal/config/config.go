package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port     int    `json:"port"`
	LogLevel string `json:"log_level"`

	// Database configuration
	DatabaseURL      string `json:"database_url"`
	DatabaseHost     string `json:"database_host"`
	DatabasePort     int    `json:"database_port"`
	DatabaseName     string `json:"database_name"`
	DatabaseUser     string `json:"database_user"`
	DatabasePassword string `json:"database_password"`
	DatabaseSSLMode  string `json:"database_ssl_mode"`

	// Redis configuration
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// API configuration
	APIKey string `json:"api_key"`

	// Authentication configuration
	JWTSecret         string `json:"jwt_secret"`
	TokenExpiry       int    `json:"token_expiry_hours"`    // hours
	RefreshExpiry     int    `json:"refresh_expiry_hours"`  // hours
	EnableAuth        bool   `json:"enable_auth"`
	EnableRateLimit   bool   `json:"enable_rate_limit"`

	// Rate limiting configuration
	RateLimitPerMinute int `json:"rate_limit_per_minute"`
	RateLimitPerHour   int `json:"rate_limit_per_hour"`
	RateLimitPerDay    int `json:"rate_limit_per_day"`

	// Security configuration
	EnableCSRF       bool     `json:"enable_csrf"`
	AllowedOrigins   []string `json:"allowed_origins"`
	TrustedProxies   []string `json:"trusted_proxies"`

	// FFmpeg configuration
	FFmpegPath  string `json:"ffmpeg_path"`
	FFprobePath string `json:"ffprobe_path"`

	// Upload configuration
	UploadDir     string `json:"upload_dir"`
	MaxFileSize   int64  `json:"max_file_size"`
	
	// Reports configuration
	ReportsDir    string `json:"reports_dir"`

	// LLM configuration (optional)
	LLMModelPath     string `json:"llm_model_path"`
	OpenRouterAPIKey string `json:"openrouter_api_key"`

	// Cloud storage configuration (optional)
	StorageProvider        string `json:"storage_provider"`
	StorageBucket          string `json:"storage_bucket"`
	StorageRegion          string `json:"storage_region"`
	StorageAccessKey       string `json:"storage_access_key"`
	StorageSecretKey       string `json:"storage_secret_key"`
	StorageEndpoint        string `json:"storage_endpoint"`
	StorageUseSSL          bool   `json:"storage_use_ssl"`
	StorageBaseURL         string `json:"storage_base_url"`
	AWSAccessKeyID         string `json:"aws_access_key_id"`
	AWSSecretAccessKey     string `json:"aws_secret_access_key"`
	AWSRegion              string `json:"aws_region"`
	GCPServiceAccount      string `json:"gcp_service_account_json"`
	AzureStorageAccount    string `json:"azure_storage_account"`
	AzureStorageKey        string `json:"azure_storage_key"`
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		// Default values
		Port:             getEnvAsInt("API_PORT", 8080),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		DatabaseHost:     getEnv("POSTGRES_HOST", "localhost"),
		DatabasePort:     getEnvAsInt("POSTGRES_PORT", 5432),
		DatabaseName:     getEnv("POSTGRES_DB", "ffprobe_api"),
		DatabaseUser:     getEnv("POSTGRES_USER", "postgres"),
		DatabasePassword: getEnv("POSTGRES_PASSWORD", ""),
		DatabaseSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvAsInt("REDIS_DB", 0),
		APIKey:             getEnv("API_KEY", ""),
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		TokenExpiry:        getEnvAsInt("TOKEN_EXPIRY_HOURS", 24),
		RefreshExpiry:      getEnvAsInt("REFRESH_EXPIRY_HOURS", 168), // 7 days
		EnableAuth:         getEnvAsBool("ENABLE_AUTH", true),
		EnableRateLimit:    getEnvAsBool("ENABLE_RATE_LIMIT", true),
		RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		RateLimitPerHour:   getEnvAsInt("RATE_LIMIT_PER_HOUR", 1000),
		RateLimitPerDay:    getEnvAsInt("RATE_LIMIT_PER_DAY", 10000),
		EnableCSRF:         getEnvAsBool("ENABLE_CSRF", false),
		AllowedOrigins:     getEnvAsStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		TrustedProxies:     getEnvAsStringSlice("TRUSTED_PROXIES", []string{}),
		FFmpegPath:         getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:        getEnv("FFPROBE_PATH", "ffprobe"),
		UploadDir:          getEnv("UPLOAD_DIR", "/tmp/uploads"),
		MaxFileSize:        getEnvAsInt64("MAX_FILE_SIZE", 50*1024*1024*1024), // 50GB default
		ReportsDir:         getEnv("REPORTS_DIR", "/tmp/reports"),
		LLMModelPath:       getEnv("LLM_MODEL_PATH", ""),
		OpenRouterAPIKey:   getEnv("OPENROUTER_API_KEY", ""),
		StorageProvider:    getEnv("STORAGE_PROVIDER", "local"),
		StorageBucket:      getEnv("STORAGE_BUCKET", "./storage"),
		StorageRegion:      getEnv("STORAGE_REGION", "us-east-1"),
		StorageAccessKey:   getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:   getEnv("STORAGE_SECRET_KEY", ""),
		StorageEndpoint:    getEnv("STORAGE_ENDPOINT", ""),
		StorageUseSSL:      getEnvAsBool("STORAGE_USE_SSL", true),
		StorageBaseURL:     getEnv("STORAGE_BASE_URL", ""),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		GCPServiceAccount:  getEnv("GCP_SERVICE_ACCOUNT_JSON", ""),
		AzureStorageAccount: getEnv("AZURE_STORAGE_ACCOUNT", ""),
		AzureStorageKey:     getEnv("AZURE_STORAGE_KEY", ""),
	}

	// Build database URL if not provided directly
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = buildDatabaseURL(cfg)
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsInt64 gets an environment variable as int64 with a fallback value
func getEnvAsInt64(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

// getEnvAsStringSlice gets an environment variable as string slice with a fallback value
func getEnvAsStringSlice(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

// buildDatabaseURL constructs a PostgreSQL connection URL
func buildDatabaseURL(cfg *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)
}