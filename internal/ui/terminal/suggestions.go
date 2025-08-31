package terminal

import "strings"

// updateSuggestions updates the command suggestions based on user input
func (m *Model) updateSuggestions() {
	value := m.input.Value()
	m.suggestions = []string{}
	m.showSuggestions = false
	m.selectedSuggestion = 0

	// Check if the input starts with /
	if strings.HasPrefix(value, "/") {
		prefix := strings.ToLower(value)
		for _, cmd := range availableCommands {
			if strings.HasPrefix(strings.ToLower(cmd.command), prefix) {
				m.suggestions = append(m.suggestions, cmd.command)
			}
		}
		if len(m.suggestions) > 0 && value != m.suggestions[0] {
			m.showSuggestions = true
		}
	}
}

// completeSuggestion completes the selected suggestion
func (m *Model) completeSuggestion() {
	if m.showSuggestions && m.selectedSuggestion < len(m.suggestions) {
		m.input.SetValue(m.suggestions[m.selectedSuggestion])
		m.input.CursorEnd()
		m.showSuggestions = false
		m.suggestions = []string{}
	}
}
