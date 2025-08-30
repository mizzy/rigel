package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/llm"
)

// Minimalist style colors
var (
	promptSymbol   = lipgloss.NewStyle().Foreground(lipgloss.Color("35")).Render(">")
	inputStyle     = lipgloss.NewStyle()
	outputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	thinkingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	errorTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type SimpleModel struct {
	provider llm.Provider
	input    textinput.Model
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
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 5000
	ti.Width = 100
	ti.Prompt = ""

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("35"))

	return &SimpleModel{
		provider: provider,
		input:    ti,
		spinner:  s,
		history:  []Exchange{},
	}
}

func (m SimpleModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
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
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyCtrlD:
			if !m.thinking && m.input.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}

		case tea.KeyEnter:
			if !m.thinking && strings.TrimSpace(m.input.Value()) != "" {
				m.currentPrompt = m.input.Value()
				m.input.SetValue("")
				m.thinking = true
				m.err = nil

				return m, tea.Batch(
					m.requestResponse(m.currentPrompt),
					m.spinner.Tick,
				)
			}
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
		s.WriteString(m.input.View())
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
