package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/llm"
)

// ChatModel represents the main chat interface
type ChatModel struct {
	provider llm.Provider
	input    textarea.Model
	spinner  spinner.Model

	history            []Exchange
	inputHistory       []string
	historyIndex       int
	currentInput       string
	thinking           bool
	currentPrompt      string
	err                error
	quitting           bool
	suggestions        []string
	selectedSuggestion int
	showSuggestions    bool
	ctrlCPressed       bool
	infoMessage        string

	// Model selection mode
	modelSelectionMode bool
	availableModels    []llm.Model
	filteredModels     []llm.Model
	selectedModelIndex int
	modelFilter        string
}

// Exchange represents a single chat exchange
type Exchange struct {
	Prompt   string
	Response string
}

// NewChatModel creates a new chat model instance
func NewChatModel(provider llm.Provider) *ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message or / for commands (Alt+Enter for new line)"
	ta.Focus()
	ta.CharLimit = 5000
	ta.SetWidth(100)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.Prompt = ""                // Remove the default prompt
	ta.EndOfBufferCharacter = ' ' // Use space instead of default character
	ta.KeyMap.InsertNewline.SetKeys("alt+enter", "ctrl+j")

	// Remove borders and styling
	noBorder := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false).
		BorderForeground(lipgloss.NoColor{}).
		Padding(0).
		Margin(0)

	ta.FocusedStyle.Base = noBorder
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("60")) // Dim blue
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()

	ta.BlurredStyle.Base = noBorder
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("60")) // Dim blue
	ta.BlurredStyle.Prompt = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.EndOfBuffer = lipgloss.NewStyle()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("87")) // Match Rigel blue

	return &ChatModel{
		provider:     provider,
		input:        ta,
		spinner:      s,
		history:      []Exchange{},
		inputHistory: []string{},
		historyIndex: -1,
	}
}

// Init initializes the chat model
func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}

