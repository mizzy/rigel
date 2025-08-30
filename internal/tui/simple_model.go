package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/agent"
	"github.com/mizzy/rigel/internal/llm"
)

// SimpleModel is a simplified TUI model for chat-only interface
type SimpleModel struct {
	agent    *agent.Agent
	provider llm.Provider
	width    int
	height   int
	ready    bool

	// Chat components
	textarea textarea.Model
	viewport viewport.Model
	messages []ChatMessage

	// UI components
	help help.Model
	keys simpleKeyMap

	// Theme
	theme *Theme

	// Error message
	err error

	// Analyzer for /init command
	analyzer *RepoAnalyzer

	// Command suggestion
	suggestion string
}

type ChatMessage struct {
	Role    string
	Content string
	Time    time.Time
}

type simpleKeyMap struct {
	Enter key.Binding
	Init  key.Binding
	Clear key.Binding
	Help  key.Binding
	Quit  key.Binding
}

func NewSimpleModel(provider llm.Provider) *SimpleModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message or /init... (Enter to send, Alt+Enter or Ctrl+J for newline)"
	ta.CharLimit = 10000
	ta.SetHeight(1) // Start with 1 line
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	// Configure newline key bindings - use Alt+Enter or Ctrl+J
	// Note: Shift+Enter doesn't work in most terminals
	ta.KeyMap.InsertNewline = key.NewBinding(
		key.WithKeys("alt+enter", "ctrl+j"),
		key.WithHelp("alt+enter/ctrl+j", "new line"),
	)
	ta.KeyMap.InsertNewline.SetEnabled(true)

	ta.Focus() // Ensure textarea has focus

	// Use a custom prompt function to show prompt only on first line
	ta.SetPromptFunc(2, func(lineIdx int) string {
		if lineIdx == 0 {
			return "> "
		}
		return "  " // Two spaces for alignment on subsequent lines
	})

	// Style the prompt
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89"))

	// Ensure cursor line style doesn't hide text
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

	vp := viewport.New(80, 20)

	a := agent.New(provider)

	m := &SimpleModel{
		agent:    a,
		provider: provider,
		textarea: ta,
		viewport: vp,
		messages: []ChatMessage{},
		help:     help.New(),
		theme:    NewDefaultTheme(),
		analyzer: NewRepoAnalyzer(provider, ""),
	}

	m.keys = simpleKeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter"), // Changed from ctrl+enter to just enter for single line
			key.WithHelp("enter", "send message"),
		),
		Init: key.NewBinding(
			key.WithKeys("ctrl+i"),
			key.WithHelp("ctrl+i", "analyze repository (/init)"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear chat"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "ctrl+h"),
			key.WithHelp("?/ctrl+h", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}

	// Add welcome message
	m.addMessage("assistant", "Welcome to Rigel! Type your message or use /init to analyze the repository. Use /quit to exit.")

	return m
}

func (m *SimpleModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		tea.EnterAltScreen,
		m.textarea.Focus(), // Ensure focus on init
	)
}

