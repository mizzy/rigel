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

// Rigel-inspired color scheme (blue-white star)
var (
	promptSymbol   = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true).Render("✦") // Light blue star symbol
	inputStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))                       // Very light blue-white
	outputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	thinkingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Italic(true) // Soft blue
	errorTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type SimpleModel struct {
	provider llm.Provider
	input    textarea.Model
	spinner  spinner.Model

	history       []Exchange
	thinking      bool
	currentPrompt string
	err           error
	quitting      bool
}

type Exchange struct {
	Prompt   string
	Response string
}

func NewSimpleModel(provider llm.Provider) *SimpleModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message (Alt+Enter or Ctrl+J for new line, Enter to send)"
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
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("68")) // Dim blue
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()

	ta.BlurredStyle.Base = noBorder
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("68")) // Dim blue
	ta.BlurredStyle.Prompt = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.EndOfBuffer = lipgloss.NewStyle()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("87")) // Match Rigel blue

	return &SimpleModel{
		provider: provider,
		input:    ta,
		spinner:  s,
		history:  []Exchange{},
	}
}

func (m SimpleModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}

type aiResponse struct {
	content string
	err     error
}

func (m SimpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle special keys first
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyCtrlD:
			if !m.thinking && m.input.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
		}

		// Check for Enter key specifically (not Alt+Enter)
		if msg.String() == "enter" && !m.thinking {
			if strings.TrimSpace(m.input.Value()) != "" {
				m.currentPrompt = m.input.Value()
				m.input.SetValue("")
				m.thinking = true
				m.err = nil

				return m, tea.Batch(
					m.requestResponse(m.currentPrompt),
					m.spinner.Tick,
				)
			}
			return m, nil
		}

		// Pass all other keys (including alt+enter and ctrl+j) to textarea
		if !m.thinking {
			m.input, cmd = m.input.Update(msg)
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

func (m SimpleModel) View() string {
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

	// Display error if any
	if m.err != nil {
		s.WriteString(errorTextStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\n")
	}

	// Display input prompt
	if !m.thinking {
		s.WriteString(promptSymbol)
		s.WriteString(" ")
		// Get the value and cursor position from textarea without its styling
		value := m.input.Value()
		if value == "" {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("68")).Render(m.input.Placeholder))
		} else {
			// Add indentation for multi-line input
			lines := strings.Split(value, "\n")
			for i, line := range lines {
				if i > 0 {
					s.WriteString("\n  ") // 2 spaces to align with prompt symbol + space
				}
				s.WriteString(line)
			}
		}
		// Add cursor
		if m.input.Focused() {
			s.WriteString("█")
		}
	}

	return s.String()
}

func (m *SimpleModel) requestResponse(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		response, err := m.provider.Generate(ctx, prompt)
		if err != nil {
			return aiResponse{err: err}
		}
		return aiResponse{content: strings.TrimSpace(response)}
	}
}
