package completion

import "strings"

// Command represents a command with its description
type Command struct {
	Command     string
	Description string
}

// AvailableCommands contains all available commands for completion
var AvailableCommands = []Command{
	{"/init", "Analyze repository and generate AGENTS.md"},
	{"/model", "Show current model and select from available models"},
	{"/provider", "Switch between LLM providers (Anthropic, Ollama, etc.)"},
	{"/status", "Show current session status and configuration"},
	{"/help", "Show available commands"},
	{"/clear", "Clear chat history"},
	{"/clearhistory", "Clear command history"},
	{"/exit", "Exit the application"},
	{"/quit", "Exit the application"},
}

// Handler handles command completion functionality
type Handler struct{}

// NewHandler creates a new completion handler
func NewHandler() *Handler {
	return &Handler{}
}

// UpdateCompletions updates the command completions based on user input
func (h *Handler) UpdateCompletions(inputValue string) ([]string, bool) {
	completions := []string{}
	showCompletions := false

	// Check if the input starts with /
	if strings.HasPrefix(inputValue, "/") {
		prefix := strings.ToLower(inputValue)
		for _, cmd := range AvailableCommands {
			if strings.HasPrefix(strings.ToLower(cmd.Command), prefix) {
				completions = append(completions, cmd.Command)
			}
		}
		if len(completions) > 0 && inputValue != completions[0] {
			showCompletions = true
		}
	}

	return completions, showCompletions
}

// GetCompletionValue returns the value to complete with based on selected completion
func (h *Handler) GetCompletionValue(completions []string, selectedCompletion int) string {
	if selectedCompletion < len(completions) {
		return completions[selectedCompletion]
	}
	return ""
}
