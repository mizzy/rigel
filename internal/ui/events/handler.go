package events

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/ui/commands"
	"github.com/mizzy/rigel/internal/ui/state"
)

// Handler handles UI events
type Handler struct {
	chatState      *state.ChatState
	selectionState *state.SelectionState
	cmdHandler     *commands.Handler
}

// NewHandler creates a new event handler
func NewHandler(chatState *state.ChatState, selectionState *state.SelectionState, cmdHandler *commands.Handler) *Handler {
	return &Handler{
		chatState:      chatState,
		selectionState: selectionState,
		cmdHandler:     cmdHandler,
	}
}

// HandleKeyMessage handles keyboard input events
func (h *Handler) HandleKeyMessage(msg tea.KeyMsg, model tea.Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle provider selection mode
	if h.selectionState.ProviderSelectionMode {
		return h.handleProviderSelectionKey(msg, model)
	}

	// Handle model selection mode
	if h.selectionState.ModelSelectionMode {
		return h.handleModelSelectionKey(msg, model)
	}

	// Handle special keys first
	switch msg.Type {
	case tea.KeyCtrlC:
		if h.chatState.CtrlCPressed {
			h.chatState.Quitting = true
			return model, tea.Quit
		}
		h.chatState.CtrlCPressed = true
		h.chatState.SetInfoMessage("Press Ctrl+C again to exit")
		return model, nil

	case tea.KeyCtrlD:
		h.chatState.CtrlCPressed = false
		h.chatState.SetInfoMessage("")
		if !h.chatState.Thinking && h.chatState.Input.Value() == "" {
			h.chatState.Quitting = true
			return model, tea.Quit
		}
	default:
		h.chatState.CtrlCPressed = false
		h.chatState.SetInfoMessage("")
	}

	// Handle Tab key for completion
	if msg.String() == "tab" && !h.chatState.Thinking && h.chatState.ShowSuggestions {
		h.completeSuggestion()
		h.chatState.CtrlCPressed = false
		h.chatState.SetInfoMessage("")
		return model, nil
	}

	// Handle arrow keys for suggestion navigation or history navigation
	if !h.chatState.Thinking {
		switch msg.String() {
		case "up":
			if h.chatState.ShowSuggestions {
				if h.chatState.SelectedSuggestion > 0 {
					h.chatState.SelectedSuggestion--
				}
			} else {
				h.navigateHistory(-1)
			}
			h.chatState.CtrlCPressed = false
			h.chatState.SetInfoMessage("")
			return model, nil
		case "down":
			if h.chatState.ShowSuggestions {
				if h.chatState.SelectedSuggestion < len(h.chatState.Suggestions)-1 {
					h.chatState.SelectedSuggestion++
				}
			} else {
				h.navigateHistory(1)
			}
			h.chatState.CtrlCPressed = false
			h.chatState.SetInfoMessage("")
			return model, nil
		}
	}

	// Handle Enter key to submit
	if msg.String() == "enter" && !h.chatState.Thinking {
		h.chatState.CtrlCPressed = false
		h.chatState.SetInfoMessage("")
		return h.submitInput(model)
	}

	// Update textarea with key input
	if !h.chatState.Thinking && !h.selectionState.IsInSelectionMode() {
		h.chatState.Input, cmd = h.chatState.Input.Update(msg)
		cmds = append(cmds, cmd)

		// Update suggestions based on input
		h.updateSuggestions()
	}

	return model, tea.Batch(cmds...)
}

// HandleSpinnerMessage handles spinner tick messages
func (h *Handler) HandleSpinnerMessage(msg spinner.TickMsg, model tea.Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	h.chatState.Spinner, cmd = h.chatState.Spinner.Update(msg)
	return model, cmd
}

// HandleAIResponse handles AI response messages
func (h *Handler) HandleAIResponse(msg interface{}, model tea.Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.AIResponse:
		h.chatState.SetThinking(false)
		if msg.Err != nil {
			h.chatState.SetError(msg.Err)
		} else {
			exchange := state.Exchange{
				Prompt:   h.chatState.CurrentPrompt,
				Response: msg.Content,
			}
			h.chatState.AddToHistory(exchange)
			h.chatState.ClearError()
		}
		h.chatState.Input.SetValue("")
		h.clearSuggestions()
		return model, nil

	default:
		return model, nil
	}
}

