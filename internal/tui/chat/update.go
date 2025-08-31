package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle provider selection mode
		if m.providerSelectionMode {
			return m.handleProviderSelectionKey(msg)
		}

		// Handle model selection mode
		if m.modelSelectionMode {
			return m.handleModelSelectionKey(msg)
		}

		// Handle special keys first
		switch msg.Type {
		case tea.KeyCtrlC:
			if m.ctrlCPressed {
				m.quitting = true
				return m, tea.Quit
			}
			m.ctrlCPressed = true
			m.infoMessage = "Press Ctrl+C again to exit"
			// Reset Ctrl+C flag after any other key
			return m, nil

		case tea.KeyCtrlD:
			// Reset Ctrl+C flag on any other key
			m.ctrlCPressed = false
			m.infoMessage = ""
			if !m.thinking && m.input.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
		default:
			// Reset Ctrl+C flag on any other key
			m.ctrlCPressed = false
			m.infoMessage = ""
		}

		// Handle Tab key for completion
		if msg.String() == "tab" && !m.thinking && m.showSuggestions {
			m.completeSuggestion()
			m.ctrlCPressed = false // Reset Ctrl+C flag
			m.infoMessage = ""
			return m, nil
		}

		// Handle arrow keys for suggestion navigation or history navigation
		if !m.thinking {
			switch msg.String() {
			case "up":
				if m.showSuggestions {
					if m.selectedSuggestion > 0 {
						m.selectedSuggestion--
					}
				} else {
					m.navigateHistory(-1)
				}
				m.ctrlCPressed = false // Reset Ctrl+C flag
				m.infoMessage = ""
				return m, nil
			case "down":
				if m.showSuggestions {
					if m.selectedSuggestion < len(m.suggestions)-1 {
						m.selectedSuggestion++
					}
				} else {
					m.navigateHistory(1)
				}
				m.ctrlCPressed = false // Reset Ctrl+C flag
				m.infoMessage = ""
				return m, nil
			}
		}

		// Check for Enter key specifically (not Alt+Enter)
		if msg.String() == "enter" && !m.thinking {
			m.ctrlCPressed = false // Reset Ctrl+C flag
			m.infoMessage = ""
			// If suggestions are shown and one is selected, complete and execute it
			if m.showSuggestions {
				m.completeSuggestion()
				// After completing suggestion, check if it's a command and execute it
				if strings.HasPrefix(m.input.Value(), "/") {
					// Treat it as if user pressed Enter with the command
					m.currentPrompt = m.input.Value()

					// Save to input history
					m.inputHistory = append(m.inputHistory, m.currentPrompt)
					m.historyIndex = -1
					m.currentInput = ""

					// Save to persistent history
					if m.historyManager != nil {
						_ = m.historyManager.Add(m.currentPrompt)
					}

					m.input.SetValue("")
					m.thinking = true
					m.err = nil
					m.showSuggestions = false

					// Handle the command
					trimmedPrompt := strings.TrimSpace(m.currentPrompt)
					cmd := m.handleCommand(trimmedPrompt)
					if cmd != nil {
						return m, tea.Batch(cmd, m.spinner.Tick)
					}
				}
				return m, nil
			}

			if strings.TrimSpace(m.input.Value()) != "" {
				m.currentPrompt = m.input.Value()

				// Save to input history
				m.inputHistory = append(m.inputHistory, m.currentPrompt)
				m.historyIndex = -1
				m.currentInput = ""

				// Save to persistent history
				if m.historyManager != nil {
					_ = m.historyManager.Add(m.currentPrompt)
				}

				m.input.SetValue("")
				m.thinking = true
				m.err = nil
				m.showSuggestions = false

				// Handle commands
				trimmedPrompt := strings.TrimSpace(m.currentPrompt)
				cmd := m.handleCommand(trimmedPrompt)
				if cmd != nil {
					return m, tea.Batch(cmd, m.spinner.Tick)
				}
			}
			return m, nil
		}

		// Pass all other keys (including alt+enter and ctrl+j) to textarea
		if !m.thinking {
			oldValue := m.input.Value()
			m.input, cmd = m.input.Update(msg)

			// Update suggestions if input changed
			if oldValue != m.input.Value() {
				m.updateSuggestions()
				m.ctrlCPressed = false // Reset Ctrl+C flag when typing
				m.infoMessage = ""

				// Reset history navigation if user types
				if m.historyIndex != -1 {
					m.historyIndex = -1
					m.currentInput = m.input.Value()
				}
			}

			return m, cmd
		}

	case providerSelectorMsg:
		if msg.err != nil {
			return m, func() tea.Msg {
				return aiResponse{err: msg.err}
			}
		}

		m.providerSelectionMode = true
		m.availableProviders = msg.providers
		m.selectedProviderIndex = 0

		// Find current provider index
		for i, p := range msg.providers {
			if p == msg.currentProvider {
				m.selectedProviderIndex = i
				break
			}
		}

		return m, nil

	case modelSelectorMsg:
		if msg.err != nil {
			return m, func() tea.Msg {
				return aiResponse{err: msg.err}
			}
		}

		m.modelSelectionMode = true
		m.availableModels = msg.models
		m.filteredModels = msg.models
		m.selectedModelIndex = 0
		m.modelFilter = ""
		m.input.SetValue("")
		m.input.Placeholder = "Type to filter models, Enter to select, Esc to cancel"

		return m, nil

	case providerSwitchResponse:
		m.provider = msg.provider
		m.thinking = false
		m.history = append(m.history, Exchange{
			Prompt:   m.currentPrompt,
			Response: fmt.Sprintf("Switched to provider: %s\nCurrent model: %s", msg.providerName, m.provider.GetCurrentModel()),
		})
		m.currentPrompt = ""
		return m, nil

	case aiResponse:
		m.thinking = false

		if msg.err != nil {
			m.err = msg.err
		} else {
			m.history = append(m.history, Exchange{
				Prompt:   m.currentPrompt,
				Response: msg.content,
			})
			m.currentPrompt = ""
		}
		return m, nil

	case spinner.TickMsg:
		if m.thinking {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.thinking {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
