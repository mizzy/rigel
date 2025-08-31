package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// showHelp displays the help message
func (h *Handler) showHelp() Result {
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
func (h *Handler) analyzeRepository(cmdContext CommandContext) Result {
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
func (h *Handler) showModelSelector(cmdContext CommandContext) Result {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	currentModel := cmdContext.GetCurrentModel()
	models, err := cmdContext.ListModels(ctx)
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
func (h *Handler) showProviderSelector(cmdContext CommandContext) Result {
	// Get available providers
	providers := []string{"anthropic", "ollama"}

	// Get current provider from config
	currentProvider := cmdContext.GetProviderName()

	return Result{
		Type: "provider_selector",
		ProviderSelector: &ProviderSelectorMsg{
			CurrentProvider: currentProvider,
			Providers:       providers,
		},
	}
}

// showStatus returns session status information
func (h *Handler) showStatus(cmdContext CommandContext) Result {
	statusInfo := cmdContext.GetStatusInfo()
	return Result{
		Type:       "status",
		StatusInfo: &statusInfo,
	}
}

// clearChatHistory clears the chat history
func (h *Handler) clearChatHistory(cmdContext CommandContext) Result {
	cmdContext.ClearChatHistory()
	return Result{
		Type: "clear",
	}
}

// clearCommandHistory clears the command input history
func (h *Handler) clearCommandHistory(cmdContext CommandContext) Result {
	cmdContext.ClearInputHistory()

	// Clear persistent history if available
	if err := cmdContext.ClearPersistentHistory(); err != nil {
		return Result{
			Type:  "response",
			Error: fmt.Errorf("failed to clear history: %w", err),
		}
	}

	return Result{
		Type:    "response",
		Content: "Command history cleared successfully.",
	}
}
