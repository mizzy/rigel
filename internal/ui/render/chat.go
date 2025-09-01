package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Exchange represents a single chat exchange
type Exchange struct {
	Prompt   string
	Response string
}

var (
	promptSymbol    = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true).Render("✦")
	promptStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)
	inputStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))
	outputStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	thinkingStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Italic(true)
	suggestionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	highlightStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)
)

// GetTerminalWidth returns the terminal width or a default value
func GetTerminalWidth() int {
	width, _, err := term.GetSize(0)
	if err != nil || width < 40 {
		return 80 // Default fallback width
	}
	if width > 120 {
		width = 120 // Max width for better readability
	}
	return width
}

// ChatHistory renders the conversation history
func ChatHistory(history []Exchange) string {
	if len(history) == 0 {
		return ""
	}

	var s strings.Builder

	// Get terminal width and calculate usable widths
	termWidth := GetTerminalWidth()
	promptWidth := termWidth - 3 // Account for prompt symbol and space
	responseWidth := termWidth - 2

	for _, ex := range history {
		// User prompt with > symbol
		s.WriteString(promptSymbol)
		s.WriteString(" ")

		// Use lipgloss Width() for proper wrapping
		promptStyle := inputStyle.Width(promptWidth)
		s.WriteString(promptStyle.Render(ex.Prompt))
		s.WriteString("\n\n")

		// Assistant response with wrapping
		responseStyle := outputStyle.Width(responseWidth)
		s.WriteString(responseStyle.Render(ex.Response))
		s.WriteString("\n\n")
	}

	return s.String()
}

// ThinkingState renders the thinking indicator
func ThinkingState(currentPrompt string, spinner string) string {
	if currentPrompt == "" {
		return ""
	}

	var s strings.Builder
	termWidth := GetTerminalWidth()
	promptWidth := termWidth - 3

	s.WriteString(promptSymbol)
	s.WriteString(" ")

	promptLineStyle := inputStyle.Width(promptWidth)
	s.WriteString(promptLineStyle.Render(currentPrompt))
	s.WriteString("\n\n")
	s.WriteString(promptStyle.Render(spinner))
	s.WriteString(thinkingStyle.Render(" Thinking..."))
	s.WriteString("\n")

	return s.String()
}

// ThinkingStateWithInput renders the thinking indicator with preserved input
func ThinkingStateWithInput(inputView string, spinner string) string {
	var s strings.Builder

	// Show input with prompt symbol
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

	// Add thinking indicator
	s.WriteString("\n\n")
	if spinner != "" {
		s.WriteString(promptStyle.Render(spinner))
		s.WriteString(" ")
	}
	s.WriteString(thinkingStyle.Render("Thinking..."))
	s.WriteString("\n")

	return s.String()
}

// Command represents a command with description
type Command struct {
	Command     string
	Description string
}

// CommandSuggestions renders command suggestions
func CommandSuggestions(suggestions []string, selectedIndex int, commands []Command) string {
	if len(suggestions) == 0 {
		return ""
	}

	var s strings.Builder
	s.WriteString("\n\n")
	s.WriteString(suggestionStyle.Render("Commands:"))
	s.WriteString("\n")

	for i, suggestion := range suggestions {
		if i == selectedIndex {
			s.WriteString(highlightStyle.Render(fmt.Sprintf("  → %s", suggestion)))
		} else {
			s.WriteString(suggestionStyle.Render(fmt.Sprintf("    %s", suggestion)))
		}

		// Add description
		for _, cmd := range commands {
			if cmd.Command == suggestion {
				s.WriteString(suggestionStyle.Render(fmt.Sprintf(" - %s", cmd.Description)))
				break
			}
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(suggestionStyle.Render("Press Tab or Enter to complete, ↑/↓ to navigate"))

	return s.String()
}

// ErrorMessage renders error messages
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error: %v", err))
}

// ThinkingText renders just the thinking text without prompt
func ThinkingText() string {
	return thinkingStyle.Render(" Thinking...")
}

// InfoMessage renders info messages
func InfoMessage(message string) string {
	if message == "" {
		return ""
	}
	return "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(message)
}
