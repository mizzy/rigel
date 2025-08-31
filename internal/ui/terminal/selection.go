package terminal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/ui/handlers"
)

func (m *Model) handleProviderSelectionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.exitProviderSelection()
		return m, nil

	case tea.KeyEnter:
		if provider, ok := m.llmState.GetSelectedProvider(); ok {
			m.exitProviderSelection()
			return m, m.switchProvider(provider)
		}
		return m, nil

	case tea.KeyUp:
		m.llmState.MoveProviderSelectionUp()
		return m, nil

	case tea.KeyDown:
		m.llmState.MoveProviderSelectionDown()
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleModelSelectionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.exitModelSelection()
		return m, nil

	case tea.KeyEnter:
		if model, ok := m.llmState.GetSelectedModel(); ok {
			m.exitModelSelection()
			return m, m.switchModel(model)
		}
		return m, nil

	case tea.KeyUp:
		m.llmState.MoveModelSelectionUp()
		return m, nil

	case tea.KeyDown:
		m.llmState.MoveModelSelectionDown()
		return m, nil

	case tea.KeyBackspace:
		// Handle backspace for filtering - manually update filter
		currentFilter := m.llmState.GetModelFilter()
		if len(currentFilter) > 0 {
			newFilter := currentFilter[:len(currentFilter)-1]
			m.llmState.SetModelFilter(newFilter)
			m.input.SetValue(newFilter)
		}
		return m, nil

	case tea.KeyRunes:
		// Handle printable characters for filtering
		runes := msg.Runes
		if len(runes) > 0 {
			// Ignore non-printable characters
			if runes[0] < 32 || runes[0] > 126 {
				return m, nil
			}

			// Add to filter
			currentFilter := m.llmState.GetModelFilter()
			newFilter := currentFilter + string(runes)
			m.llmState.SetModelFilter(newFilter)
			m.input.SetValue(newFilter)
		}
		return m, nil

	default:
		// Explicitly ignore all other keys to prevent unintended behavior
		return m, nil
	}
}

func (m *Model) exitProviderSelection() {
	m.llmState.DeactivateProviderSelection()
	m.chatState.SetThinking(false)
}

func (m *Model) exitModelSelection() {
	m.llmState.DeactivateModelSelection()
	m.input.SetValue("")
	m.input.Placeholder = "Type a message or / for commands (Alt+Enter for new line)"
	m.chatState.SetThinking(false)
}

func (m *Model) switchProvider(provider llm.Provider) tea.Cmd {
	return func() tea.Msg {
		// Update config
		if m.config != nil {
			m.config.Provider = provider.GetName()
		}

		return providerSwitchResponse{
			provider:     provider,
			providerName: provider.GetName(),
		}
	}
}

func (m *Model) switchModel(model llm.Model) tea.Cmd {
	// Actually switch the model
	provider := m.llmState.GetCurrentProvider()
	if provider != nil {
		provider.SetModel(model)
		m.llmState.SetCurrentModel(model)
	}

	return func() tea.Msg {
		return handlers.AIResponse{
			Content: fmt.Sprintf("Switched to model: %s", model.Name),
		}
	}
}
