package render

import (
	"strings"
)

// InputPrompt renders the input prompt with proper alignment
func InputPrompt(inputView string) string {
	var s strings.Builder

	s.WriteString(promptSymbol)
	s.WriteString(" ")

	// Handle multi-line alignment by replacing newlines with proper indentation
	lines := strings.Split(inputView, "\n")
	for i, line := range lines {
		if i > 0 {
			s.WriteString("\n  ") // 2 spaces to align with prompt symbol + space
		}
		s.WriteString(line)
	}

	return s.String()
}
