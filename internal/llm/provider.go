package llm

import (
	"context"
	"fmt"

	"github.com/mizzy/rigel/internal/config"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
	GenerateWithHistory(ctx context.Context, messages []Message, opts GenerateOptions) (string, error)
	Stream(ctx context.Context, prompt string) (<-chan StreamResponse, error)
	ListModels(ctx context.Context) ([]Model, error)
	GetCurrentModel() string
	SetModel(model string)
}

type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

type GenerateOptions struct {
	Temperature  float32
	MaxTokens    int
	SystemPrompt string
	Model        string
}

type StreamResponse struct {
	Content string
	Error   error
	Done    bool
}

type Model struct {
	Name       string       `json:"name"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	ModifiedAt string       `json:"modified_at,omitempty"`
	Details    ModelDetails `json:"details,omitempty"`
}

type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

func NewProvider(cfg *config.Config) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	switch cfg.Provider {
	case "anthropic":
		return NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.Model)
	case "openai":
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	case "ollama":
		return NewOllamaProvider(cfg.OllamaBaseURL, cfg.Model)
	default:
		if cfg.AnthropicAPIKey != "" {
			return NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.Model)
		}
		return nil, fmt.Errorf("no valid LLM provider configured")
	}
}
