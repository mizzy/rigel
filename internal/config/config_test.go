package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "actual",
			expected:     "actual",
		},
		{
			name:         "environment variable does not exist",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "empty default value",
			key:          "EMPTY_DEFAULT",
			defaultValue: "",
			envValue:     "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       map[string]string
		configFile     string
		expectedConfig *Config
		createEnvFile  bool
		envFileContent string
	}{
		{
			name:     "default configuration with ollama",
			setupEnv: map[string]string{},
			expectedConfig: &Config{
				Provider:      "ollama",
				OllamaBaseURL: "http://localhost:11434",
				Model:         "gpt-oss:20b",
				LogLevel:      "info",
			},
		},
		{
			name: "anthropic provider configuration",
			setupEnv: map[string]string{
				"PROVIDER":          "anthropic",
				"ANTHROPIC_API_KEY": "test-anthropic-key",
			},
			expectedConfig: &Config{
				Provider:        "anthropic",
				AnthropicAPIKey: "test-anthropic-key",
				Model:           "claude-sonnet-4-20250514",
				LogLevel:        "info",
			},
		},
		{
			name: "openai provider configuration",
			setupEnv: map[string]string{
				"PROVIDER":       "openai",
				"OPENAI_API_KEY": "test-openai-key",
			},
			expectedConfig: &Config{
				Provider:     "openai",
				OpenAIAPIKey: "test-openai-key",
				Model:        "gpt-4-turbo-preview",
				LogLevel:     "info",
			},
		},
		{
			name: "custom model configuration",
			setupEnv: map[string]string{
				"PROVIDER":          "anthropic",
				"ANTHROPIC_API_KEY": "test-key",
				"MODEL":             "claude-3-opus-20240229",
			},
			expectedConfig: &Config{
				Provider:        "anthropic",
				AnthropicAPIKey: "test-key",
				Model:           "claude-3-opus-20240229",
				LogLevel:        "info",
			},
		},
		{
			name: "custom log level",
			setupEnv: map[string]string{
				"PROVIDER":          "anthropic",
				"ANTHROPIC_API_KEY": "test-key",
				"RIGEL_LOG_LEVEL":   "debug",
			},
			expectedConfig: &Config{
				Provider:        "anthropic",
				AnthropicAPIKey: "test-key",
				Model:           "claude-sonnet-4-20250514",
				LogLevel:        "debug",
			},
		},
		{
			name:          "load from .env file",
			createEnvFile: true,
			envFileContent: `PROVIDER=anthropic
ANTHROPIC_API_KEY=env-file-key
MODEL=claude-3-opus-20240229
RIGEL_LOG_LEVEL=warn`,
			expectedConfig: &Config{
				Provider:        "anthropic",
				AnthropicAPIKey: "env-file-key",
				Model:           "claude-3-opus-20240229",
				LogLevel:        "warn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.setupEnv {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			var configFile string
			if tt.createEnvFile {
				tmpDir := t.TempDir()
				configFile = filepath.Join(tmpDir, ".env")
				err := os.WriteFile(configFile, []byte(tt.envFileContent), 0644)
				require.NoError(t, err)
			}

			cfg, err := Load(configFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig.Provider, cfg.Provider)
			if tt.expectedConfig.AnthropicAPIKey != "" {
				assert.Equal(t, tt.expectedConfig.AnthropicAPIKey, cfg.AnthropicAPIKey)
			}
			if tt.expectedConfig.OpenAIAPIKey != "" {
				assert.Equal(t, tt.expectedConfig.OpenAIAPIKey, cfg.OpenAIAPIKey)
			}
			if tt.expectedConfig.OllamaBaseURL != "" {
				assert.Equal(t, tt.expectedConfig.OllamaBaseURL, cfg.OllamaBaseURL)
			}
			assert.Equal(t, tt.expectedConfig.Model, cfg.Model)
			assert.Equal(t, tt.expectedConfig.LogLevel, cfg.LogLevel)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid anthropic config",
			config: &Config{
				Provider:        "anthropic",
				AnthropicAPIKey: "test-key",
			},
			expectError: false,
		},
		{
			name: "missing anthropic API key",
			config: &Config{
				Provider: "anthropic",
			},
			expectError: true,
			errorMsg:    "ANTHROPIC_API_KEY is required",
		},
		{
			name: "valid openai config",
			config: &Config{
				Provider:     "openai",
				OpenAIAPIKey: "test-key",
			},
			expectError: false,
		},
		{
			name: "missing openai API key",
			config: &Config{
				Provider: "openai",
			},
			expectError: true,
			errorMsg:    "OPENAI_API_KEY is required",
		},
		{
			name: "valid google config",
			config: &Config{
				Provider:     "google",
				GoogleAPIKey: "test-key",
			},
			expectError: false,
		},
		{
			name: "missing google API key",
			config: &Config{
				Provider: "google",
			},
			expectError: true,
			errorMsg:    "GOOGLE_API_KEY is required",
		},
		{
			name: "valid azure config",
			config: &Config{
				Provider:    "azure",
				AzureAPIKey: "test-key",
			},
			expectError: false,
		},
		{
			name: "missing azure API key",
			config: &Config{
				Provider: "azure",
			},
			expectError: true,
			errorMsg:    "AZURE_OPENAI_API_KEY is required",
		},
		{
			name: "valid ollama config",
			config: &Config{
				Provider:      "ollama",
				OllamaBaseURL: "http://localhost:11434",
			},
			expectError: false,
		},
		{
			name: "unsupported provider",
			config: &Config{
				Provider: "unsupported",
			},
			expectError: true,
			errorMsg:    "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultipleProviderKeys(t *testing.T) {
	setupEnv := map[string]string{
		"PROVIDER":             "anthropic",
		"ANTHROPIC_API_KEY":    "anthropic-key",
		"OPENAI_API_KEY":       "openai-key",
		"GOOGLE_API_KEY":       "google-key",
		"AZURE_OPENAI_API_KEY": "azure-key",
	}

	for key, value := range setupEnv {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "anthropic", cfg.Provider)
	assert.Equal(t, "anthropic-key", cfg.AnthropicAPIKey)
	assert.Equal(t, "openai-key", cfg.OpenAIAPIKey)
	assert.Equal(t, "google-key", cfg.GoogleAPIKey)
	assert.Equal(t, "azure-key", cfg.AzureAPIKey)

	err = cfg.Validate()
	assert.NoError(t, err)
}

func TestDefaultOllamaConfiguration(t *testing.T) {
	// Ensure no environment variables interfere with the test
	envVars := []string{
		"PROVIDER",
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"GOOGLE_API_KEY",
		"AZURE_OPENAI_API_KEY",
		"OLLAMA_BASE_URL",
		"MODEL",
		"RIGEL_LOG_LEVEL",
	}

	for _, v := range envVars {
		if val := os.Getenv(v); val != "" {
			os.Unsetenv(v)
			defer os.Setenv(v, val)
		}
	}

	// Use a temp directory to avoid loading any existing .env file
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "ollama", cfg.Provider)
	assert.Equal(t, "gpt-oss:20b", cfg.Model)
	assert.Equal(t, "http://localhost:11434", cfg.OllamaBaseURL)
	assert.Equal(t, "info", cfg.LogLevel)

	err = cfg.Validate()
	assert.NoError(t, err)
}
