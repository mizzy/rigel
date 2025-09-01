package terminal

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/command"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/history"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/state"
)

// Model represents the main chat interface
type Model struct {
	config  *config.Config
	input   textarea.Model
	spinner spinner.Model

	chatState          *state.ChatState
	inputHistory       []string
	historyIndex       int
	currentInput       string
	quitting           bool
	completions        []string
	selectedCompletion int
	showCompletions    bool
	ctrlCPressed       bool
	infoMessage        string
	historyManager     *history.Manager // Add history manager
	llmState           *state.LLMState

	// Handlers
	completionHandler *command.CompletionHandler
}

// Exchange represents a single chat exchange - using state.Exchange
type Exchange = state.Exchange

// NewModel creates a new chat model instance
func NewModel(provider llm.Provider, cfg *config.Config) *Model {
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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true) // Same as prompt symbol

	// Initialize history manager
	histManager, err := history.NewManager()
	if err != nil {
		// If we can't create history manager, continue without it
		histManager = nil
	} else {
		// Load existing history
		_ = histManager.Load()
	}

	llmState := state.NewLLMState()
	if cfg != nil {
		llmState.SetCurrentProvider(provider)
	}

	m := &Model{
		config:            cfg,
		input:             ta,
		spinner:           s,
		chatState:         state.NewChatState(),
		inputHistory:      []string{},
		historyIndex:      -1,
		historyManager:    histManager,
		llmState:          llmState,
		completionHandler: command.NewCompletionHandler(),
	}

	// Load input history from manager if available
	if histManager != nil {
		m.inputHistory = histManager.GetCommands()
	}

	return m
}

// Init initializes the chat model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}
