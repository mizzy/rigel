package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Provider        string
	AnthropicAPIKey string
	OpenAIAPIKey    string
	GoogleAPIKey    string
	AzureAPIKey     string
	OllamaBaseURL   string
	Model           string
	LogLevel        string
}

func Load(configFile string) (*Config, error) {
	if configFile == "" {
		configFile = ".env"
	}

	if _, err := os.Stat(configFile); err == nil {
		if err := godotenv.Load(configFile); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	viper.SetEnvPrefix("RIGEL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		Provider:        getEnv("PROVIDER", "ollama"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		GoogleAPIKey:    os.Getenv("GOOGLE_API_KEY"),
		AzureAPIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
		OllamaBaseURL:   getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		Model:           getEnv("MODEL", ""),
		LogLevel:        getEnv("RIGEL_LOG_LEVEL", "info"),
	}

	if cfg.Provider == "anthropic" && cfg.Model == "" {
		cfg.Model = "claude-3-5-sonnet-20241022"
	} else if cfg.Provider == "openai" && cfg.Model == "" {
		cfg.Model = "gpt-4-turbo-preview"
	} else if cfg.Provider == "ollama" && cfg.Model == "" {
		cfg.Model = "gpt-oss:20b"
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	switch c.Provider {
	case "anthropic":
		if c.AnthropicAPIKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required for Anthropic provider")
		}
	case "openai":
		if c.OpenAIAPIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required for OpenAI provider")
		}
	case "google":
		if c.GoogleAPIKey == "" {
			return fmt.Errorf("GOOGLE_API_KEY is required for Google provider")
		}
	case "azure":
		if c.AzureAPIKey == "" {
			return fmt.Errorf("AZURE_OPENAI_API_KEY is required for Azure provider")
		}
	case "ollama":
		// Ollama doesn't require API key, just base URL which has a default
	default:
		return fmt.Errorf("unsupported provider: %s", c.Provider)
	}
	return nil
}
