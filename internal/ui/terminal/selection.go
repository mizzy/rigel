package terminal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/llm"
)

func (m *Model) handleProviderSelectionKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
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

func (m *Model) handleModelSelectionKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.exitModelSelection()
		return m, nil

	case tea.KeyEnter:
		if model, ok := m.llmState.GetSelectedModel(); ok {
			m.exitModelSelection()
			return m, m.switchModel(model.Name)
		}
		return m, nil

	case tea.KeyUp:
		m.llmState.MoveModelSelectionUp()
		return m, nil

	case tea.KeyDown:
		m.llmState.MoveModelSelectionDown()
		return m, nil

	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		newFilter := m.input.Value()
		if newFilter != m.llmState.GetModelFilter() {
			m.llmState.SetModelFilter(newFilter)
		}

		return m, cmd
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

func (m *Model) switchProvider(providerName string) tea.Cmd {
	return func() tea.Msg {
		// Update config
		if m.config != nil {
			m.config.Provider = providerName
		}

		// Create new provider
		newProvider, err := llm.NewProvider(m.config)
		if err != nil {
			return aiResponse{err: fmt.Errorf("failed to switch provider: %w", err)}
		}

		return providerSwitchResponse{
			provider:     newProvider,
			providerName: providerName,
		}
	}
}

func (m *Model) switchModel(modelName string) tea.Cmd {
	// Actually switch the model
	provider := m.llmState.GetProvider()
	if provider != nil {
		provider.SetModel(modelName)
		m.llmState.SetCurrentModel(modelName)
	}

	return func() tea.Msg {
		return aiResponse{
			content: fmt.Sprintf("Switched to model: %s", modelName),
		}
	}
}
