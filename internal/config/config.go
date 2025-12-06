package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port     int    `json:"port"`
	Host     string `json:"host"`
	BaseURL  string `json:"base_url"`
	LogLevel string `json:"log_level"`

	// Database configuration
	DatabaseType string `json:"database_type"` // sqlite only
	DatabaseURL  string `json:"database_url"`
	DatabasePath string `json:"database_path"` // for SQLite

	// Valkey configuration (Redis-compatible)
	ValkeyHost     string `json:"valkey_host"`
	ValkeyPort     int    `json:"valkey_port"`
	ValkeyPassword string `json:"valkey_password"`
	ValkeyDB       int    `json:"valkey_db"`

	// API configuration
	APIKey string `json:"api_key"`

	// Authentication configuration
	JWTSecret       string `json:"jwt_secret"`
	TokenExpiry     int    `json:"token_expiry_hours"`   // hours
	RefreshExpiry   int    `json:"refresh_expiry_hours"` // hours
	EnableAuth      bool   `json:"enable_auth"`
	EnableRateLimit bool   `json:"enable_rate_limit"`

	// Rate limiting configuration
	RateLimitPerMinute int `json:"rate_limit_per_minute"`
	RateLimitPerHour   int `json:"rate_limit_per_hour"`
	RateLimitPerDay    int `json:"rate_limit_per_day"`

	// Security configuration
	EnableCSRF     bool     `json:"enable_csrf"`
	AllowedOrigins []string `json:"allowed_origins"`
	TrustedProxies []string `json:"trusted_proxies"`

	// FFmpeg configuration
	FFmpegPath  string `json:"ffmpeg_path"`
	FFprobePath string `json:"ffprobe_path"`

	// Upload configuration
	UploadDir   string `json:"upload_dir"`
	MaxFileSize int64  `json:"max_file_size"`

	// Reports configuration
	ReportsDir string `json:"reports_dir"`

	// LLM configuration (required by default)
	LLMModelPath        string `json:"llm_model_path"`
	LLMTimeout          int    `json:"llm_timeout"` // seconds, default 120
	OpenRouterAPIKey    string `json:"openrouter_api_key"`
	EnableLocalLLM      bool   `json:"enable_local_llm"`
	OllamaURL           string `json:"ollama_url"`
	OllamaModel         string `json:"ollama_model"`
	OllamaFallbackModel string `json:"ollama_fallback_model"`
	RequireLLM          bool   `json:"require_llm"` // Make LLM mandatory for analysis

	// Cloud deployment configuration
	CloudMode          bool `json:"cloud_mode"`           // Enable cloud deployment mode
	SkipAuthValidation bool `json:"skip_auth_validation"` // Skip strict auth validation in cloud mode

	// Circuit breaker configuration
	EnableCircuitBreaker   bool `json:"enable_circuit_breaker"`   // Enable circuit breaker protection
	CircuitBreakerTimeout  int  `json:"circuit_breaker_timeout"`  // Timeout in seconds before half-open
	CircuitBreakerInterval int  `json:"circuit_breaker_interval"` // Interval in seconds to reset counters

	// Cloud storage configuration (optional)
	StorageProvider     string `json:"storage_provider"`
	StorageBucket       string `json:"storage_bucket"`
	StorageRegion       string `json:"storage_region"`
	StorageAccessKey    string `json:"storage_access_key"`
	StorageSecretKey    string `json:"storage_secret_key"`
	StorageEndpoint     string `json:"storage_endpoint"`
	StorageUseSSL       bool   `json:"storage_use_ssl"`
	StorageBaseURL      string `json:"storage_base_url"`
	AWSAccessKeyID      string `json:"aws_access_key_id"`
	AWSSecretAccessKey  string `json:"aws_secret_access_key"`
	AWSRegion           string `json:"aws_region"`
	GCPServiceAccount   string `json:"gcp_service_account_json"`
	AzureStorageAccount string `json:"azure_storage_account"`
	AzureStorageKey     string `json:"azure_storage_key"`
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		// Default values
		Port:                   getEnvAsInt("API_PORT", 8080),
		Host:                   getEnv("API_HOST", "localhost"),
		BaseURL:                getEnv("BASE_URL", ""),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		DatabaseType:           getEnv("DB_TYPE", "sqlite"),
		DatabasePath:           getEnv("DB_PATH", "./data/ffprobe.db"),
		ValkeyHost:             getEnv("VALKEY_HOST", "localhost"),
		ValkeyPort:             getEnvAsInt("VALKEY_PORT", 6379),
		ValkeyPassword:         getEnv("VALKEY_PASSWORD", ""),
		ValkeyDB:               getEnvAsInt("VALKEY_DB", 0),
		APIKey:                 getEnv("API_KEY", ""),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		TokenExpiry:            getEnvAsInt("TOKEN_EXPIRY_HOURS", 24),
		RefreshExpiry:          getEnvAsInt("REFRESH_EXPIRY_HOURS", 168), // 7 days
		EnableAuth:             getEnvAsBool("ENABLE_AUTH", true),
		EnableRateLimit:        getEnvAsBool("ENABLE_RATE_LIMIT", true),
		RateLimitPerMinute:     getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		RateLimitPerHour:       getEnvAsInt("RATE_LIMIT_PER_HOUR", 1000),
		RateLimitPerDay:        getEnvAsInt("RATE_LIMIT_PER_DAY", 10000),
		EnableCSRF:             getEnvAsBool("ENABLE_CSRF", false),
		AllowedOrigins:         getEnvAsStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		TrustedProxies:         getEnvAsStringSlice("TRUSTED_PROXIES", []string{}),
		FFmpegPath:             getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:            getEnv("FFPROBE_PATH", "ffprobe"),
		UploadDir:              getEnv("UPLOAD_DIR", "/tmp/uploads"),
		MaxFileSize:            getEnvAsInt64("MAX_FILE_SIZE", 50*1024*1024*1024), // 50GB default
		ReportsDir:             getEnv("REPORTS_DIR", "/tmp/reports"),
		LLMModelPath:           getEnv("LLM_MODEL_PATH", ""),
		OpenRouterAPIKey:       getEnv("OPENROUTER_API_KEY", ""),
		EnableLocalLLM:         getEnvAsBool("ENABLE_LOCAL_LLM", true),
		OllamaURL:              getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:            getEnv("OLLAMA_MODEL", "gemma3:270m"),
		OllamaFallbackModel:    getEnv("OLLAMA_FALLBACK_MODEL", "phi3:mini"),
		RequireLLM:             getEnvAsBool("REQUIRE_LLM", true), // LLM is mandatory by default
		CloudMode:              getEnvAsBool("CLOUD_MODE", false), // Detect cloud deployment
		SkipAuthValidation:     getEnvAsBool("SKIP_AUTH_VALIDATION", false),
		EnableCircuitBreaker:   getEnvAsBool("ENABLE_CIRCUIT_BREAKER", true),
		CircuitBreakerTimeout:  getEnvAsInt("CIRCUIT_BREAKER_TIMEOUT", 30),
		CircuitBreakerInterval: getEnvAsInt("CIRCUIT_BREAKER_INTERVAL", 60),
		StorageProvider:        getEnv("STORAGE_PROVIDER", "local"),
		StorageBucket:          getEnv("STORAGE_BUCKET", "./storage"),
		StorageRegion:          getEnv("STORAGE_REGION", "us-east-1"),
		StorageAccessKey:       getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:       getEnv("STORAGE_SECRET_KEY", ""),
		StorageEndpoint:        getEnv("STORAGE_ENDPOINT", ""),
		StorageUseSSL:          getEnvAsBool("STORAGE_USE_SSL", true),
		StorageBaseURL:         getEnv("STORAGE_BASE_URL", ""),
		AWSAccessKeyID:         getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:     getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:              getEnv("AWS_REGION", "us-east-1"),
		GCPServiceAccount:      getEnv("GCP_SERVICE_ACCOUNT_JSON", ""),
		AzureStorageAccount:    getEnv("AZURE_STORAGE_ACCOUNT", ""),
		AzureStorageKey:        getEnv("AZURE_STORAGE_KEY", ""),
	}

	// Build database URL if not provided directly
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = buildDatabaseURL(cfg)
	}

	// Build base URL if not provided directly
	if cfg.BaseURL == "" {
		cfg.BaseURL = buildBaseURL(cfg)
	}

	// Validate critical configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
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

