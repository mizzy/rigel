package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mizzy/rigel/internal/agent"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/tools"
)

type TestTool struct {
	tools.BaseTool
}

func (t *TestTool) Execute(ctx context.Context, input string) (string, error) {
	return "Executed: " + input, nil
}

func TestAgentIntegration(t *testing.T) {
	t.Run("Agent with mock provider and tools", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			generateResponse: "This is a test response from the mock provider",
		}

		agent := agent.New(mockProvider)

		testTool := &TestTool{
			BaseTool: tools.BaseTool{},
		}
		agent.RegisterTool(testTool)

		ctx := context.Background()
		response, err := agent.Execute(ctx, "Test task")

		require.NoError(t, err)
		assert.Equal(t, "This is a test response from the mock provider", response)

		history, ok := agent.GetContext("history")
		assert.False(t, ok)
		assert.Nil(t, history)

		agent.SetContext("test-key", "test-value")
		value, ok := agent.GetContext("test-key")
		assert.True(t, ok)
		assert.Equal(t, "test-value", value)
	})

	t.Run("Agent conversation flow", func(t *testing.T) {
		responses := []string{
			"First response",
			"Second response based on history",
			"Third response concluding the conversation",
		}

		mockProvider := &MockLLMProvider{
			multiResponse: responses,
			callCount:     0,
		}

		agent := agent.New(mockProvider)
		ctx := context.Background()

		resp1, err := agent.Execute(ctx, "First question")
		require.NoError(t, err)
		assert.Equal(t, "First response", resp1)

		resp2, err := agent.Execute(ctx, "Follow-up question")
		require.NoError(t, err)
		assert.Equal(t, "Second response based on history", resp2)

		resp3, err := agent.Execute(ctx, "Final question")
		require.NoError(t, err)
		assert.Equal(t, "Third response concluding the conversation", resp3)

		agent.ClearMemory()

		resp4, err := agent.Execute(ctx, "New conversation")
		require.NoError(t, err)
		assert.Equal(t, "First response", resp4)
	})
}

func TestConfigIntegration(t *testing.T) {
	t.Run("Load config with environment variables", func(t *testing.T) {
		os.Setenv("PROVIDER", "anthropic")
		os.Setenv("ANTHROPIC_API_KEY", "test-integration-key")
		os.Setenv("MODEL", "claude-3-opus-20240229")
		os.Setenv("RIGEL_LOG_LEVEL", "debug")
		defer func() {
			os.Unsetenv("PROVIDER")
			os.Unsetenv("ANTHROPIC_API_KEY")
			os.Unsetenv("MODEL")
			os.Unsetenv("RIGEL_LOG_LEVEL")
		}()

		cfg, err := config.Load("")
		require.NoError(t, err)

		assert.Equal(t, "anthropic", cfg.Provider)
		assert.Equal(t, "test-integration-key", cfg.AnthropicAPIKey)
		assert.Equal(t, "claude-3-opus-20240229", cfg.Model)
		assert.Equal(t, "debug", cfg.LogLevel)

		err = cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("Config validation with provider", func(t *testing.T) {
		testCases := []struct {
			provider string
			apiKey   string
			valid    bool
		}{
			{"anthropic", "key", true},
			{"anthropic", "", false},
			{"openai", "key", true},
			{"openai", "", false},
			{"invalid", "key", false},
		}

		for _, tc := range testCases {
			cfg := &config.Config{
				Provider: tc.provider,
			}

			switch tc.provider {
			case "anthropic":
				cfg.AnthropicAPIKey = tc.apiKey
			case "openai":
				cfg.OpenAIAPIKey = tc.apiKey
			}

			err := cfg.Validate()
			if tc.valid {
				assert.NoError(t, err, "Expected valid config for provider %s with key %s", tc.provider, tc.apiKey)
			} else {
				assert.Error(t, err, "Expected invalid config for provider %s with key %s", tc.provider, tc.apiKey)
			}
		}
	})
}

func TestEndToEndWorkflow(t *testing.T) {
	t.Run("Complete workflow simulation", func(t *testing.T) {
		os.Setenv("PROVIDER", "anthropic")
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer func() {
			os.Unsetenv("PROVIDER")
			os.Unsetenv("ANTHROPIC_API_KEY")
		}()

		cfg, err := config.Load("")
		require.NoError(t, err)

		err = cfg.Validate()
		require.NoError(t, err)

		mockProvider := &MockLLMProvider{
			generateResponse: "Generated code: func Add(a, b int) int { return a + b }",
		}

		agent := agent.New(mockProvider)

		codeTool := &TestTool{
			BaseTool: tools.BaseTool{},
		}
		agent.RegisterTool(codeTool)

		agent.SetContext("language", "Go")
		agent.SetContext("project", "test-project")

		ctx := context.Background()
		response, err := agent.Execute(ctx, "Generate an add function")
		require.NoError(t, err)
		assert.Contains(t, response, "func Add")

		lang, ok := agent.GetContext("language")
		assert.True(t, ok)
		assert.Equal(t, "Go", lang)

		response2, err := agent.Execute(ctx, "Add error handling")
		require.NoError(t, err)
		assert.NotEmpty(t, response2)

		agent.ClearMemory()

		_, ok = agent.GetContext("language")
		assert.False(t, ok)
	})
}

type MockLLMProvider struct {
	generateResponse string
	multiResponse    []string
	callCount        int
}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return m.GenerateWithOptions(ctx, prompt, llm.GenerateOptions{})
}

func (m *MockLLMProvider) GenerateWithOptions(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if len(m.multiResponse) > 0 {
		response := m.multiResponse[m.callCount%len(m.multiResponse)]
		m.callCount++
		return response, nil
	}
	return m.generateResponse, nil
}

func (m *MockLLMProvider) Stream(ctx context.Context, prompt string) (<-chan llm.StreamResponse, error) {
	ch := make(chan llm.StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- llm.StreamResponse{
			Content: m.generateResponse,
			Done:    true,
		}
	}()
	return ch, nil
}

func (m *MockLLMProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{
		{Name: "test-model-1"},
		{Name: "test-model-2"},
	}, nil
}

func (m *MockLLMProvider) GetCurrentModel() string {
	return "test-model-1"
}

func (m *MockLLMProvider) SetModel(model string) {
	// Mock implementation - no-op
}
