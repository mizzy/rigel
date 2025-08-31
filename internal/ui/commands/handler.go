package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/analyzer"
	"github.com/mizzy/rigel/internal/ui/state"
)

// Command represents a chat command
type Command struct {
	Name        string
	Description string
}

// Handler handles command execution
type Handler struct {
	chatState      *state.ChatState
	selectionState *state.SelectionState
}

// NewHandler creates a new command handler
func NewHandler(chatState *state.ChatState, selectionState *state.SelectionState) *Handler {
	return &Handler{
		chatState:      chatState,
		selectionState: selectionState,
	}
}

// AvailableCommands returns the list of available commands
func AvailableCommands() []Command {
	return []Command{
		{"/init", "Analyze repository and generate AGENTS.md"},
		{"/model", "Show current model and select from available models"},
		{"/provider", "Switch between LLM providers (Anthropic, Ollama, etc.)"},
		{"/status", "Show current session status and configuration"},
		{"/help", "Show available commands"},
		{"/clear", "Clear chat history"},
		{"/clearhistory", "Clear command history"},
		{"/exit", "Exit the application"},
		{"/quit", "Exit the application"},
	}
}

// Execute processes commands and returns the appropriate tea.Cmd
func (h *Handler) Execute(trimmedPrompt string) tea.Cmd {
	switch trimmedPrompt {
	case "/init":
		return h.analyzeRepository()
	case "/model":
		return h.showModelSelector()
	case "/provider":
		return h.showProviderSelector()
	case "/status":
		return h.showStatus()
	case "/help":
		return h.showHelp()
	case "/clear":
		h.chatState.History = []state.Exchange{}
		h.chatState.SetThinking(false)
		return nil
	case "/clearhistory":
		return h.clearHistory()
	case "/exit", "/quit":
		h.chatState.Quitting = true
		return tea.Quit
	default:
		if strings.HasPrefix(trimmedPrompt, "/") {
			h.chatState.SetError(fmt.Errorf("unknown command: %s, type /help for available commands", trimmedPrompt))
			h.chatState.SetThinking(false)
			return nil
		}
		return h.requestResponse()
	}
}

// analyzeRepository handles /init command
func (h *Handler) analyzeRepository() tea.Cmd {
	return func() tea.Msg {
		h.chatState.SetThinking(true)

		// Create analyzer (uses current directory by default)
		a := analyzer.NewRepoAnalyzer(h.chatState.Provider)
		analysis, err := a.Analyze()
		if err != nil {
			return fmt.Errorf("failed to analyze repository: %w", err)
		}

		// Generate AGENTS.md using LLM
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		prompt := fmt.Sprintf(`Please analyze this repository structure and code patterns, then create a comprehensive AGENTS.md file that documents the project:

%s

Create an AGENTS.md file that includes:
1. Project overview and purpose
2. Architecture and key components
3. Development patterns and conventions
4. Important files and their roles
5. Setup and development instructions
6. Any other information that would help someone understand this codebase

Format it as a proper markdown file.`, analysis)

		response, err := h.chatState.Provider.Generate(ctx, prompt)
		if err != nil {
			return fmt.Errorf("failed to generate AGENTS.md: %w", err)
		}

		// Write AGENTS.md
		if err := os.WriteFile("AGENTS.md", []byte(response), 0644); err != nil {
			return fmt.Errorf("failed to write AGENTS.md: %w", err)
		}

		// Add to chat history
		exchange := state.Exchange{
			Prompt:   "/init",
			Response: fmt.Sprintf("%s\n\n✅ AGENTS.md has been generated and saved to the current directory.", response),
		}

		return AIResponse{
			Content: exchange.Response,
			Err:     nil,
		}
	}
}

// showModelSelector handles /model command
func (h *Handler) showModelSelector() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		models, err := h.chatState.Provider.ListModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to get available models: %w", err)
		}

		currentModel := h.chatState.Config.Model
		return ModelSelectorMsg{
			Models:       models,
			CurrentModel: currentModel,
		}
	}
}

// showProviderSelector handles /provider command
func (h *Handler) showProviderSelector() tea.Cmd {
	return func() tea.Msg {
		providers := []string{"anthropic", "ollama"}
		return ProviderSelectorMsg{
			Providers: providers,
		}
	}
}

// showStatus handles /status command
func (h *Handler) showStatus() tea.Cmd {
	return func() tea.Msg {
		status := fmt.Sprintf("Current Configuration:\n\n"+
			"Provider: %s\n"+
			"Model: %s\n"+
			"Ollama URL: %s\n"+
			"Chat History: %d exchanges\n"+
			"Input History: %d commands",
			h.chatState.Config.Provider,
			h.chatState.Config.Model,
			h.chatState.Config.OllamaBaseURL,
			len(h.chatState.History),
			len(h.chatState.InputHistory))

		exchange := state.Exchange{
			Prompt:   "/status",
			Response: status,
		}

		return AIResponse{
			Content: exchange.Response,
			Err:     nil,
		}
	}
}

// showHelp handles /help command
func (h *Handler) showHelp() tea.Cmd {
	return func() tea.Msg {
		var help strings.Builder
		help.WriteString("Available commands:\n\n")
		for _, cmd := range AvailableCommands() {
			help.WriteString(fmt.Sprintf("  %s - %s\n", cmd.Name, cmd.Description))
		}
		help.WriteString("\nKeyboard shortcuts:\n")
		help.WriteString("  Tab       - Complete command\n")
		help.WriteString("  ↑/↓       - Navigate history\n")
		help.WriteString("  Alt+Enter - New line\n")
		help.WriteString("  Ctrl+C    - Cancel current operation\n")
		help.WriteString("  ESC       - Exit selection mode\n")

		exchange := state.Exchange{
			Prompt:   "/help",
			Response: help.String(),
		}

		return AIResponse{
			Content: exchange.Response,
			Err:     nil,
		}
	}
}

// clearHistory handles /clearhistory command
func (h *Handler) clearHistory() tea.Cmd {
	return func() tea.Msg {
		if h.chatState.HistoryManager != nil {
			h.chatState.HistoryManager.Clear()
			_ = h.chatState.HistoryManager.Save()
			h.chatState.InputHistory = []string{}
		}
		h.chatState.HistoryIndex = -1

		exchange := state.Exchange{
			Prompt:   "/clearhistory",
			Response: "✅ Command history cleared",
		}

		return AIResponse{
			Content: exchange.Response,
			Err:     nil,
		}
	}
}

// requestResponse handles normal chat messages
func (h *Handler) requestResponse() tea.Cmd {
	return func() tea.Msg {
		h.chatState.SetThinking(true)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		response, err := h.chatState.Provider.Generate(ctx, h.chatState.CurrentPrompt)
		if err != nil {
			return AIResponse{
				Content: "",
				Err:     err,
			}
		}

		return AIResponse{
			Content: response,
			Err:     nil,
		}
	}
}

// Message types for command responses
