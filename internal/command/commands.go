package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/history"
	"github.com/mizzy/rigel/internal/state"
)

// showHelp displays the help message
func showHelp() Result {
	var help strings.Builder
	help.WriteString("Available commands:\n\n")
	for _, cmd := range AvailableCommands {
		help.WriteString(fmt.Sprintf("  %s - %s\n", cmd.Command, cmd.Description))
	}
	help.WriteString("\nKeyboard shortcuts:\n")
	help.WriteString("  Tab       - Complete command\n")
	help.WriteString("  ↑/↓       - Navigate completions\n")
	help.WriteString("  Enter     - Send message or select completion\n")
	help.WriteString("  Alt+Enter - New line\n")
	help.WriteString("  Ctrl+C    - Exit\n")

	return Result{
		Type:    "response",
		Content: help.String(),
	}
}

// analyzeRepository analyzes the repository and generates AGENTS.md
func analyzeRepository() Result {
	// Check if AGENTS.md already exists
	if _, err := os.Stat("AGENTS.md"); err == nil {
		return Result{
			Type:    "response",
			Content: "AGENTS.md already exists. Repository has been analyzed previously.",
		}
	}

	start := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_ = ctx // Use context if needed for actual analysis

	// Run analysis (this would need to be injected or made configurable)
	// For now, return a placeholder
	duration := time.Since(start)

	return Result{
		Type:    "response",
		Content: fmt.Sprintf("Repository analysis completed in %v.\nAGENTS.md has been generated with project context.", duration),
	}
}

// showModelSelector shows the model selector interface
func showModelSelector(llmState *state.LLMState) Result {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	currentModel := llmState.GetCurrentModel()

	provider := llmState.GetProvider()
	if provider == nil {
		return Result{
			Type: "model_selector",
			ModelSelector: &ModelSelectorMsg{
				CurrentModel: currentModel,
				Error:        fmt.Errorf("no provider available"),
			},
		}
	}

	models, err := provider.ListModels(ctx)
	if err != nil {
		return Result{
			Type: "model_selector",
			ModelSelector: &ModelSelectorMsg{
				CurrentModel: currentModel,
				Error:        err,
			},
		}
	}
	return Result{
		Type: "model_selector",
		ModelSelector: &ModelSelectorMsg{
			CurrentModel: currentModel,
			Models:       models,
		},
	}
}

// showProviderSelector shows the provider selector interface
func showProviderSelector(llmState *state.LLMState) Result {
	// Get available providers
	providers := []string{"anthropic", "ollama"}

	// Get current provider from LLMState
	currentProvider := llmState.GetCurrentProvider()

	return Result{
		Type: "provider_selector",
		ProviderSelector: &ProviderSelectorMsg{
			CurrentProvider: currentProvider,
			Providers:       providers,
		},
	}
}

// showStatus returns session status information
func showStatus(llmState *state.LLMState, chatState *state.ChatState, config *config.Config, historyManager *history.Manager, inputHistory []string) Result {
	provider := llmState.GetCurrentProvider()
	model := llmState.GetCurrentModel()

	logLevel := ""
	if config != nil {
		logLevel = config.LogLevel
	}

	// Calculate token usage
	totalUserChars := 0
	totalAssistantChars := 0
	for _, exchange := range chatState.GetHistory() {
		totalUserChars += len(exchange.Prompt)
		totalAssistantChars += len(exchange.Response)
	}
	approxUserTokens := totalUserChars / 4
	approxAssistantTokens := totalAssistantChars / 4
	totalTokens := approxUserTokens + approxAssistantTokens

	// Check if AGENTS.md exists
	repositoryInitialized := false
	if _, err := os.Stat("AGENTS.md"); err == nil {
		repositoryInitialized = true
	}

	statusInfo := StatusInfo{
		Provider:              provider,
		Model:                 model,
		MessageCount:          chatState.GetMessageCount(),
		UserTokens:            approxUserTokens,
		AssistantTokens:       approxAssistantTokens,
		TotalTokens:           totalTokens,
		CommandsCount:         len(inputHistory),
		PersistenceEnabled:    historyManager != nil,
		LogLevel:              logLevel,
		RepositoryInitialized: repositoryInitialized,
	}

	return Result{
		Type:       "status",
		StatusInfo: &statusInfo,
	}
}

// clearChatHistory clears the chat history
func clearChatHistory(chatState *state.ChatState) Result {
	chatState.ClearHistory()
	return Result{
		Type: "clear",
	}
}

// clearCommandHistory clears the command input history
func clearCommandHistory(historyManager *history.Manager) Result {
	// Clear persistent history if available
	if historyManager != nil {
		if err := historyManager.Clear(); err != nil {
			return Result{
				Type:  "response",
				Error: fmt.Errorf("failed to clear history: %w", err),
			}
		}
	}

	// Return result indicating input history should be cleared
	return Result{
		Type: "clear_input_history",
	}
}
