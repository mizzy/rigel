package llm

import (
	"context"
	"testing"
)

// TestConversationProvider implements Provider interface for testing conversation history
type TestConversationProvider struct {
	generateHistoryCalled bool
	messagesReceived      []Message
	responseToReturn      string
	errorToReturn         error
}

func (m *TestConversationProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return m.responseToReturn, m.errorToReturn
}

func (m *TestConversationProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	return m.responseToReturn, m.errorToReturn
}

func (m *TestConversationProvider) GenerateWithHistory(ctx context.Context, messages []Message, opts GenerateOptions) (string, error) {
	m.generateHistoryCalled = true
	m.messagesReceived = messages
	return m.responseToReturn, m.errorToReturn
}

func (m *TestConversationProvider) Stream(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse, 1)
	ch <- StreamResponse{Content: m.responseToReturn, Done: true}
	close(ch)
	return ch, nil
}

func (m *TestConversationProvider) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{{Name: "test-model"}}, nil
}

func (m *TestConversationProvider) GetCurrentModel() string {
	return "test-model"
}

func (m *TestConversationProvider) SetModel(model string) {
	// no-op for mock
}

func TestGenerateWithHistory(t *testing.T) {
	tests := []struct {
		name             string
		messages         []Message
		expectedMessages int
	}{
		{
			name: "single user message",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			expectedMessages: 1,
		},
		{
			name: "conversation with history",
			messages: []Message{
				{Role: "user", Content: "What is 2+2?"},
				{Role: "assistant", Content: "2+2 equals 4."},
				{Role: "user", Content: "What about 3+3?"},
			},
			expectedMessages: 3,
		},
		{
			name: "longer conversation",
			messages: []Message{
				{Role: "user", Content: "Tell me about Go"},
				{Role: "assistant", Content: "Go is a programming language."},
				{Role: "user", Content: "What are its main features?"},
				{Role: "assistant", Content: "Go has goroutines, channels, and a simple syntax."},
				{Role: "user", Content: "How about error handling?"},
			},
			expectedMessages: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &TestConversationProvider{
				responseToReturn: "test response",
			}

			ctx := context.Background()
			_, err := mock.GenerateWithHistory(ctx, tt.messages, GenerateOptions{})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !mock.generateHistoryCalled {
				t.Error("GenerateWithHistory was not called")
			}

			if len(mock.messagesReceived) != tt.expectedMessages {
				t.Errorf("expected %d messages, got %d", tt.expectedMessages, len(mock.messagesReceived))
			}

			// Verify message order and content
			for i, msg := range tt.messages {
				if i < len(mock.messagesReceived) {
					if mock.messagesReceived[i].Role != msg.Role {
						t.Errorf("message %d: expected role %s, got %s", i, msg.Role, mock.messagesReceived[i].Role)
					}
					if mock.messagesReceived[i].Content != msg.Content {
						t.Errorf("message %d: expected content %s, got %s", i, msg.Content, mock.messagesReceived[i].Content)
					}
				}
			}
		})
	}
}