// HandleSelectionMessages handles selection-related messages
func (h *Handler) HandleSelectionMessages(msg interface{}, model tea.Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.ModelSelectorMsg:
		h.selectionState.ModelSelectionMode = true
		h.selectionState.AvailableModels = msg.Models
		h.selectionState.FilteredModels = msg.Models
		h.selectionState.SelectedModelIndex = h.findCurrentModelIndex(msg.Models, msg.CurrentModel)
		return model, nil

	case commands.ProviderSelectorMsg:
		h.selectionState.ProviderSelectionMode = true
		h.selectionState.AvailableProviders = msg.Providers
		h.selectionState.SelectedProviderIndex = h.findCurrentProviderIndex(msg.Providers)
		return model, nil

	case commands.ProviderSwitchResponse:
		h.chatState.Provider = msg.Provider
		h.chatState.Config.Provider = msg.ProviderName
		h.selectionState.ClearSelectionModes()
		h.chatState.SetInfoMessage(fmt.Sprintf("Switched to provider: %s", msg.ProviderName))
		return model, nil

	default:
		return model, nil
	}
}

// submitInput handles input submission
func (h *Handler) submitInput(model tea.Model) (tea.Model, tea.Cmd) {
	prompt := strings.TrimSpace(h.chatState.Input.Value())
	if prompt == "" {
		return model, nil
	}

	h.chatState.CurrentPrompt = prompt
	h.chatState.AddToInputHistory(prompt)
	h.clearSuggestions()

	// Execute command or chat
	cmd := h.cmdHandler.Execute(prompt)
	return model, cmd
}

// Helper methods

func (h *Handler) completeSuggestion() {
	if len(h.chatState.Suggestions) > 0 && h.chatState.SelectedSuggestion >= 0 && h.chatState.SelectedSuggestion < len(h.chatState.Suggestions) {
		suggestion := h.chatState.Suggestions[h.chatState.SelectedSuggestion]
		currentValue := h.chatState.Input.Value()

		// Find the position where the suggestion should be inserted
		if strings.HasPrefix(suggestion, "/") {
			// If it's a command suggestion, replace the entire input
			h.chatState.Input.SetValue(suggestion + " ")
		} else {
			// For other suggestions, append to current input
			h.chatState.Input.SetValue(currentValue + suggestion)
		}

		h.clearSuggestions()
		h.chatState.Input.CursorEnd()
	}
}

func (h *Handler) clearSuggestions() {
	h.chatState.Suggestions = []string{}
	h.chatState.ShowSuggestions = false
	h.chatState.SelectedSuggestion = 0
}

func (h *Handler) updateSuggestions() {
	value := h.chatState.Input.Value()
	if strings.HasPrefix(value, "/") && !strings.Contains(value, " ") {
		// Command completion
		var suggestions []string
		for _, cmd := range commands.AvailableCommands() {
			if strings.HasPrefix(cmd.Name, value) {
				suggestions = append(suggestions, cmd.Name)
			}
		}
		h.chatState.Suggestions = suggestions
		h.chatState.ShowSuggestions = len(suggestions) > 0
		h.chatState.SelectedSuggestion = 0
	} else {
		h.clearSuggestions()
	}
}

func (h *Handler) navigateHistory(direction int) {
	if len(h.chatState.InputHistory) == 0 {
		return
	}

	// Store current input if we're starting to navigate
	if h.chatState.HistoryIndex == -1 {
		h.chatState.CurrentInput = h.chatState.Input.Value()
	}

	// Calculate new index
	newIndex := h.chatState.HistoryIndex + direction

	if newIndex >= len(h.chatState.InputHistory) {
		return
	}

	if newIndex < -1 {
		return
	}

	h.chatState.HistoryIndex = newIndex

	// Set input value
	if h.chatState.HistoryIndex == -1 {
		h.chatState.Input.SetValue(h.chatState.CurrentInput)
	} else {
		historyIndex := len(h.chatState.InputHistory) - 1 - h.chatState.HistoryIndex
		if historyIndex >= 0 && historyIndex < len(h.chatState.InputHistory) {
			h.chatState.Input.SetValue(h.chatState.InputHistory[historyIndex])
		}
	}

	h.chatState.Input.CursorEnd()
	h.clearSuggestions()
}