func (m *SimpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 1
		// Fixed input area height (minimum space to prevent layout shift)
		inputHeight := 2 // Minimal height (1 line + 1 for border)

		if !m.ready {
			m.viewport = viewport.New(m.width, m.height-headerHeight-footerHeight-inputHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.renderMessages())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = m.height - headerHeight - footerHeight - inputHeight
		}

		m.textarea.SetWidth(m.width - 4)

	case tea.KeyMsg:
		// Check for special control keys first
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Init):
			m.handleInput("/init")
			// Show immediate feedback
			m.addMessage("assistant", "🔍 Analyzing repository structure...")
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			cmds = append(cmds, m.runInit())
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.Clear):
			m.messages = []ChatMessage{
				{Role: "assistant", Content: "Chat cleared. How can I help you?", Time: time.Now()},
			}
			m.agent.ClearMemory()
			m.viewport.SetContent(m.renderMessages())
			return m, nil

		case key.Matches(msg, m.keys.Help):
			// Toggle help in messages
			helpText := `Available Commands:
• /init - Analyze repository and generate AGENTS.md
• /quit - Exit the application
• Enter - Send message
• Alt+Enter or Ctrl+J - New line in message
• Ctrl+L - Clear chat
• Ctrl+C - Force quit
• ? - Show this help`
			m.addMessage("system", helpText)
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			return m, nil
		}

		// Handle Tab for command completion
		if msg.String() == "tab" {
			currentValue := m.textarea.Value()
			if strings.HasPrefix(currentValue, "/") && !strings.Contains(currentValue, " ") {
				if currentValue == "/" {
					// Default to /init when just "/"
					m.textarea.SetValue("/init")
					return m, nil
				} else if strings.HasPrefix("/init", currentValue) {
					m.textarea.SetValue("/init")
					return m, nil
				} else if strings.HasPrefix("/quit", currentValue) {
					m.textarea.SetValue("/quit")
					return m, nil
				}
			}
		}

		// Check if Enter is pressed (but not Alt+Enter or Ctrl+J)
		if msg.Type == tea.KeyEnter && !msg.Alt {
			content := strings.TrimSpace(m.textarea.Value())
			if content != "" {
				m.handleInput(content)
				m.textarea.Reset()
				m.textarea.SetHeight(1)          // Reset to 1 line after sending
				m.textarea.Focus()               // Re-focus after reset
				m.textarea.SetWidth(m.width - 4) // Reset width
				m.suggestion = ""                // Clear suggestion after sending
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()

				if strings.HasPrefix(content, "/init") {
					// Show immediate feedback
					m.addMessage("assistant", "🔍 Analyzing repository structure...")
					m.viewport.SetContent(m.renderMessages())
					m.viewport.GotoBottom()
					cmds = append(cmds, m.runInit())
				} else if content == "/quit" || content == "/exit" {
					return m, tea.Quit
				} else {
					cmds = append(cmds, m.sendMessage(content))
				}
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}

		// Pre-expand height BEFORE processing newline to prevent scrolling
		if msg.String() == "alt+enter" || msg.String() == "ctrl+j" {
			currentValue := m.textarea.Value()
			currentLines := strings.Count(currentValue, "\n") + 1
			// If we're about to add a newline from 1 line, immediately expand to 3
			if currentLines == 1 {
				m.textarea.SetHeight(3)
			} else if currentLines >= 3 {
				// For 3+ lines, set exact height for the new line
				newHeight := currentLines + 1
				if newHeight > 10 {
					newHeight = 10
				}
				m.textarea.SetHeight(newHeight)
			}
		}

		// Let textarea handle all other input (including Shift+Enter)
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		// Update height based on actual content
		currentValue := m.textarea.Value()
		lines := strings.Count(currentValue, "\n") + 1

		// Set height based on content
		// For 2 lines, ensure height is 3 to show both lines
		// For 3+ lines, use exact height
		desiredHeight := lines
		if lines == 2 {
			desiredHeight = 3 // Prevent scrolling for 2 lines
		}
		// For 3+ lines, use exact height (no extra)
		if desiredHeight > 10 {
			desiredHeight = 10 // Max 10 lines visible
			// When at max height with 11+ lines, ensure cursor is visible
			if lines > 10 {
				// The textarea should handle internal scrolling
				// Force a cursor position update to trigger scroll
				currentLine := m.textarea.Line()
				if currentLine >= 10 {
					// Cursor is beyond visible area, textarea should auto-scroll
				}
			}
		}

		// Only update height if it changed
		if m.textarea.Height() != desiredHeight {
			m.textarea.SetHeight(desiredHeight)
		}

		// Update command suggestions when typing /
		if strings.HasPrefix(currentValue, "/") && !strings.Contains(currentValue, " ") {
			// Show command suggestions
			if currentValue == "/" {
				m.suggestion = "Available: /init, /quit"
			} else if strings.HasPrefix("/init", currentValue) && currentValue != "/init" {
				m.suggestion = "Press Tab to complete: /init"
			} else if strings.HasPrefix("/quit", currentValue) && currentValue != "/quit" {
				m.suggestion = "Press Tab to complete: /quit"
			} else {
				m.suggestion = ""
			}
		} else {
			m.suggestion = ""
		}

	case aiResponseMsg:
		m.addMessage("assistant", msg.content)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case initCompleteMsg:
		if msg.err != nil {
			m.addMessage("error", fmt.Sprintf("Failed to analyze repository: %v", msg.err))
		} else {
			m.addMessage("assistant", "✅ Repository analyzed successfully! AGENTS.md has been created.")
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case errorMsg:
		m.err = msg.error
		m.addMessage("error", fmt.Sprintf("Error: %v", msg.error))
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}

	// Update viewport (textarea already updated above in KeyMsg handling)
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *SimpleModel) View() string {
	if !m.ready {
		return "\n  Initializing Rigel..."
	}

	// Use fixed heights for layout
	headerHeight := 3
	footerHeight := 1
	// Fixed input area height (minimum space to prevent layout shift)
	inputHeight := 2 // Minimal height (1 line + 1 for border)

	// Adjust viewport height based on current textarea size
	viewportHeight := m.height - headerHeight - footerHeight - inputHeight
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	// Header
	header := m.theme.HeaderStyle.
		Width(m.width).
		Align(lipgloss.Center).
		Render("🤖 Rigel AI Assistant")

	// Chat viewport
	chat := m.theme.BaseStyle.
		Width(m.width).
		Height(viewportHeight).
		Render(m.viewport.View())

	// Input area with suggestion
	textareaView := m.textarea.View()

	// Trim extra empty lines based on actual content
	actualContent := m.textarea.Value()
	actualLines := strings.Count(actualContent, "\n") + 1
	viewLines := strings.Split(textareaView, "\n")

	// When height is greater than actual lines AND we're not at max capacity, remove trailing empty lines
	// Don't trim if we have 10+ lines as the textarea needs to handle scrolling
	if actualLines < 10 && m.textarea.Height() > actualLines && len(viewLines) > actualLines {
		// Keep only the actual content lines
		textareaView = strings.Join(viewLines[:actualLines], "\n")
	}

	var inputContent string
	if m.suggestion != "" {
		suggestionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")).
			Italic(true)
		inputContent = textareaView + "\n" + suggestionStyle.Render(m.suggestion)
	} else {
		inputContent = textareaView
	}

	inputBox := m.theme.InputBoxStyle.
		Width(m.width - 2).
		Render(inputContent)

	// Footer with help
	footer := m.theme.HelpBarStyle.
		Width(m.width).
		Render(" /init: Analyze | Tab: Complete | Enter: Send | Alt+Enter: Newline | ?: Help")

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		chat,
		inputBox,
		footer,
	)
}

