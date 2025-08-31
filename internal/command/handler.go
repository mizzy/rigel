package command

import (
	"fmt"
	"strings"
)

// Handler handles command processing
type Handler struct{}

// NewHandler creates a new command handler
func NewHandler() *Handler {
	return &Handler{}
}

// HandleCommand processes a command and returns the result
func (h *Handler) HandleCommand(command string, cmdContext CommandContext) Result {
	switch command {
	case "/init":
		return h.analyzeRepository(cmdContext)

	case "/model":
		return h.showModelSelector(cmdContext)

	case "/provider":
		return h.showProviderSelector(cmdContext)

	case "/status":
		return h.showStatus(cmdContext)

	case "/help":
		return h.showHelp()

	case "/clear":
		return h.clearChatHistory(cmdContext)

	case "/clearhistory":
		return h.clearCommandHistory(cmdContext)

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
