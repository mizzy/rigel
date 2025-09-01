package command

import (
	"testing"

	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/state"
)

func TestHandleCommand_SpaceHandling(t *testing.T) {
	llmState := state.NewLLMState()
	chatState := state.NewChatState()
	cfg := &config.Config{}

	tests := []struct {
		name     string
		input    string
		expected string // "request" for regular prompt, "response" for command
	}{
		{
			name:     "command without space",
			input:    "/help",
			expected: "response",
		},
		{
			name:     "command with leading space",
			input:    " /help",
			expected: "request", // Should be treated as regular prompt
		},
		{
			name:     "command with multiple leading spaces",
			input:    "  /help",
			expected: "request", // Should be treated as regular prompt
		},
		{
			name:     "text containing slash",
			input:    "hello/world",
			expected: "request",
		},
		{
			name:     "text with space before slash",
			input:    "hello /world",
			expected: "request",
		},
		{
			name:     "valid command with trailing content",
			input:    "/help me",
			expected: "response", // Unknown command, but still command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleCommand(tt.input, llmState, chatState, cfg, nil, []string{})
			if result.Type != tt.expected {
				t.Errorf("HandleCommand(%q) = %q, want %q", tt.input, result.Type, tt.expected)
			}
		})
	}
}
