package state

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/history"
	"github.com/mizzy/rigel/internal/llm"
)

// Exchange represents a single chat exchange
type Exchange struct {
	Prompt   string
	Response string
}

// ChatState holds the core chat state
type ChatState struct {
	// Core components
	Provider llm.Provider
	Config   *config.Config
	Input    textarea.Model
	Spinner  spinner.Model

	// Chat history
	History        []Exchange
	InputHistory   []string
	HistoryIndex   int
	CurrentInput   string
	HistoryManager *history.Manager

	// UI state
	Thinking      bool
	CurrentPrompt string
	Err           error
	Quitting      bool
	CtrlCPressed  bool
	InfoMessage   string

	// Suggestions
	Suggestions        []string
	SelectedSuggestion int
	ShowSuggestions    bool
}

// SelectionState holds model and provider selection state
type SelectionState struct {
	// Model selection
	ModelSelectionMode bool
	AvailableModels    []llm.Model
	FilteredModels     []llm.Model
	SelectedModelIndex int
	ModelFilter        string

	// Provider selection
	ProviderSelectionMode bool
	AvailableProviders    []string
	SelectedProviderIndex int
}

// NewChatState creates a new chat state instance
func NewChatState(provider llm.Provider, cfg *config.Config, input textarea.Model, spinner spinner.Model) *ChatState {
	// Initialize history manager
	histManager, err := history.NewManager()
	if err != nil {
		histManager = nil
	} else {
		_ = histManager.Load()
	}

	state := &ChatState{
		Provider:       provider,
		Config:         cfg,
		Input:          input,
		Spinner:        spinner,
		History:        []Exchange{},
		InputHistory:   []string{},
		HistoryIndex:   -1,
		HistoryManager: histManager,
	}

	// Load input history from manager if available
	if histManager != nil {
		state.InputHistory = histManager.GetCommands()
	}

	return state
}

// NewSelectionState creates a new selection state instance
func NewSelectionState() *SelectionState {
	return &SelectionState{
		SelectedModelIndex:    0,
		SelectedProviderIndex: 0,
	}
}

// IsInSelectionMode returns true if any selection mode is active
func (s *SelectionState) IsInSelectionMode() bool {
	return s.ModelSelectionMode || s.ProviderSelectionMode
}

// ClearSelectionModes clears all selection modes
func (s *SelectionState) ClearSelectionModes() {
	s.ModelSelectionMode = false
	s.ProviderSelectionMode = false
	s.ModelFilter = ""
}

// AddToHistory adds a new exchange to the chat history
func (c *ChatState) AddToHistory(exchange Exchange) {
	c.History = append(c.History, exchange)
}

// AddToInputHistory adds a command to the input history
func (c *ChatState) AddToInputHistory(command string) {
	if c.HistoryManager != nil {
		_ = c.HistoryManager.Add(command)
		_ = c.HistoryManager.Save()
		c.InputHistory = c.HistoryManager.GetCommands()
	}
	c.HistoryIndex = -1
}

// SetThinking updates the thinking state
func (c *ChatState) SetThinking(thinking bool) {
	c.Thinking = thinking
}

// SetError sets an error state
func (c *ChatState) SetError(err error) {
	c.Err = err
}

// ClearError clears the error state
func (c *ChatState) ClearError() {
	c.Err = nil
}

// SetInfoMessage sets an info message
func (c *ChatState) SetInfoMessage(message string) {
	c.InfoMessage = message
}
