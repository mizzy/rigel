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
	thinking           bool
	currentPrompt      string
	err                error
	quitting           bool
	suggestions        []string
	selectedSuggestion int
	showSuggestions    bool
	ctrlCPressed       bool
	infoMessage        string
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
		provider: provider,
		input:    ta,
		spinner:  s,
		history:  []Exchange{},
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

		// Handle arrow keys for suggestion navigation
		if m.showSuggestions && !m.thinking {
			switch msg.String() {
			case "up":
				if m.selectedSuggestion > 0 {
					m.selectedSuggestion--
				}
				m.ctrlCPressed = false // Reset Ctrl+C flag
				m.infoMessage = ""
				return m, nil
			case "down":
				if m.selectedSuggestion < len(m.suggestions)-1 {
					m.selectedSuggestion++
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
			}

			return m, cmd
		}

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
		// Get the value and cursor position from textarea without its styling
		value := m.input.Value()
		if value == "" {
			// Add cursor first when empty
			if m.input.Focused() {
				s.WriteString("█")
			}
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("60")).Render(" " + m.input.Placeholder))
		} else {
			// Add indentation for multi-line input
			lines := strings.Split(value, "\n")
			for i, line := range lines {
				if i > 0 {
					s.WriteString("\n  ") // 2 spaces to align with prompt symbol + space
				}
				s.WriteString(line)
			}
			// Add cursor after text
			if m.input.Focused() {
				s.WriteString("█")
			}
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

// NewSimpleModel is an alias for NewChatModel for backward compatibility
func NewSimpleModel(provider llm.Provider) *ChatModel {
	return NewChatModel(provider)
}

// SimpleModel is an alias for ChatModel for backward compatibility
type SimpleModel = ChatModel