// buildDatabaseURL constructs a database connection URL
func buildDatabaseURL(cfg *Config) string {
	if cfg.DatabaseType != "sqlite" {
		// Force SQLite
		cfg.DatabaseType = "sqlite"
	}
	return fmt.Sprintf("sqlite3://%s", cfg.DatabasePath)
}

// buildBaseURL constructs the base URL for the API
func buildBaseURL(cfg *Config) string {
	protocol := "http"
	if cfg.Port == 443 {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, cfg.Host, cfg.Port)
}

// validateConfig validates critical configuration values
func validateConfig(cfg *Config) error {
	var errors []string

	// Skip strict auth validation in cloud mode or when explicitly disabled
	if !cfg.CloudMode && !cfg.SkipAuthValidation {
		// Validate required security settings for production
		if cfg.APIKey == "" {
			errors = append(errors, "API_KEY is required for authentication")
		} else if len(cfg.APIKey) < 32 {
			errors = append(errors, "API_KEY must be at least 32 characters long")
		}

		if cfg.JWTSecret == "" {
			errors = append(errors, "JWT_SECRET is required for authentication")
		} else if cfg.JWTSecret == "your-super-secret-jwt-key-change-in-production" {
			errors = append(errors, "JWT_SECRET must be changed from default value")
		} else if len(cfg.JWTSecret) < 32 {
			errors = append(errors, "JWT_SECRET must be at least 32 characters long")
		}
	} else {
		// In cloud mode, provide reasonable defaults if not set
		if cfg.APIKey == "" {
			cfg.APIKey = "cloud-default-key-" + generateRandomString(32)
		}
		if cfg.JWTSecret == "your-super-secret-jwt-key-change-in-production" || cfg.JWTSecret == "" {
			cfg.JWTSecret = "cloud-jwt-secret-" + generateRandomString(32)
		}
	}

	// Validate database configuration - only SQLite supported
	if cfg.DatabaseType != "sqlite" {
		errors = append(errors, "DB_TYPE must be 'sqlite' (PostgreSQL no longer supported)")
	}

	if cfg.DatabasePath == "" {
		errors = append(errors, "DB_PATH is required when using SQLite")
	}

	// Validate ports
	if cfg.Port <= 0 || cfg.Port > 65535 {
		errors = append(errors, "API_PORT must be between 1 and 65535")
	}

	// Validate host
	if cfg.Host == "" {
		errors = append(errors, "API_HOST is required")
	}

	// Validate base URL format if provided
	if cfg.BaseURL != "" {
		if !strings.HasPrefix(cfg.BaseURL, "http://") && !strings.HasPrefix(cfg.BaseURL, "https://") {
			errors = append(errors, "BASE_URL must start with http:// or https://")
		}
	}

	// Validate file paths and directories
	if cfg.UploadDir == "" {
		errors = append(errors, "UPLOAD_DIR is required")
	} else {
		if err := validateDirectory(cfg.UploadDir); err != nil {
			errors = append(errors, fmt.Sprintf("UPLOAD_DIR validation failed: %v", err))
		}
	}

	if cfg.ReportsDir == "" {
		errors = append(errors, "REPORTS_DIR is required")
	} else {
		if err := validateDirectory(cfg.ReportsDir); err != nil {
			errors = append(errors, fmt.Sprintf("REPORTS_DIR validation failed: %v", err))
		}
	}

	// Validate file size limits
	if cfg.MaxFileSize <= 0 {
		errors = append(errors, "MAX_FILE_SIZE must be greater than 0")
	}

	// Validate rate limiting
	if cfg.EnableRateLimit {
		if cfg.RateLimitPerMinute <= 0 {
			errors = append(errors, "RATE_LIMIT_PER_MINUTE must be greater than 0 when rate limiting is enabled")
		}
		if cfg.RateLimitPerHour <= 0 {
			errors = append(errors, "RATE_LIMIT_PER_HOUR must be greater than 0 when rate limiting is enabled")
		}
		if cfg.RateLimitPerDay <= 0 {
			errors = append(errors, "RATE_LIMIT_PER_DAY must be greater than 0 when rate limiting is enabled")
		}
	}

	// Validate FFmpeg paths
	if cfg.FFmpegPath == "" {
		errors = append(errors, "FFMPEG_PATH is required")
	}

	if cfg.FFprobePath == "" {
		errors = append(errors, "FFPROBE_PATH is required")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if cfg.LogLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		errors = append(errors, "LOG_LEVEL must be one of: debug, info, warn, error, fatal, panic")
	}

	// Validate LLM configuration - LLM is now mandatory by default
	if cfg.RequireLLM {
		if !cfg.EnableLocalLLM && cfg.OpenRouterAPIKey == "" {
			errors = append(errors, "LLM is required for analysis: either enable local LLM (ENABLE_LOCAL_LLM=true) or provide OpenRouter API key (OPENROUTER_API_KEY)")
		}
	}

	if cfg.EnableLocalLLM {
		if cfg.OllamaURL == "" {
			errors = append(errors, "OLLAMA_URL is required when local LLM is enabled")
		}
		if cfg.OllamaModel == "" {
			errors = append(errors, "OLLAMA_MODEL is required when local LLM is enabled")
		}
	}

	// Validate OpenRouter configuration if API key is provided
	if cfg.OpenRouterAPIKey != "" {
		if len(cfg.OpenRouterAPIKey) < 10 {
			errors = append(errors, "OPENROUTER_API_KEY appears to be invalid (too short)")
		}
	}

	// Validate token expiry values
	if cfg.TokenExpiry <= 0 {
		errors = append(errors, "TOKEN_EXPIRY_HOURS must be greater than 0")
	}
	if cfg.RefreshExpiry <= 0 {
		errors = append(errors, "REFRESH_EXPIRY_HOURS must be greater than 0")
	}
	if cfg.TokenExpiry >= cfg.RefreshExpiry {
		errors = append(errors, "REFRESH_EXPIRY_HOURS must be greater than TOKEN_EXPIRY_HOURS")
	}

	// Validate Valkey configuration if used for rate limiting
	if cfg.EnableRateLimit {
		if cfg.ValkeyPort <= 0 || cfg.ValkeyPort > 65535 {
			errors = append(errors, "VALKEY_PORT must be between 1 and 65535")
		}
		if cfg.ValkeyHost == "" {
			errors = append(errors, "VALKEY_HOST is required when rate limiting is enabled")
		}
	}

	// Validate CORS configuration
	if len(cfg.AllowedOrigins) > 0 {
		for _, origin := range cfg.AllowedOrigins {
			if origin != "*" && !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
				errors = append(errors, fmt.Sprintf("invalid CORS origin format: %s (must start with http:// or https:// or be '*')", origin))
			}
		}
	}

	// Validate enhanced storage configuration
	if cfg.StorageProvider != "local" {
		switch cfg.StorageProvider {
		case "s3":
			if cfg.AWSAccessKeyID == "" {
				errors = append(errors, "AWS_ACCESS_KEY_ID is required when using S3 storage")
			}
			if cfg.AWSSecretAccessKey == "" {
				errors = append(errors, "AWS_SECRET_ACCESS_KEY is required when using S3 storage")
			}
			if cfg.StorageBucket == "" {
				errors = append(errors, "STORAGE_BUCKET is required when using S3 storage")
			}
		case "gcs":
			if cfg.GCPServiceAccount == "" {
				errors = append(errors, "GCP_SERVICE_ACCOUNT_JSON is required when using GCS storage")
			}
			if cfg.StorageBucket == "" {
				errors = append(errors, "STORAGE_BUCKET is required when using GCS storage")
			}
		case "azure":
			if cfg.AzureStorageAccount == "" {
				errors = append(errors, "AZURE_STORAGE_ACCOUNT is required when using Azure storage")
			}
			if cfg.AzureStorageKey == "" {
				errors = append(errors, "AZURE_STORAGE_KEY is required when using Azure storage")
			}
		case "smb", "cifs":
			if cfg.StorageEndpoint == "" {
				errors = append(errors, "STORAGE_ENDPOINT (SMB share path) is required when using SMB/CIFS storage")
			}
		case "nfs":
			if cfg.StorageEndpoint == "" {
				errors = append(errors, "STORAGE_ENDPOINT (NFS mount path) is required when using NFS storage")
			}
		case "ftp":
			if cfg.StorageEndpoint == "" {
				errors = append(errors, "STORAGE_ENDPOINT (FTP server) is required when using FTP storage")
			}
		case "nas":
			if cfg.StorageEndpoint == "" {
				errors = append(errors, "STORAGE_ENDPOINT (NAS path) is required when using NAS storage")
			}
		default:
			errors = append(errors, "STORAGE_PROVIDER must be one of: local, s3, gcs, azure, smb, cifs, nfs, ftp, nas")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation errors:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// validateDirectory checks if a directory exists or can be created
func validateDirectory(dir string) error {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if directory exists
	if stat, err := os.Stat(absDir); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", absDir)
		}
		// Directory exists, check if writable
		testFile := filepath.Join(absDir, ".write_test")
		if f, err := os.Create(testFile); err != nil {
			return fmt.Errorf("directory is not writable: %s", absDir)
		} else {
			f.Close()
			os.Remove(testFile)
		}
		return nil
	} else if os.IsNotExist(err) {
		// Directory doesn't exist, try to create it
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	} else {
		return fmt.Errorf("failed to check directory: %w", err)
	}
}

// generateRandomString generates a random hex string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	_, _ = rand.Read(bytes) // Error intentionally ignored - crypto/rand is the fallback
	return hex.EncodeToString(bytes)
}
