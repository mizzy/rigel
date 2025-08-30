package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	args := m.Called(ctx, prompt, opts)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) GenerateWithHistory(ctx context.Context, messages []Message, opts GenerateOptions) (string, error) {
	args := m.Called(ctx, messages, opts)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) Stream(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	args := m.Called(ctx, prompt)
	if ch, ok := args.Get(0).(<-chan StreamResponse); ok {
		return ch, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProvider) ListModels(ctx context.Context) ([]Model, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Model), args.Error(1)
}

func (m *MockProvider) GetCurrentModel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) SetModel(model string) {
	m.Called(model)
}

func TestGenerateOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     GenerateOptions
		expected GenerateOptions
	}{
		{
			name: "default options",
			opts: GenerateOptions{},
			expected: GenerateOptions{
				Temperature:  0,
				MaxTokens:    0,
				SystemPrompt: "",
			},
		},
		{
			name: "custom options",
			opts: GenerateOptions{
				Temperature:  0.8,
				MaxTokens:    2048,
				SystemPrompt: "You are a helpful assistant",
			},
			expected: GenerateOptions{
				Temperature:  0.8,
				MaxTokens:    2048,
				SystemPrompt: "You are a helpful assistant",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.Temperature, tt.opts.Temperature)
			assert.Equal(t, tt.expected.MaxTokens, tt.opts.MaxTokens)
			assert.Equal(t, tt.expected.SystemPrompt, tt.opts.SystemPrompt)
		})
	}
}

func TestProviderInterface(t *testing.T) {
	mockProvider := new(MockProvider)
	ctx := context.Background()

	t.Run("Generate method", func(t *testing.T) {
		prompt := "Test prompt"
		expectedResponse := "Test response"

		mockProvider.On("Generate", ctx, prompt).Return(expectedResponse, nil).Once()

		response, err := mockProvider.Generate(ctx, prompt)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		mockProvider.AssertExpectations(t)
	})

	t.Run("GenerateWithOptions method", func(t *testing.T) {
		prompt := "Test prompt with options"
		opts := GenerateOptions{
			Temperature:  0.7,
			MaxTokens:    1000,
			SystemPrompt: "System prompt",
		}
		expectedResponse := "Response with options"

		mockProvider.On("GenerateWithOptions", ctx, prompt, opts).Return(expectedResponse, nil).Once()

		response, err := mockProvider.GenerateWithOptions(ctx, prompt, opts)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		mockProvider.AssertExpectations(t)
	})
}
