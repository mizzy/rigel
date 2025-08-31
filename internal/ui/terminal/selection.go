package terminal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/llm"
)

func (m *Model) handleProviderSelectionKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.exitProviderSelection()
		return m, nil

	case tea.KeyEnter:
		if m.selectedProviderIndex < len(m.availableProviders) {
			selectedProvider := m.availableProviders[m.selectedProviderIndex]
			m.exitProviderSelection()
			return m, m.switchProvider(selectedProvider)
		}
		return m, nil

	case tea.KeyUp:
		if m.selectedProviderIndex > 0 {
			m.selectedProviderIndex--
		}
		return m, nil

	case tea.KeyDown:
		if m.selectedProviderIndex < len(m.availableProviders)-1 {
			m.selectedProviderIndex++
		}
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
		if len(m.filteredModels) > 0 && m.selectedModelIndex < len(m.filteredModels) {
			selectedModel := m.filteredModels[m.selectedModelIndex]
			m.exitModelSelection()
			return m, m.switchModel(selectedModel.Name)
		}
		return m, nil

	case tea.KeyUp:
		if m.selectedModelIndex > 0 {
			m.selectedModelIndex--
		}
		return m, nil

	case tea.KeyDown:
		if m.selectedModelIndex < len(m.filteredModels)-1 {
			m.selectedModelIndex++
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		newFilter := m.input.Value()
		if newFilter != m.modelFilter {
			m.modelFilter = newFilter
			m.filterModels()
			m.selectedModelIndex = 0
		}

		return m, cmd
	}
}

func (m *Model) exitProviderSelection() {
	m.providerSelectionMode = false
	m.availableProviders = nil
	m.selectedProviderIndex = 0
	m.chatState.SetThinking(false)
}

func (m *Model) exitModelSelection() {
	m.modelSelectionMode = false
	m.input.SetValue("")
	m.input.Placeholder = "Type a message or / for commands (Alt+Enter for new line)"
	m.modelFilter = ""
	m.filteredModels = nil
	m.availableModels = nil
	m.selectedModelIndex = 0
	m.chatState.SetThinking(false)
}

func (m *Model) filterModels() {
	if m.modelFilter == "" {
		m.filteredModels = m.availableModels
		return
	}

	filter := strings.ToLower(m.modelFilter)
	m.filteredModels = nil

	for _, model := range m.availableModels {
		if strings.Contains(strings.ToLower(model.Name), filter) {
			m.filteredModels = append(m.filteredModels, model)
		}
	}
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
	m.provider.SetModel(modelName)

	return func() tea.Msg {
		return aiResponse{
			content: fmt.Sprintf("Switched to model: %s", modelName),
		}
	}
}
