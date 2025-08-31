package chat

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/ui/commands"
	"github.com/mizzy/rigel/internal/ui/events"
	"github.com/mizzy/rigel/internal/ui/state"
)

// RefactoredModel represents the refactored chat interface using the new architecture
type RefactoredModel struct {
	chatState      *state.ChatState
	selectionState *state.SelectionState
	eventHandler   *events.Handler
	cmdHandler     *commands.Handler
}

// NewRefactoredModel creates a new refactored chat model instance
func NewRefactoredModel(provider llm.Provider, cfg *config.Config) *RefactoredModel {
	// Create textarea with same configuration as original
	ta := textarea.New()
	ta.Placeholder = "Type a message or / for commands (Alt+Enter for new line)"
	ta.Focus()
	ta.CharLimit = 5000
	ta.SetWidth(100)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.EndOfBufferCharacter = ' '
	ta.KeyMap.InsertNewline.SetKeys("alt+enter", "ctrl+j")

	// Remove borders and styling
	noBorder := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false).
		BorderForeground(lipgloss.NoColor{}).
		Padding(0).
		Margin(0)

	ta.FocusedStyle.Base = noBorder
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()

	ta.BlurredStyle.Base = noBorder
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.EndOfBuffer = lipgloss.NewStyle()

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))

	// Create state objects
	chatState := state.NewChatState(provider, cfg, ta, s)
	selectionState := state.NewSelectionState()

	// Create handlers
	cmdHandler := commands.NewHandler(chatState, selectionState)
	eventHandler := events.NewHandler(chatState, selectionState, cmdHandler)

	return &RefactoredModel{
		chatState:      chatState,
		selectionState: selectionState,
		eventHandler:   eventHandler,
		cmdHandler:     cmdHandler,
	}
}

// Init initializes the refactored chat model
func (m RefactoredModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.chatState.Spinner.Tick,
	)
}

// Update handles messages and updates the model using the new event handler
func (m RefactoredModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.eventHandler.HandleKeyMessage(msg, m)

	case spinner.TickMsg:
		return m.eventHandler.HandleSpinnerMessage(msg, m)

	case commands.AIResponse, commands.ModelSelectorMsg, commands.ProviderSelectorMsg, commands.ProviderSwitchResponse:
		if isSelectionMessage(msg) {
			return m.eventHandler.HandleSelectionMessages(msg, m)
		}
		return m.eventHandler.HandleAIResponse(msg, m)

	default:
		// Handle other message types as needed
		return m, nil
	}
}

// View renders the refactored model - this will delegate to the existing view logic
func (m RefactoredModel) View() string {
	// For now, create a legacy model for rendering
	legacyModel := &Model{
		provider:              m.chatState.Provider,
		config:                m.chatState.Config,
		input:                 m.chatState.Input,
		spinner:               m.chatState.Spinner,
		history:               convertExchanges(m.chatState.History),
		inputHistory:          m.chatState.InputHistory,
		historyIndex:          m.chatState.HistoryIndex,
		currentInput:          m.chatState.CurrentInput,
		thinking:              m.chatState.Thinking,
		currentPrompt:         m.chatState.CurrentPrompt,
		err:                   m.chatState.Err,
		quitting:              m.chatState.Quitting,
		suggestions:           m.chatState.Suggestions,
		selectedSuggestion:    m.chatState.SelectedSuggestion,
		showSuggestions:       m.chatState.ShowSuggestions,
		ctrlCPressed:          m.chatState.CtrlCPressed,
		infoMessage:           m.chatState.InfoMessage,
		historyManager:        m.chatState.HistoryManager,
		modelSelectionMode:    m.selectionState.ModelSelectionMode,
		availableModels:       m.selectionState.AvailableModels,
		filteredModels:        m.selectionState.FilteredModels,
		selectedModelIndex:    m.selectionState.SelectedModelIndex,
		modelFilter:           m.selectionState.ModelFilter,
		providerSelectionMode: m.selectionState.ProviderSelectionMode,
		availableProviders:    m.selectionState.AvailableProviders,
		selectedProviderIndex: m.selectionState.SelectedProviderIndex,
	}

	return legacyModel.View()
}

// Helper functions

func convertExchanges(newExchanges []state.Exchange) []Exchange {
	legacyExchanges := make([]Exchange, len(newExchanges))
	for i, exchange := range newExchanges {
		legacyExchanges[i] = Exchange{
			Prompt:   exchange.Prompt,
			Response: exchange.Response,
		}
	}
	return legacyExchanges
}

func isSelectionMessage(msg interface{}) bool {
	switch msg.(type) {
	case commands.ModelSelectorMsg, commands.ProviderSelectorMsg, commands.ProviderSwitchResponse:
		return true
	default:
		return false
	}
}
