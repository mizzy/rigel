package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/tools"
)

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) GenerateWithOptions(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	args := m.Called(ctx, prompt, opts)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) GenerateWithHistory(ctx context.Context, messages []llm.Message, opts llm.GenerateOptions) (string, error) {
	args := m.Called(ctx, messages, opts)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) Stream(ctx context.Context, prompt string) (<-chan llm.StreamResponse, error) {
	args := m.Called(ctx, prompt)
	if ch := args.Get(0); ch != nil {
		return ch.(<-chan llm.StreamResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	args := m.Called(ctx)
	if models := args.Get(0); models != nil {
		return models.([]llm.Model), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProvider) GetCurrentModel() llm.Model {
	args := m.Called()
	return args.Get(0).(llm.Model)
}

func (m *MockProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) SetModel(model llm.Model) {
	m.Called(model)
}

type MockTool struct {
	mock.Mock
}

func (m *MockTool) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTool) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTool) Execute(ctx context.Context, input string) (string, error) {
	args := m.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func TestNew(t *testing.T) {
	mockProvider := new(MockProvider)
	agent := New(mockProvider)

	assert.NotNil(t, agent)
	assert.Equal(t, mockProvider, agent.provider)
	assert.NotNil(t, agent.memory)
	assert.Empty(t, agent.memory.conversationHistory)
	assert.NotNil(t, agent.memory.context)
	assert.Empty(t, agent.tools)
}

func TestRegisterTool(t *testing.T) {
	mockProvider := new(MockProvider)
	agent := New(mockProvider)

	mockTool := new(MockTool)
	mockTool.On("Name").Return("test-tool")

	agent.RegisterTool(mockTool)

	assert.Len(t, agent.tools, 1)
	assert.Equal(t, mockTool, agent.tools[0])
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name          string
		task          string
		expectedResp  string
		expectedError error
		setupMock     func(*MockProvider)
	}{
		{
			name:         "successful execution",
			task:         "Write a hello world function",
			expectedResp: "Here's a hello world function in Go:\n```go\nfunc HelloWorld() string {\n    return \"Hello, World!\"\n}\n```",
			setupMock: func(m *MockProvider) {
				// First call for intent analysis
				m.On("Generate", mock.Anything, mock.MatchedBy(func(prompt string) bool {
					return strings.Contains(prompt, "intent analyzer")
				})).Return(`[{"intent":"none","filepath":"","content":""}]`, nil)
				// Second call for actual response
				m.On("GenerateWithOptions", mock.Anything, mock.Anything, mock.Anything).
					Return("Here's a hello world function in Go:\n```go\nfunc HelloWorld() string {\n    return \"Hello, World!\"\n}\n```", nil)
			},
		},
		{
			name:          "execution with error",
			task:          "Invalid task",
			expectedError: assert.AnError,
			setupMock: func(m *MockProvider) {
				// First call for intent analysis
				m.On("Generate", mock.Anything, mock.MatchedBy(func(prompt string) bool {
					return strings.Contains(prompt, "intent analyzer")
				})).Return(`[{"intent":"none","filepath":"","content":""}]`, nil)
				// Second call fails
				m.On("GenerateWithOptions", mock.Anything, mock.Anything, mock.Anything).
					Return("", assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockProvider)
			tt.setupMock(mockProvider)

			agent := New(mockProvider)
			ctx := context.Background()

			resp, err := agent.Execute(ctx, tt.task)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
				assert.Len(t, agent.memory.conversationHistory, 2)
				assert.Equal(t, "user", agent.memory.conversationHistory[0].Role)
				assert.Equal(t, tt.task, agent.memory.conversationHistory[0].Content)
				assert.Equal(t, "assistant", agent.memory.conversationHistory[1].Role)
				assert.Equal(t, tt.expectedResp, agent.memory.conversationHistory[1].Content)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	tests := []struct {
		name           string
		tools          []tools.Tool
		expectedPrompt string
	}{
		{
			name:  "without tools",
			tools: []tools.Tool{},
			expectedPrompt: `You are Rigel, an intelligent AI coding assistant.
You help developers write clean, efficient, and maintainable code.
You can analyze code, suggest improvements, and generate new code based on requirements.
Always follow best practices and coding standards.
Be concise but thorough in your explanations.`,
		},
		{
			name: "with tools",
			tools: func() []tools.Tool {
				mockTool := new(MockTool)
				mockTool.On("Name").Return("code-analyzer")
				mockTool.On("Description").Return("Analyzes code for quality and suggests improvements")
				return []tools.Tool{mockTool}
			}(),
			expectedPrompt: `You are Rigel, an intelligent AI coding assistant.
You help developers write clean, efficient, and maintainable code.
You can analyze code, suggest improvements, and generate new code based on requirements.
Always follow best practices and coding standards.
Be concise but thorough in your explanations.

Available tools:
- code-analyzer: Analyzes code for quality and suggests improvements`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockProvider)
			agent := New(mockProvider)

			for _, tool := range tt.tools {
				agent.RegisterTool(tool)
			}

			prompt := agent.buildSystemPrompt()
			assert.Equal(t, tt.expectedPrompt, prompt)
		})
	}
}

func TestBuildUserPrompt(t *testing.T) {
	tests := []struct {
		name           string
		task           string
		history        []Message
		expectedPrompt string
	}{
		{
			name:           "without history",
			task:           "Write a function to add two numbers",
			history:        []Message{},
			expectedPrompt: "Write a function to add two numbers",
		},
		{
			name: "with history",
			task: "Now make it handle errors",
			history: []Message{
				{Role: "user", Content: "Write a function to add two numbers"},
				{Role: "assistant", Content: "Here's the function: func Add(a, b int) int { return a + b }"},
			},
			expectedPrompt: `Previous conversation:
user: Write a function to add two numbers
assistant: Here's the function: func Add(a, b int) int { return a + b }

Current task: Now make it handle errors`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockProvider)
			agent := New(mockProvider)
			agent.memory.conversationHistory = tt.history

			prompt := agent.buildUserPrompt(tt.task)
			assert.Equal(t, tt.expectedPrompt, prompt)
		})
	}
}

