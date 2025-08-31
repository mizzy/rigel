package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/command"
)

// convertCommandResultToCmd converts a Result to a tea.Cmd
func (m *Model) convertCommandResultToCmd(result command.Result) tea.Cmd {
	switch result.Type {
	case "response":
		return func() tea.Msg {
			return result
		}

	case "model_selector":
		if result.ModelSelector != nil {
			return func() tea.Msg {
				return *result.ModelSelector
			}
		}

	case "provider_selector":
		if result.ProviderSelector != nil {
			return func() tea.Msg {
				return *result.ProviderSelector
			}
		}

	case "status":
		if result.StatusInfo != nil {
			return func() tea.Msg {
				return *result.StatusInfo
			}
		}

	case "quit":
		m.quitting = true
		return tea.Quit

	case "clear":
		m.chatState.SetThinking(false)
		return nil

	case "request":
		return m.requestResponse(result.Prompt)
	}

	return nil
}
