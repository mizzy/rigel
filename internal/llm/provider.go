package llm

import (
	"context"
	"fmt"

	"github.com/mizzy/rigel/internal/config"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
	Stream(ctx context.Context, prompt string) (<-chan StreamResponse, error)
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

func NewProvider(cfg *config.Config) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	switch cfg.Provider {
	case "anthropic":
		return NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.Model)
	case "openai":
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	default:
		if cfg.AnthropicAPIKey != "" {
			return NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.Model)
		}
		return nil, fmt.Errorf("no valid LLM provider configured")
	}
}