func TestClearMemory(t *testing.T) {
	mockProvider := new(MockProvider)
	agent := New(mockProvider)

	agent.memory.conversationHistory = []Message{
		{Role: "user", Content: "test"},
		{Role: "assistant", Content: "response"},
	}
	agent.memory.context["key"] = "value"

	agent.ClearMemory()

	assert.Empty(t, agent.memory.conversationHistory)
	assert.Empty(t, agent.memory.context)
}

func TestContextManagement(t *testing.T) {
	mockProvider := new(MockProvider)
	agent := New(mockProvider)

	t.Run("SetContext and GetContext", func(t *testing.T) {
		agent.SetContext("project", "rigel")
		agent.SetContext("version", "1.0.0")
		agent.SetContext("features", []string{"code-generation", "analysis"})

		val, ok := agent.GetContext("project")
		assert.True(t, ok)
		assert.Equal(t, "rigel", val)

		val, ok = agent.GetContext("version")
		assert.True(t, ok)
		assert.Equal(t, "1.0.0", val)

		val, ok = agent.GetContext("features")
		assert.True(t, ok)
		assert.Equal(t, []string{"code-generation", "analysis"}, val)

		val, ok = agent.GetContext("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("Overwrite context", func(t *testing.T) {
		agent.SetContext("project", "new-project")

		val, ok := agent.GetContext("project")
		assert.True(t, ok)
		assert.Equal(t, "new-project", val)
	})
}

func TestMemoryPersistenceAcrossExecutions(t *testing.T) {
	mockProvider := new(MockProvider)
	agent := New(mockProvider)

	// Mock intent analysis for first call
	mockProvider.On("Generate", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, "intent analyzer") && strings.Contains(prompt, "First task")
	})).Return(`[{"intent":"none","filepath":"","content":""}]`, nil).Once()

	mockProvider.On("GenerateWithOptions", mock.Anything, "First task", mock.Anything).
		Return("First response", nil).Once()

	ctx := context.Background()
	resp1, err := agent.Execute(ctx, "First task")
	require.NoError(t, err)
	assert.Equal(t, "First response", resp1)

	expectedPrompt := `Previous conversation:
user: First task
assistant: First response

Current task: Second task`

	// Mock intent analysis for second call
	mockProvider.On("Generate", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, "intent analyzer") && strings.Contains(prompt, "Second task")
	})).Return(`[{"intent":"none","filepath":"","content":""}]`, nil).Once()

	mockProvider.On("GenerateWithOptions", mock.Anything, expectedPrompt, mock.Anything).
		Return("Second response", nil).Once()

	resp2, err := agent.Execute(ctx, "Second task")
	require.NoError(t, err)
	assert.Equal(t, "Second response", resp2)

	assert.Len(t, agent.memory.conversationHistory, 4)
	mockProvider.AssertExpectations(t)
}
