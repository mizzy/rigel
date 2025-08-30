package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/agent"
	"github.com/mizzy/rigel/internal/llm"
)

type sessionState int

const (
	stateChat sessionState = iota
	stateFileExplorer
	stateCommandPalette
	stateHelp
)

type Model struct {
	agent    *agent.Agent
	provider llm.Provider
	state    sessionState
	width    int
	height   int
	ready    bool

	// Chat components
	textarea textarea.Model
	viewport viewport.Model
	messages []Message

	// UI components
	sidebar   *Sidebar
	statusbar *StatusBar
	help      help.Model
	keys      keyMap

	// File explorer
	fileExplorer *FileExplorer

	// Command palette
	commandPalette *CommandPalette

	// Theme
	theme *Theme

	// History
	history []HistoryEntry

	// Error message
	err error
}

type Message struct {
	Role    string
	Content string
	Time    string
}

type HistoryEntry struct {
	Input  string
	Output string
	Time   string
}

type keyMap struct {
	Up             key.Binding
	Down           key.Binding
	Enter          key.Binding
	Tab            key.Binding
	ShiftTab       key.Binding
	Help           key.Binding
	Quit           key.Binding
	FileExplorer   key.Binding
	CommandPalette key.Binding
	ClearChat      key.Binding
	SaveSession    key.Binding
}

func NewModel(provider llm.Provider) *Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message here... (Ctrl+Enter to send)"
	ta.CharLimit = 10000
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to Rigel! Start typing to chat with the AI assistant.")

	a := agent.New(provider)

	m := &Model{
		agent:    a,
		provider: provider,
		state:    stateChat,
		textarea: ta,
		viewport: vp,
		messages: []Message{},
		help:     help.New(),
		theme:    NewDefaultTheme(),
		history:  []HistoryEntry{},
	}

	m.keys = keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "ctrl+p"),
			key.WithHelp("↑/ctrl+p", "previous"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "ctrl+n"),
			key.WithHelp("↓/ctrl+n", "next"),
		),
		Enter: key.NewBinding(
			key.WithKeys("ctrl+enter"),
			key.WithHelp("ctrl+enter", "send message"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus next"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "focus previous"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "ctrl+h"),
			key.WithHelp("?/ctrl+h", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c/ctrl+q", "quit"),
		),
		FileExplorer: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "file explorer"),
		),
		CommandPalette: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "command palette"),
		),
		ClearChat: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear chat"),
		),
		SaveSession: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save session"),
		),
	}

	m.sidebar = NewSidebar(m.theme)
	m.statusbar = NewStatusBar(m.theme)
	m.fileExplorer = NewFileExplorer(m.theme)
	m.commandPalette = NewCommandPalette(m.theme)

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.fileExplorer.Init(),
		m.commandPalette.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(m.width-m.sidebar.Width()-2, m.height-m.statusbar.Height()-m.textarea.Height()-4)
			m.viewport.YPosition = 1
			m.viewport.SetContent(m.renderMessages())
			m.ready = true
		} else {
			m.viewport.Width = m.width - m.sidebar.Width() - 2
			m.viewport.Height = m.height - m.statusbar.Height() - m.textarea.Height() - 4
		}

		m.textarea.SetWidth(m.width - m.sidebar.Width() - 2)

	case tea.KeyMsg:
		switch m.state {
		case stateChat:
			cmds = append(cmds, m.handleChatKeys(msg)...)
		case stateFileExplorer:
			cmds = append(cmds, m.handleFileExplorerKeys(msg)...)
		case stateCommandPalette:
			cmds = append(cmds, m.handleCommandPaletteKeys(msg)...)
		case stateHelp:
			cmds = append(cmds, m.handleHelpKeys(msg)...)
		}

	case responseMsg:
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: msg.content,
			Time:    msg.time,
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case errorMsg:
		m.err = msg.error

	case clearChatMsg:
		m.messages = []Message{}
		m.viewport.SetContent("Chat cleared. Start a new conversation!")
	}

	// Update components
	if m.state == stateChat {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleChatKeys(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, m.keys.Enter):
		content := m.textarea.Value()
		if strings.TrimSpace(content) != "" {
			m.messages = append(m.messages, Message{
				Role:    "user",
				Content: content,
				Time:    getCurrentTime(),
			})
			m.textarea.Reset()
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			cmds = append(cmds, m.sendMessage(content))
		}

	case key.Matches(msg, m.keys.FileExplorer):
		m.state = stateFileExplorer

	case key.Matches(msg, m.keys.CommandPalette):
		m.state = stateCommandPalette

	case key.Matches(msg, m.keys.Help):
		m.state = stateHelp

	case key.Matches(msg, m.keys.ClearChat):
		cmds = append(cmds, clearChat())

	case key.Matches(msg, m.keys.Quit):
		cmds = append(cmds, tea.Quit)
	}

	return cmds
}

