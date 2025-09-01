package handlers

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/state"
)

// AIResponse represents the response from AI processing
type AIResponse struct {
	Content string
	Error   error
}

// RequestResponse sends a request to the LLM provider with conversation history
func RequestResponse(prompt string, llmState *state.LLMState, chatState *state.ChatState) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Build message history from chat exchanges
		history := chatState.GetHistory()
		messages := make([]llm.Message, 0, len(history)*2+1)

		// Add previous exchanges to maintain context
		for _, exchange := range history {
			messages = append(messages, llm.Message{
				Role:    "user",
				Content: exchange.Prompt,
			})
			messages = append(messages, llm.Message{
				Role:    "assistant",
				Content: exchange.Response,
			})
		}

		// Add current prompt
		messages = append(messages, llm.Message{
			Role:    "user",
			Content: prompt,
		})

		// Use GenerateWithHistory to send full conversation context
		provider := llmState.GetCurrentProvider()
		if provider == nil {
			return AIResponse{Error: fmt.Errorf("no provider available")}
		}
		response, err := provider.GenerateWithHistory(ctx, messages, llm.GenerateOptions{})
		if err != nil {
			return AIResponse{Error: err}
		}
		return AIResponse{Content: strings.TrimSpace(response)}
	}
}

// ProviderSwitchResponse represents a provider switch response
type ProviderSwitchResponse struct {
	Provider     llm.Provider
	ProviderName string
}

// SelectionResult represents the result of a selection key handling
type SelectionResult struct {
	ShouldExit   bool
	ShouldSwitch bool
	SwitchCmd    tea.Cmd
	InputValue   string
	Placeholder  string
}

// HandleProviderSelectionKey handles key input during provider selection
func HandleProviderSelectionKey(msg tea.KeyMsg, llmState *state.LLMState, chatState *state.ChatState, cfg *config.Config) SelectionResult {
	switch msg.Type {
	case tea.KeyEsc:
		llmState.DeactivateProviderSelection()
		chatState.SetThinking(false)
		return SelectionResult{ShouldExit: true}

	case tea.KeyEnter:
		if provider, ok := llmState.GetSelectedProvider(); ok {
			llmState.DeactivateProviderSelection()
			chatState.SetThinking(false)
			cmd := CreateProviderSwitchCommand(provider, cfg)
			return SelectionResult{ShouldExit: true, ShouldSwitch: true, SwitchCmd: cmd}
		}
		return SelectionResult{}

	case tea.KeyUp:
		llmState.MoveProviderSelectionUp()
		return SelectionResult{}

	case tea.KeyDown:
		llmState.MoveProviderSelectionDown()
		return SelectionResult{}

	default:
		return SelectionResult{}
	}
}

// HandleModelSelectionKey handles key input during model selection
func HandleModelSelectionKey(msg tea.KeyMsg, llmState *state.LLMState, chatState *state.ChatState, input InputUpdater) SelectionResult {
	switch msg.Type {
	case tea.KeyEsc:
		llmState.DeactivateModelSelection()
		chatState.SetThinking(false)
		return SelectionResult{
			ShouldExit:  true,
			InputValue:  "",
			Placeholder: "Type a message or / for commands (Alt+Enter for new line)",
		}

	case tea.KeyEnter:
		if model, ok := llmState.GetSelectedModel(); ok {
			llmState.DeactivateModelSelection()
			chatState.SetThinking(false)
			cmd := CreateModelSwitchCommand(model, llmState)
			return SelectionResult{
				ShouldExit:   true,
				ShouldSwitch: true,
				SwitchCmd:    cmd,
				InputValue:   "",
				Placeholder:  "Type a message or / for commands (Alt+Enter for new line)",
			}
		}
		return SelectionResult{}

	case tea.KeyUp:
		llmState.MoveModelSelectionUp()
		return SelectionResult{}

	case tea.KeyDown:
		llmState.MoveModelSelectionDown()
		return SelectionResult{}

	case tea.KeyBackspace:
		// Handle backspace for filtering
		currentFilter := llmState.GetModelFilter()
		if len(currentFilter) > 0 {
			newFilter := currentFilter[:len(currentFilter)-1]
			llmState.SetModelFilter(newFilter)
			return SelectionResult{InputValue: newFilter}
		}
		return SelectionResult{}

	case tea.KeyRunes:
		// Handle printable characters for filtering
		runes := msg.Runes
		if len(runes) > 0 {
			// Ignore non-printable characters
			if runes[0] < 32 || runes[0] > 126 {
				return SelectionResult{}
			}

			// Add to filter
			currentFilter := llmState.GetModelFilter()
			newFilter := currentFilter + string(runes)
			llmState.SetModelFilter(newFilter)
			return SelectionResult{InputValue: newFilter}
		}
		return SelectionResult{}

	default:
		return SelectionResult{}
	}
}

// CreateProviderSwitchCommand creates a command to switch providers
func CreateProviderSwitchCommand(provider llm.Provider, cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		// Update config
		if cfg != nil {
			cfg.Provider = provider.GetName()
		}

		return ProviderSwitchResponse{
			Provider:     provider,
			ProviderName: provider.GetName(),
		}
	}
}

// CreateModelSwitchCommand creates a command to switch models
func CreateModelSwitchCommand(model llm.Model, llmState *state.LLMState) tea.Cmd {
	// Actually switch the model
	provider := llmState.GetCurrentProvider()
	if provider != nil {
		provider.SetModel(model)
		llmState.SetCurrentModel(model)
	}

	return func() tea.Msg {
		return AIResponse{
			Content: fmt.Sprintf("Switched to model: %s", model.Name),
		}
	}
}
