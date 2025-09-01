package command

import "strings"

// CompletionHandler handles command completion functionality
type CompletionHandler struct{}

// NewCompletionHandler creates a new completion handler
func NewCompletionHandler() *CompletionHandler {
	return &CompletionHandler{}
}

// UpdateCompletions updates the command completions based on user input
func (h *CompletionHandler) UpdateCompletions(inputValue string) ([]string, bool) {
	completions := []string{}
	showCompletions := false

	// Check if the input starts with / (without leading spaces)
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
func (h *CompletionHandler) GetCompletionValue(completions []string, selectedCompletion int) string {
	if selectedCompletion < len(completions) {
		return completions[selectedCompletion]
	}
	return ""
}
