package command

import (
	"fmt"
	"strings"

	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/history"
	"github.com/mizzy/rigel/internal/state"
)

// HandleCommand processes a command and returns the result
// This function is stateless and doesn't need a Handler struct
func HandleCommand(command string, llmState *state.LLMState, chatState *state.ChatState, cfg *config.Config, historyManager *history.Manager, inputHistory []string) Result {
	// Only treat as command if it starts with / without any leading whitespace
	if !strings.HasPrefix(command, "/") {
		return Result{
			Type:   "request",
			Prompt: command,
		}
	}

	// Use trimmed version for command processing
	command = strings.TrimSpace(command)
	switch command {
	case "/init":
		return analyzeRepository(llmState)

	case "/model":
		return showModelSelector(llmState)

	case "/provider":
		return showProviderSelector(llmState)

	case "/status":
		return showStatus(llmState, chatState, cfg, historyManager, inputHistory)

	case "/help":
		return showHelp()

	case "/clear":
		return clearChatHistory(chatState)

	case "/clearhistory":
		return clearCommandHistory(historyManager)

	case "/exit", "/quit":
		return Result{Type: "quit"}

	default:
		if strings.HasPrefix(command, "/") {
			return Result{
				Type:  "response",
				Error: fmt.Errorf("unknown command: %s, type /help for available commands", command),
			}
		}
		return Result{
			Type:   "request",
			Prompt: command,
		}
	}
}
