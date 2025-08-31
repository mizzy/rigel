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
)

// Model represents the main chat interface
type Model struct {
	provider llm.Provider
	config   *config.Config
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
	completions        []string
	selectedCompletion int
	showCompletions    bool
	ctrlCPressed       bool
	infoMessage        string
	historyManager     *history.Manager // Add history manager

	// Model selection mode
	modelSelectionMode bool
	availableModels    []llm.Model
	filteredModels     []llm.Model
	selectedModelIndex int
	modelFilter        string

	// Provider selection mode
	providerSelectionMode bool
	availableProviders    []string
	selectedProviderIndex int

	// Handlers
	completionHandler *command.CompletionHandler
	commandHandler    *command.Handler
}

// Exchange represents a single chat exchange
type Exchange struct {
	Prompt   string
	Response string
}

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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("87")) // Match Rigel blue

	// Initialize history manager
	histManager, err := history.NewManager()
	if err != nil {
		// If we can't create history manager, continue without it
		histManager = nil
	} else {
		// Load existing history
		_ = histManager.Load()
	}

	m := &Model{
		provider:          provider,
		config:            cfg,
		input:             ta,
		spinner:           s,
		history:           []Exchange{},
		inputHistory:      []string{},
		historyIndex:      -1,
		historyManager:    histManager,
		completionHandler: command.NewCompletionHandler(),
		commandHandler:    command.NewHandler(),
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