func (m *SimpleModel) addMessage(role, content string) {
	m.messages = append(m.messages, ChatMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})
}

func (m *SimpleModel) handleInput(input string) {
	// Don't add /init command to chat history
	if !strings.HasPrefix(input, "/init") {
		m.addMessage("user", input)
	}
}

func (m *SimpleModel) renderMessages() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		timeStr := msg.Time.Format("15:04")

		var roleStyle lipgloss.Style
		switch msg.Role {
		case "user":
			roleStyle = m.theme.RoleStyle.Foreground(m.theme.Primary)
		case "assistant":
			roleStyle = m.theme.RoleStyle.Foreground(m.theme.Secondary)
		case "system":
			roleStyle = m.theme.RoleStyle.Foreground(m.theme.Info)
		case "error":
			roleStyle = m.theme.ErrorStyle
		}

		role := roleStyle.Render(msg.Role)
		time := m.theme.TimeStyle.Render(timeStr)

		// Apply syntax highlighting if needed
		content := msg.Content
		if strings.Contains(content, "```") {
			content = applySyntaxHighlighting(content)
		}

		sb.WriteString(fmt.Sprintf("%s %s\n%s\n\n", role, time, content))
	}

	return sb.String()
}

func (m *SimpleModel) sendMessage(content string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		response, err := m.agent.Execute(ctx, content)
		if err != nil {
			return errorMsg{error: err}
		}

		return aiResponseMsg{content: response}
	}
}

func (m *SimpleModel) runInit() tea.Cmd {
	return func() tea.Msg {
		// The "Analyzing..." message is already shown before this runs
		ctx := context.Background()
		err := m.analyzer.AnalyzeRepository(ctx)

		return initCompleteMsg{err: err}
	}
}

// Message types
type aiResponseMsg struct {
	content string
}

type initCompleteMsg struct {
	err error
}

// Reuse errorMsg from model.go