func (m *Model) handleFileExplorerKeys(msg tea.KeyMsg) []tea.Cmd {
	if msg.String() == "esc" {
		m.state = stateChat
		return nil
	}

	// Handle file explorer specific keys
	// Implementation depends on FileExplorer component
	return nil
}

func (m *Model) handleCommandPaletteKeys(msg tea.KeyMsg) []tea.Cmd {
	if msg.String() == "esc" {
		m.state = stateChat
		return nil
	}

	// Handle command palette specific keys
	// Implementation depends on CommandPalette component
	return nil
}

func (m *Model) handleHelpKeys(msg tea.KeyMsg) []tea.Cmd {
	if msg.String() == "esc" || key.Matches(msg, m.keys.Help) {
		m.state = stateChat
	}
	return nil
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var content string

	switch m.state {
	case stateChat:
		content = m.renderChatView()
	case stateFileExplorer:
		content = m.renderFileExplorerView()
	case stateCommandPalette:
		content = m.renderCommandPaletteView()
	case stateHelp:
		content = m.renderHelpView()
	}

	// Add status bar at the bottom
	content = lipgloss.JoinVertical(
		lipgloss.Top,
		content,
		m.statusbar.Render(m.width, m.state, len(m.messages)),
	)

	return content
}

func (m *Model) renderChatView() string {
	// Main chat area with sidebar
	chatArea := lipgloss.JoinVertical(
		lipgloss.Top,
		m.theme.HeaderStyle.Render("💬 Chat"),
		m.viewport.View(),
		m.theme.InputBoxStyle.Render(m.textarea.View()),
	)

	if m.err != nil {
		errorMsg := m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		chatArea = lipgloss.JoinVertical(
			lipgloss.Top,
			errorMsg,
			chatArea,
		)
	}

	// Join sidebar and chat area horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.sidebar.Render(m.height-m.statusbar.Height()),
		chatArea,
	)
}

func (m *Model) renderFileExplorerView() string {
	return m.fileExplorer.Render(m.width, m.height-m.statusbar.Height())
}

func (m *Model) renderCommandPaletteView() string {
	return m.commandPalette.Render(m.width, m.height-m.statusbar.Height())
}

func (m *Model) renderHelpView() string {
	helpText := `
╔══════════════════════════════════════════════════════════════════╗
║                         Rigel Help                               ║
╚══════════════════════════════════════════════════════════════════╝

🎮 Navigation:
  • Ctrl+Enter    Send message
  • Ctrl+E        Open file explorer
  • Ctrl+P        Open command palette
  • Ctrl+L        Clear chat
  • Ctrl+S        Save session
  • Tab           Focus next element
  • Shift+Tab     Focus previous element
  • ?/Ctrl+H      Toggle this help
  • Ctrl+C/Ctrl+Q Quit

📝 Chat Features:
  • Multiline input supported
  • Code syntax highlighting
  • Markdown rendering
  • Session history

🛠️ Commands:
  Type "/" in the command palette to see available commands

Press ESC to return to chat
`
	return m.theme.HelpStyle.Render(helpText)
}

func (m *Model) renderMessages() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		role := m.theme.RoleStyle.Render(msg.Role)
		time := m.theme.TimeStyle.Render(msg.Time)
		content := msg.Content

		// Apply syntax highlighting if needed
		if strings.Contains(content, "```") {
			content = applySyntaxHighlighting(content)
		}

		sb.WriteString(fmt.Sprintf("%s %s\n%s\n\n", role, time, content))
	}

	return sb.String()
}

func (m *Model) sendMessage(content string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		response, err := m.agent.Execute(ctx, content)
		if err != nil {
			return errorMsg{error: err}
		}

		return responseMsg{
			content: response,
			time:    getCurrentTime(),
		}
	}
}

// Message types
type responseMsg struct {
	content string
	time    string
}

type errorMsg struct {
	error error
}

type clearChatMsg struct{}

func clearChat() tea.Cmd {
	return func() tea.Msg {
		return clearChatMsg{}
	}
}

func getCurrentTime() string {
	return "now" // Simplified for now
}