func (h *Handler) findCurrentModelIndex(models []llm.Model, currentModel string) int {
	for i, model := range models {
		if model.Name == currentModel {
			return i
		}
	}
	return 0
}

func (h *Handler) findCurrentProviderIndex(providers []string) int {
	currentProvider := h.chatState.Config.Provider
	for i, provider := range providers {
		if provider == currentProvider {
			return i
		}
	}
	return 0
}

// Selection key handlers
func (h *Handler) handleModelSelectionKey(msg tea.KeyMsg, model tea.Model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		h.selectionState.ClearSelectionModes()
		return model, nil
	case "up":
		if h.selectionState.SelectedModelIndex > 0 {
			h.selectionState.SelectedModelIndex--
		}
		return model, nil
	case "down":
		if h.selectionState.SelectedModelIndex < len(h.selectionState.FilteredModels)-1 {
			h.selectionState.SelectedModelIndex++
		}
		return model, nil
	case "enter":
		if h.selectionState.SelectedModelIndex < len(h.selectionState.FilteredModels) {
			selectedModel := h.selectionState.FilteredModels[h.selectionState.SelectedModelIndex]
			h.chatState.Config.Model = selectedModel.Name
			h.selectionState.ClearSelectionModes()
			h.chatState.SetInfoMessage(fmt.Sprintf("Model set to: %s", selectedModel.Name))
		}
		return model, nil
	default:
		// Handle filtering
		if len(msg.String()) == 1 {
			h.selectionState.ModelFilter += msg.String()
			h.filterModels()
		} else if msg.String() == "backspace" && len(h.selectionState.ModelFilter) > 0 {
			h.selectionState.ModelFilter = h.selectionState.ModelFilter[:len(h.selectionState.ModelFilter)-1]
			h.filterModels()
		}
		return model, nil
	}
}

func (h *Handler) handleProviderSelectionKey(msg tea.KeyMsg, model tea.Model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		h.selectionState.ClearSelectionModes()
		return model, nil
	case "up":
		if h.selectionState.SelectedProviderIndex > 0 {
			h.selectionState.SelectedProviderIndex--
		}
		return model, nil
	case "down":
		if h.selectionState.SelectedProviderIndex < len(h.selectionState.AvailableProviders)-1 {
			h.selectionState.SelectedProviderIndex++
		}
		return model, nil
	case "enter":
		if h.selectionState.SelectedProviderIndex < len(h.selectionState.AvailableProviders) {
			selectedProvider := h.selectionState.AvailableProviders[h.selectionState.SelectedProviderIndex]
			cmd := h.switchProvider(selectedProvider)
			return model, cmd
		}
		return model, nil
	}
	return model, nil
}

func (h *Handler) filterModels() {
	if h.selectionState.ModelFilter == "" {
		h.selectionState.FilteredModels = h.selectionState.AvailableModels
	} else {
		h.selectionState.FilteredModels = []llm.Model{}
		filter := strings.ToLower(h.selectionState.ModelFilter)
		for _, model := range h.selectionState.AvailableModels {
			if strings.Contains(strings.ToLower(model.Name), filter) {
				h.selectionState.FilteredModels = append(h.selectionState.FilteredModels, model)
			}
		}
	}
	if h.selectionState.SelectedModelIndex >= len(h.selectionState.FilteredModels) {
		h.selectionState.SelectedModelIndex = len(h.selectionState.FilteredModels) - 1
	}
	if h.selectionState.SelectedModelIndex < 0 {
		h.selectionState.SelectedModelIndex = 0
	}
}

func (h *Handler) switchProvider(providerName string) tea.Cmd {
	return func() tea.Msg {
		// This would need to be implemented to create the new provider
		// For now, return a placeholder
		return commands.ProviderSwitchResponse{
			Provider:     nil, // Would create new provider here
			ProviderName: providerName,
		}
	}
}