// Update handles messages and updates the model
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			// If suggestions are shown and one is selected, complete it
			if m.showSuggestions {
				m.completeSuggestion()
				return m, nil
			}

			if strings.TrimSpace(m.input.Value()) != "" {
				m.currentPrompt = m.input.Value()

				// Save to input history
				m.inputHistory = append(m.inputHistory, m.currentPrompt)
				m.historyIndex = -1
				m.currentInput = ""

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

func (m *ChatModel) handleModelSelectionKey(msg tea.KeyMsg) (*ChatModel, tea.Cmd) {
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

func (m *ChatModel) exitModelSelection() {
	m.modelSelectionMode = false
	m.input.SetValue("")
	m.input.Placeholder = "Type a message or / for commands (Alt+Enter for new line)"
	m.modelFilter = ""
	m.filteredModels = nil
	m.availableModels = nil
	m.selectedModelIndex = 0
}

func (m *ChatModel) filterModels() {
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

func (m *ChatModel) switchModel(modelName string) tea.Cmd {
	// Actually switch the model
	m.provider.SetModel(modelName)

	return func() tea.Msg {
		return aiResponse{
			content: fmt.Sprintf("Switched to model: %s", modelName),
		}
	}
}

// View renders the chat interface
func (m ChatModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Display history
	for _, ex := range m.history {
		// User prompt with > symbol
		s.WriteString(promptSymbol)
		s.WriteString(" ")
		s.WriteString(inputStyle.Render(ex.Prompt))
		s.WriteString("\n\n")

		// Assistant response
		s.WriteString(outputStyle.Render(ex.Response))
		s.WriteString("\n\n")
	}

	// Display model selection interface if in model selection mode
	if m.modelSelectionMode {
		s.WriteString("\nModel Selection\n\n")
		s.WriteString(fmt.Sprintf("Current: %s\n\n", currentModelStyle.Render(m.provider.GetCurrentModel())))

		if len(m.filteredModels) == 0 {
			s.WriteString("No models match your filter.\n")
		} else {
			s.WriteString("Available models:\n")
			for i, model := range m.filteredModels {
				prefix := "  "
				isSelected := i == m.selectedModelIndex
				isCurrent := model.Name == m.provider.GetCurrentModel()

				modelDisplay := model.Name
				if model.Details.ParameterSize != "" {
					modelDisplay += fmt.Sprintf(" (%s)", model.Details.ParameterSize)
				}

				// Apply styles based on state
				if isCurrent {
					// Current model gets blue bold style
					modelDisplay = currentModelStyle.Render(modelDisplay)
					if isSelected {
						prefix = currentModelStyle.Render("❯ ")
					}
				} else if isSelected {
					// Selected (cursor) model gets yellow style
					modelDisplay = selectedModelStyle.Render(modelDisplay)
					prefix = selectedModelStyle.Render("❯ ")
				}

				s.WriteString(fmt.Sprintf("%s%s\n", prefix, modelDisplay))
			}
		}

		s.WriteString("\nFilter: ")
		s.WriteString(m.input.View())
		s.WriteString("\n\nUse up/down arrows to navigate, Enter to select, Esc to cancel")

		return s.String()
	}

	// Display current prompt if thinking
	if m.thinking && m.currentPrompt != "" {
		s.WriteString(promptSymbol)
		s.WriteString(" ")
		s.WriteString(inputStyle.Render(m.currentPrompt))
		s.WriteString("\n\n")
		s.WriteString(m.spinner.View())
		s.WriteString(thinkingStyle.Render(" Thinking..."))
		s.WriteString("\n")
	}

	// Display input prompt
	if !m.thinking {
		s.WriteString(promptSymbol)
		s.WriteString(" ")
		// Use textarea's native rendering to handle IME and cursor properly
		textareaView := m.input.View()
		// Handle multi-line alignment by replacing newlines with proper indentation
		lines := strings.Split(textareaView, "\n")
		for i, line := range lines {
			if i > 0 {
				s.WriteString("\n  ") // 2 spaces to align with prompt symbol + space
			}
			s.WriteString(line)
		}

		// Display command suggestions
		if m.showSuggestions && len(m.suggestions) > 0 {
			s.WriteString("\n\n")
			s.WriteString(suggestionStyle.Render("Commands:"))
			s.WriteString("\n")
			for i, suggestion := range m.suggestions {
				if i == m.selectedSuggestion {
					s.WriteString(highlightStyle.Render(fmt.Sprintf("  → %s", suggestion)))
				} else {
					s.WriteString(suggestionStyle.Render(fmt.Sprintf("    %s", suggestion)))
				}
				// Add description
				for _, cmd := range availableCommands {
					if cmd.command == suggestion {
						s.WriteString(suggestionStyle.Render(fmt.Sprintf(" - %s", cmd.description)))
						break
					}
				}
				s.WriteString("\n")
			}
			s.WriteString("\n")
			s.WriteString(suggestionStyle.Render("Press Tab or Enter to complete, ↑/↓ to navigate"))
		}

		// Display info message or error at the bottom
		if m.infoMessage != "" {
			s.WriteString("\n\n")
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(m.infoMessage))
		} else if m.err != nil {
			s.WriteString("\n\n")
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err.Error()))
		}
	}

	return s.String()
}

// requestResponse sends a request to the LLM provider
func (m *ChatModel) requestResponse(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		response, err := m.provider.Generate(ctx, prompt)
		if err != nil {
			return aiResponse{err: err}
		}
		return aiResponse{content: strings.TrimSpace(response)}
	}
}

// aiResponse represents a response from the AI
type aiResponse struct {
	content string
	err     error
}

// navigateHistory navigates through input history
func (m *ChatModel) navigateHistory(direction int) {
	if len(m.inputHistory) == 0 {
		return
	}

	// Save current input if we're starting to navigate history
	if m.historyIndex == -1 {
		m.currentInput = m.input.Value()
	}

	if direction < 0 {
		// Going up (backward) in history
		if m.historyIndex == -1 {
			// Start from the most recent item
			m.historyIndex = 0
		} else if m.historyIndex < len(m.inputHistory)-1 {
			m.historyIndex++
		}

		if m.historyIndex < len(m.inputHistory) {
			historyPos := len(m.inputHistory) - 1 - m.historyIndex
			m.input.SetValue(m.inputHistory[historyPos])
		}
	} else {
		// Going down (forward) in history
		if m.historyIndex > 0 {
			m.historyIndex--
			historyPos := len(m.inputHistory) - 1 - m.historyIndex
			m.input.SetValue(m.inputHistory[historyPos])
		} else if m.historyIndex == 0 {
			// Return to current input
			m.historyIndex = -1
			m.input.SetValue(m.currentInput)
		}
	}
}
