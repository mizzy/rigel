package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTool struct {
	mock.Mock
	BaseTool
}

func (m *MockTool) Execute(ctx context.Context, input string) (string, error) {
	args := m.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func TestBaseTool(t *testing.T) {
	tests := []struct {
		name            string
		toolName        string
		toolDescription string
		expectedName    string
		expectedDesc    string
	}{
		{
			name:            "basic tool properties",
			toolName:        "code-formatter",
			toolDescription: "Formats code according to language conventions",
			expectedName:    "code-formatter",
			expectedDesc:    "Formats code according to language conventions",
		},
		{
			name:            "empty properties",
			toolName:        "",
			toolDescription: "",
			expectedName:    "",
			expectedDesc:    "",
		},
		{
			name:            "special characters in properties",
			toolName:        "test-tool-123",
			toolDescription: "A tool that handles: testing, validation & verification!",
			expectedName:    "test-tool-123",
			expectedDesc:    "A tool that handles: testing, validation & verification!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &BaseTool{
				name:        tt.toolName,
				description: tt.toolDescription,
			}

			assert.Equal(t, tt.expectedName, tool.Name())
			assert.Equal(t, tt.expectedDesc, tool.Description())
		})
	}
}

func TestToolInterface(t *testing.T) {
	mockTool := new(MockTool)
	mockTool.BaseTool = BaseTool{
		name:        "test-tool",
		description: "A tool for testing",
	}

	t.Run("tool interface implementation", func(t *testing.T) {
		var _ Tool = mockTool

		assert.Equal(t, "test-tool", mockTool.Name())
		assert.Equal(t, "A tool for testing", mockTool.Description())

		ctx := context.Background()
		input := "test input"
		expectedOutput := "test output"

		mockTool.On("Execute", ctx, input).Return(expectedOutput, nil).Once()

		output, err := mockTool.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output)
		mockTool.AssertExpectations(t)
	})

	t.Run("execute with error", func(t *testing.T) {
		ctx := context.Background()
		input := "error input"

		mockTool.On("Execute", ctx, input).Return("", assert.AnError).Once()

		output, err := mockTool.Execute(ctx, input)

		assert.Error(t, err)
		assert.Empty(t, output)
		mockTool.AssertExpectations(t)
	})
}

type CustomTool struct {
	BaseTool
	executeFunc func(ctx context.Context, input string) (string, error)
}

func (c *CustomTool) Execute(ctx context.Context, input string) (string, error) {
	if c.executeFunc != nil {
		return c.executeFunc(ctx, input)
	}
	return "", nil
}

func TestCustomToolImplementation(t *testing.T) {
	t.Run("custom tool with execute function", func(t *testing.T) {
		customTool := &CustomTool{
			BaseTool: BaseTool{
				name:        "custom-processor",
				description: "Processes custom inputs",
			},
			executeFunc: func(ctx context.Context, input string) (string, error) {
				return "Processed: " + input, nil
			},
		}

		ctx := context.Background()
		result, err := customTool.Execute(ctx, "test data")

		assert.NoError(t, err)
		assert.Equal(t, "Processed: test data", result)
		assert.Equal(t, "custom-processor", customTool.Name())
		assert.Equal(t, "Processes custom inputs", customTool.Description())
	})

	t.Run("custom tool without execute function", func(t *testing.T) {
		customTool := &CustomTool{
			BaseTool: BaseTool{
				name:        "empty-tool",
				description: "Does nothing",
			},
		}

		ctx := context.Background()
		result, err := customTool.Execute(ctx, "any input")

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}
