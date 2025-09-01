package command

import "testing"

func TestCompletionHandler_UpdateCompletions_SpaceHandling(t *testing.T) {
	handler := NewCompletionHandler()

	tests := []struct {
		name        string
		input       string
		expectShow  bool
		expectCount int
	}{
		{
			name:        "command without space",
			input:       "/he",
			expectShow:  true,
			expectCount: 1, // /help
		},
		{
			name:        "command with leading space",
			input:       " /he",
			expectShow:  false, // Should not show completions
			expectCount: 0,
		},
		{
			name:        "command with multiple leading spaces",
			input:       "  /he",
			expectShow:  false, // Should not show completions
			expectCount: 0,
		},
		{
			name:        "text with slash in middle",
			input:       "hello/he",
			expectShow:  false,
			expectCount: 0,
		},
		{
			name:        "just slash",
			input:       "/",
			expectShow:  true,
			expectCount: len(AvailableCommands), // All commands should match
		},
		{
			name:        "space then slash",
			input:       " /",
			expectShow:  false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, showCompletions := handler.UpdateCompletions(tt.input)

			if showCompletions != tt.expectShow {
				t.Errorf("UpdateCompletions(%q) showCompletions = %v, want %v", tt.input, showCompletions, tt.expectShow)
			}

			if len(completions) != tt.expectCount {
				t.Errorf("UpdateCompletions(%q) completions count = %d, want %d", tt.input, len(completions), tt.expectCount)
			}
		})
	}
}
