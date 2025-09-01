package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/command"
	"github.com/mizzy/rigel/internal/ui/handlers"
)

// Update handles incoming messages and returns updated application state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle provider selection mode
		if m.llmState.IsProviderSelectionActive() {
			return m.handleProviderSelectionKey(msg)
		}

		// Handle model selection mode
		if m.llmState.IsModelSelectionActive() {
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
			if !m.chatState.IsThinking() && m.input.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
		default:
			// Reset Ctrl+C flag on any other key
			m.ctrlCPressed = false
			m.infoMessage = ""
		}

		// Handle Tab key for completion
		if msg.String() == "tab" && !m.chatState.IsThinking() && m.showCompletions {
			completionValue := m.completionHandler.GetCompletionValue(m.completions, m.selectedCompletion)
			if completionValue != "" {
				m.input.SetValue(completionValue)
				m.input.CursorEnd()
				m.showCompletions = false
				m.completions = []string{}
			}
			m.ctrlCPressed = false // Reset Ctrl+C flag
			m.infoMessage = ""
			return m, nil
		}

		// Handle arrow keys for suggestion navigation or history navigation
		if !m.chatState.IsThinking() {
			switch msg.String() {
			case "up":
				if m.showCompletions {
					if m.selectedCompletion > 0 {
						m.selectedCompletion--
					}
				} else {
					histState := m.getHistoryNavigationState()
					handlers.NavigateHistory(-1, &m.input, histState)
					m.updateFromHistoryNavigationState(histState)
				}
				m.ctrlCPressed = false // Reset Ctrl+C flag
				m.infoMessage = ""
				return m, nil
			case "down":
				if m.showCompletions {
					if m.selectedCompletion < len(m.completions)-1 {
						m.selectedCompletion++
					}
				} else {
					histState := m.getHistoryNavigationState()
					handlers.NavigateHistory(1, &m.input, histState)
					m.updateFromHistoryNavigationState(histState)
				}
				m.ctrlCPressed = false // Reset Ctrl+C flag
				m.infoMessage = ""
				return m, nil
			}
		}

		// Check for Enter key specifically (not Alt+Enter)
		if msg.String() == "enter" && !m.chatState.IsThinking() {
			m.ctrlCPressed = false // Reset Ctrl+C flag
			m.infoMessage = ""
			// If completions are shown and one is selected, complete and execute it
			if m.showCompletions {
				completionValue := m.completionHandler.GetCompletionValue(m.completions, m.selectedCompletion)
				if completionValue != "" {
					m.input.SetValue(completionValue)
					m.input.CursorEnd()
				}
				m.showCompletions = false
				m.completions = []string{}
				// After completing suggestion, check if it's a command and execute it
				if strings.HasPrefix(m.input.Value(), "/") {
					// Treat it as if user pressed Enter with the command
					prompt := m.input.Value()
					m.chatState.SetCurrentPrompt(prompt)

					// Save to input history
					m.inputHistory = append(m.inputHistory, prompt)
					m.historyIndex = -1
					m.currentInput = ""

					// Save to persistent history
					if m.historyManager != nil {
						_ = m.historyManager.Add(prompt)
					}

					m.input.SetValue("")
					m.chatState.SetThinking(true)
					m.chatState.ClearError()
					m.showCompletions = false

					// Handle the command
					trimmedPrompt := strings.TrimSpace(m.chatState.GetCurrentPrompt())
					result := command.HandleCommand(trimmedPrompt, m.llmState, m.chatState, m.config, m.historyManager, m.inputHistory)
					cmd := func() tea.Msg { return result }
					return m, tea.Batch(cmd, m.spinner.Tick)
				}
				return m, nil
			}

			if strings.TrimSpace(m.input.Value()) != "" {
				prompt := m.input.Value()
				m.chatState.SetCurrentPrompt(prompt)

				// Save to input history
				m.inputHistory = append(m.inputHistory, prompt)
				m.historyIndex = -1
				m.currentInput = ""

				// Save to persistent history
				if m.historyManager != nil {
					_ = m.historyManager.Add(prompt)
				}

				m.input.SetValue("")
				m.chatState.SetThinking(true)
				m.chatState.ClearError()
				m.showCompletions = false

				// Handle commands
				trimmedPrompt := strings.TrimSpace(m.chatState.GetCurrentPrompt())
				result := command.HandleCommand(trimmedPrompt, m.llmState, m.chatState, m.config, m.historyManager, m.inputHistory)
				cmd := func() tea.Msg { return result }
				return m, tea.Batch(cmd, m.spinner.Tick)
			}
			return m, nil
		}

		// Pass all other keys (including alt+enter and ctrl+j) to textarea
		if !m.chatState.IsThinking() && !m.llmState.IsModelSelectionActive() && !m.llmState.IsProviderSelectionActive() {
			oldValue := m.input.Value()
			m.input, cmd = m.input.Update(msg)

			// Update completions if input changed
			if oldValue != m.input.Value() {
				m.completions, m.showCompletions = m.completionHandler.UpdateCompletions(m.input.Value())
				m.selectedCompletion = 0
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
				return handlers.AIResponse{Error: msg.err}
			}
		}

		m.llmState.ActivateProviderSelection(msg.providers, msg.currentProvider)
		return m, nil

	case modelSelectorMsg:
		if msg.err != nil {
			return m, func() tea.Msg {
				return handlers.AIResponse{Error: msg.err}
			}
		}

		m.llmState.ActivateModelSelection(msg.models)
		m.input.SetValue("")
		m.input.Placeholder = "Type to filter models, Enter to select, Esc to cancel"

		return m, nil

	case command.Result:
		m.chatState.SetThinking(false)
		if msg.Error != nil {
			m.chatState.SetError(msg.Error)
		} else {
			switch msg.Type {
			case "clear_input_history":
				// Clear input history
				m.inputHistory = []string{}
				m.historyIndex = -1
				m.currentInput = ""
				m.chatState.AddExchange(m.chatState.GetCurrentPrompt(), "Command history cleared successfully.")
				m.chatState.ClearCurrentPrompt()
			case "request":
				// Handle normal prompts (non-commands)
				return m, handlers.RequestResponse(msg.Prompt, m.llmState, m.chatState)
			case "model_selector":
				if msg.ModelSelector != nil {
					return m, func() tea.Msg { return *msg.ModelSelector }
				}
			case "provider_selector":
				if msg.ProviderSelector != nil {
					return m, func() tea.Msg { return *msg.ProviderSelector }
				}
			case "status":
				if msg.StatusInfo != nil {
					return m, func() tea.Msg { return *msg.StatusInfo }
				}
			case "quit":
				m.quitting = true
				return m, tea.Quit
			case "clear":
				// Clear handled above by SetThinking(false)
				return m, nil
			default:
				if msg.Content != "" {
					m.chatState.AddExchange(m.chatState.GetCurrentPrompt(), msg.Content)
					m.chatState.ClearCurrentPrompt()
				}
			}
		}
		return m, nil

	case command.ModelSelectorMsg:
		if msg.Error != nil {
			return m, func() tea.Msg {
				return handlers.AIResponse{Error: msg.Error}
			}
		}

		m.llmState.ActivateModelSelection(msg.Models)
		m.input.SetValue("")
		m.input.Placeholder = "Type to filter models, Enter to select, Esc to cancel"

		return m, nil

	case command.ProviderSelectorMsg:
		m.llmState.ActivateProviderSelection(msg.Providers, msg.CurrentProvider)
		return m, nil

	case command.StatusInfo:
		m.chatState.SetThinking(false)
		// Convert StatusInfo to formatted string and display it
		statusContent := fmt.Sprintf("✦ Rigel Session Status\n\n"+
			"🤖 LLM Configuration\n"+
			"  Provider: %s\n"+
			"  Model: %s\n\n"+
			"💬 Chat History\n"+
			"  Messages: %d\n"+
			"  User tokens: ~%d\n"+
			"  Assistant tokens: ~%d\n"+
			"  Total tokens: ~%d\n\n"+
			"📝 Command History\n"+
			"  Commands saved: %d\n"+
			"  Persistence: %s\n\n"+
			"🔧 Environment\n"+
			"  Log level: %s\n"+
			"  Repository context: %s\n",
			msg.Provider, msg.Model,
			msg.MessageCount,
			msg.UserTokens, msg.AssistantTokens, msg.TotalTokens,
			msg.CommandsCount,
			map[bool]string{true: "✓ Enabled", false: "✗ Disabled"}[msg.PersistenceEnabled],
			msg.LogLevel,
			map[bool]string{true: "✓ AGENTS.md loaded", false: "✗ Not initialized (run /init)"}[msg.RepositoryInitialized])

		m.chatState.AddExchange(m.chatState.GetCurrentPrompt(), statusContent)
		m.chatState.ClearCurrentPrompt()
		return m, nil

	case providerSwitchResponse:
		m.llmState.SetCurrentProvider(msg.provider)
		m.chatState.SetThinking(false)
		response := fmt.Sprintf("Switched to provider: %s\nCurrent model: %s", msg.providerName, m.llmState.GetCurrentModel().Name)
		m.chatState.AddExchange(m.chatState.GetCurrentPrompt(), response)
		m.chatState.ClearCurrentPrompt()
		return m, nil

	case handlers.AIResponse:
		m.chatState.SetThinking(false)

		if msg.Error != nil {
			m.chatState.SetError(msg.Error)
		} else {
			m.chatState.AddExchange(m.chatState.GetCurrentPrompt(), msg.Content)
			m.chatState.ClearCurrentPrompt()
		}
		return m, nil

	case spinner.TickMsg:
		if m.chatState.IsThinking() {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.chatState.IsThinking() && !m.llmState.IsModelSelectionActive() && !m.llmState.IsProviderSelectionActive() {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
