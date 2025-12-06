package config

import (
	"os"
	"testing"
)

// Helper to set environment variables for tests
func setTestEnv(t *testing.T, envVars map[string]string) func() {
	t.Helper()
	originalValues := make(map[string]string)

	for key, value := range envVars {
		originalValues[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	return func() {
		for key := range envVars {
			if original, exists := originalValues[key]; exists && original != "" {
				os.Setenv(key, original)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_UNSET_VAR",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_SET_VAR",
			defaultValue: "default_value",
			envValue:     "env_value",
			expected:     "env_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				cleanup := setTestEnv(t, map[string]string{tt.key: tt.envValue})
				defer cleanup()
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s; want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_INT_UNSET",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
		{
			name:         "returns parsed int when valid",
			key:          "TEST_INT_VALID",
			defaultValue: 100,
			envValue:     "42",
			expected:     42,
		},
		{
			name:         "returns default when invalid int",
			key:          "TEST_INT_INVALID",
			defaultValue: 100,
			envValue:     "not_a_number",
			expected:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				cleanup := setTestEnv(t, map[string]string{tt.key: tt.envValue})
				defer cleanup()
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsInt(%s, %d) = %d; want %d", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_BOOL_UNSET",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "returns true for 'true'",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "returns true for '1'",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "returns false for 'false'",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "returns false for '0'",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				cleanup := setTestEnv(t, map[string]string{tt.key: tt.envValue})
				defer cleanup()
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsBool(%s, %v) = %v; want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

// createValidConfig creates a valid config with all required fields
func createValidConfig() *Config {
	return &Config{
		Port:               8080,
		Host:               "localhost",
		LogLevel:           "info",
		DatabaseType:       "sqlite",
		DatabasePath:       "/tmp/test.db",
		ValkeyHost:         "localhost",
		ValkeyPort:         6379,
		ValkeyPassword:     "",
		ValkeyDB:           0,
		APIKey:             "valid-api-key-that-is-at-least-32-characters-long",
		JWTSecret:          "valid-jwt-secret-that-is-at-least-32-characters-long",
		TokenExpiry:        24,
		RefreshExpiry:      168,
		EnableAuth:         true,
		EnableRateLimit:    true,
		RateLimitPerMinute: 60,
		RateLimitPerHour:   1000,
		RateLimitPerDay:    10000,
		FFmpegPath:         "ffmpeg",
		FFprobePath:        "ffprobe",
		UploadDir:          "/tmp/uploads",
		ReportsDir:         "/tmp/reports",
		MaxFileSize:        1024,
		EnableLocalLLM:     true,
		OllamaURL:          "http://localhost:11434",
		OllamaModel:        "gemma3:270m",
		RequireLLM:         true,
		StorageProvider:    "local",
		CloudMode:          false,
		SkipAuthValidation: false,
	}
}

func TestValidateConfig_JWTSecret(t *testing.T) {
	tests := []struct {
		name        string
		jwtSecret   string
		cloudMode   bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty JWT secret fails in non-cloud mode",
			jwtSecret:   "",
			cloudMode:   false,
			expectError: true,
			errorMsg:    "JWT_SECRET is required",
		},
		{
			name:        "default JWT secret fails in non-cloud mode",
			jwtSecret:   "your-super-secret-jwt-key-change-in-production",
			cloudMode:   false,
			expectError: true,
			errorMsg:    "JWT_SECRET must be changed from default",
		},
		{
			name:        "short JWT secret fails in non-cloud mode",
			jwtSecret:   "short",
			cloudMode:   false,
			expectError: true,
			errorMsg:    "JWT_SECRET must be at least 32 characters",
		},
		{
			name:        "valid JWT secret passes in non-cloud mode",
			jwtSecret:   "this-is-a-valid-jwt-secret-key-that-is-long-enough",
			cloudMode:   false,
			expectError: false,
		},
		{
			name:        "empty JWT secret auto-generates in cloud mode",
			jwtSecret:   "",
			cloudMode:   true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.JWTSecret = tt.jwtSecret
			cfg.CloudMode = tt.cloudMode

			err := validateConfig(cfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestValidateConfig_APIKey(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		cloudMode   bool
		expectError bool
	}{
		{
			name:        "empty API key fails in non-cloud mode",
			apiKey:      "",
			cloudMode:   false,
			expectError: true,
		},
		{
			name:        "short API key fails in non-cloud mode",
			apiKey:      "short",
			cloudMode:   false,
			expectError: true,
		},
		{
			name:        "valid API key passes",
			apiKey:      "valid-api-key-that-is-at-least-32-characters-long",
			cloudMode:   false,
			expectError: false,
		},
		{
			name:        "empty API key auto-generates in cloud mode",
			apiKey:      "",
			cloudMode:   true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.APIKey = tt.apiKey
			cfg.CloudMode = tt.cloudMode

			err := validateConfig(cfg)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateConfig_DatabaseType(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		expectError  bool
	}{
		{
			name:         "sqlite passes",
			databaseType: "sqlite",
			expectError:  false,
		},
		{
			name:         "postgres fails",
			databaseType: "postgres",
			expectError:  true,
		},
		{
			name:         "mysql fails",
			databaseType: "mysql",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.DatabaseType = tt.databaseType

			err := validateConfig(cfg)

			if tt.expectError && err == nil {
				t.Error("expected error for non-sqlite database, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateConfig_Port(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{
			name:        "valid port 8080",
			port:        8080,
			expectError: false,
		},
		{
			name:        "valid port 443",
			port:        443,
			expectError: false,
		},
		{
			name:        "invalid port 0",
			port:        0,
			expectError: true,
		},
		{
			name:        "invalid port negative",
			port:        -1,
			expectError: true,
		},
		{
			name:        "invalid port too high",
			port:        65536,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.Port = tt.port

			err := validateConfig(cfg)

			if tt.expectError && err == nil {
				t.Errorf("expected error for port %d, got nil", tt.port)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for port %d, got %v", tt.port, err)
			}
		})
	}
}

func TestValidateConfig_LogLevel(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range validLevels {
		t.Run("valid_"+level, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.LogLevel = level

			err := validateConfig(cfg)
			if err != nil {
				t.Errorf("expected no error for log level %s, got %v", level, err)
			}
		})
	}

	t.Run("invalid_log_level", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.LogLevel = "invalid"

		err := validateConfig(cfg)
		if err == nil {
			t.Error("expected error for invalid log level, got nil")
		}
	})
}

func TestGenerateRandomString(t *testing.T) {
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		t.Run("length_"+string(rune(length)), func(t *testing.T) {
			result := generateRandomString(length)

			if len(result) != length {
				t.Errorf("generateRandomString(%d) returned string of length %d", length, len(result))
			}

			// Verify it's hex encoded (only hex chars)
			for _, c := range result {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("generateRandomString returned non-hex character: %c", c)
				}
			}
		})
	}

	// Test uniqueness
	t.Run("uniqueness", func(t *testing.T) {
		s1 := generateRandomString(32)
		s2 := generateRandomString(32)

		if s1 == s2 {
			t.Error("generateRandomString returned same value twice")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
