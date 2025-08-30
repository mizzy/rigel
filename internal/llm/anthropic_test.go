package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnthropicProvider(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		model         string
		expectedModel string
		expectedError bool
	}{
		{
			name:          "valid with custom model",
			apiKey:        "test-api-key",
			model:         "claude-3-opus-20240229",
			expectedModel: "claude-3-opus-20240229",
			expectedError: false,
		},
		{
			name:          "valid with default model",
			apiKey:        "test-api-key",
			model:         "",
			expectedModel: "claude-3-5-sonnet-20241022",
			expectedError: false,
		},
		{
			name:          "missing API key",
			apiKey:        "",
			model:         "claude-3-opus-20240229",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewAnthropicProvider(tt.apiKey, tt.model)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				assert.Contains(t, err.Error(), "API key is required")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, provider)
				assert.NotNil(t, provider.client)
				assert.Equal(t, tt.expectedModel, provider.model)
			}
		})
	}
}

func TestAnthropicProvider_Generate(t *testing.T) {
	t.Skip("Skipping integration test that requires real API key")
}

func TestAnthropicProvider_GenerateWithOptions(t *testing.T) {
	apiKey := "test-api-key"
	provider, err := NewAnthropicProvider(apiKey, "claude-3-5-sonnet-20241022")
	require.NoError(t, err)

	ctx := context.Background()
	prompt := "Test prompt"

	tests := []struct {
		name string
		opts GenerateOptions
	}{
		{
			name: "with default options",
			opts: GenerateOptions{},
		},
		{
			name: "with custom temperature",
			opts: GenerateOptions{
				Temperature: 0.8,
			},
		},
		{
			name: "with custom max tokens",
			opts: GenerateOptions{
				MaxTokens: 2048,
			},
		},
		{
			name: "with system prompt",
			opts: GenerateOptions{
				SystemPrompt: "You are a helpful assistant",
			},
		},
		{
			name: "with all options",
			opts: GenerateOptions{
				Model:        "claude-3-opus-20240229",
				Temperature:  0.9,
				MaxTokens:    1000,
				SystemPrompt: "You are a coding assistant",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Skipping integration test that requires real API key")

			response, err := provider.GenerateWithOptions(ctx, prompt, tt.opts)
			if err != nil {
				t.Logf("Expected error without valid API key: %v", err)
			} else {
				assert.NotEmpty(t, response)
			}
		})
	}
}

func TestAnthropicProvider_Stream(t *testing.T) {
	t.Skip("Skipping integration test that requires real API key")
}
