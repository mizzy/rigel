package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/analyzer"
	"github.com/mizzy/rigel/internal/llm"
)

// Available commands
var availableCommands = []struct {
	command     string
	description string
}{
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

// handleCommand processes commands and returns the appropriate tea.Cmd
func (m *ChatModel) handleCommand(trimmedPrompt string) tea.Cmd {
	switch trimmedPrompt {
	case "/init":
		return m.analyzeRepository()

	case "/model":
		return m.showModelSelector()

	case "/provider":
		return m.showProviderSelector()

	case "/status":
		return m.showStatus()

	case "/help":
		return m.showHelp()

	case "/clear":
		m.history = []Exchange{}
		m.thinking = false
		return nil

	case "/clearhistory":
		return m.clearHistory()

	case "/exit", "/quit":
		m.quitting = true
		return tea.Quit

	default:
		if strings.HasPrefix(trimmedPrompt, "/") {
			m.err = fmt.Errorf("unknown command: %s, type /help for available commands", trimmedPrompt)
			m.thinking = false
			return nil
		}
		return m.requestResponse(m.currentPrompt)
	}
}

// showHelp displays the help message
func (m *ChatModel) showHelp() tea.Cmd {
	return func() tea.Msg {
		var help strings.Builder
		help.WriteString("Available commands:\n\n")
		for _, cmd := range availableCommands {
			help.WriteString(fmt.Sprintf("  %s - %s\n", cmd.command, cmd.description))
		}
		help.WriteString("\nKeyboard shortcuts:\n")
		help.WriteString("  Tab       - Complete command\n")
		help.WriteString("  ↑/↓       - Navigate suggestions\n")
		help.WriteString("  Enter     - Send message or select suggestion\n")
		help.WriteString("  Alt+Enter - New line\n")
		help.WriteString("  Ctrl+C    - Exit\n")

		return aiResponse{
			content: help.String(),
		}
	}
}

// analyzeRepository runs the repository analysis
func (m *ChatModel) analyzeRepository() tea.Cmd {
	return func() tea.Msg {
		// Analyze the repository and generate AGENTS.md
		analyzer := analyzer.NewRepoAnalyzer(m.provider)
		content, err := analyzer.Analyze()
		if err != nil {
			return aiResponse{err: err}
		}

		// Write the AGENTS.md file
		err = analyzer.WriteAgentsFile(content)
		if err != nil {
			return aiResponse{err: err}
		}

		return aiResponse{
			content: "Repository analyzed successfully! AGENTS.md has been created.\n\n" +
				"The file contains:\n" +
				"• Repository structure and overview\n" +
				"• Key components and their responsibilities\n" +
				"• File purposes and dependencies\n" +
				"• Testing and configuration information",
		}
	}
}

type modelSelectorMsg struct {
	currentModel string
	models       []llm.Model
	err          error
}

func (m *ChatModel) showModelSelector() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		currentModel := m.provider.GetCurrentModel()
		models, err := m.provider.ListModels(ctx)
		if err != nil {
			return modelSelectorMsg{
				currentModel: currentModel,
				err:          err,
			}
		}

		return modelSelectorMsg{
			currentModel: currentModel,
			models:       models,
		}
	}
}

type providerSelectorMsg struct {
	currentProvider string
	providers       []string
	err             error
}

func (m *ChatModel) showProviderSelector() tea.Cmd {
	return func() tea.Msg {
		// Get available providers
		providers := []string{"anthropic", "ollama"}

		// Get current provider from config
		currentProvider := ""
		if m.config != nil {
			currentProvider = m.config.Provider
		}

		return providerSelectorMsg{
			currentProvider: currentProvider,
			providers:       providers,
		}
	}
}

// showStatus displays the current session status and configuration
func (m *ChatModel) showStatus() tea.Cmd {
	return func() tea.Msg {
		var status strings.Builder
		status.WriteString("📊 Current Session Status\n")
		status.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		// Provider and Model Information
		status.WriteString("🤖 LLM Configuration:\n")
		if m.config != nil {
			status.WriteString(fmt.Sprintf("  Provider: %s\n", m.config.Provider))
		}
		if m.provider != nil {
			status.WriteString(fmt.Sprintf("  Model: %s\n", m.provider.GetCurrentModel()))
		}
		status.WriteString("\n")

		// Chat History Status
		status.WriteString("💬 Chat History:\n")
		status.WriteString(fmt.Sprintf("  Messages: %d\n", len(m.history)))

		// Calculate token usage (approximate)
		totalUserChars := 0
		totalAssistantChars := 0
		for _, exchange := range m.history {
			totalUserChars += len(exchange.Prompt)
			totalAssistantChars += len(exchange.Response)
		}
		approxUserTokens := totalUserChars / 4 // Rough approximation: 1 token ≈ 4 chars
		approxAssistantTokens := totalAssistantChars / 4

		status.WriteString(fmt.Sprintf("  Approximate tokens (user): ~%d\n", approxUserTokens))
		status.WriteString(fmt.Sprintf("  Approximate tokens (assistant): ~%d\n", approxAssistantTokens))
		status.WriteString(fmt.Sprintf("  Total approximate tokens: ~%d\n", approxUserTokens+approxAssistantTokens))
		status.WriteString("\n")

		// Command History Status
		status.WriteString("📝 Command History:\n")
		status.WriteString(fmt.Sprintf("  Commands in history: %d\n", len(m.inputHistory)))
		if m.historyManager != nil {
			status.WriteString("  Persistent history: Enabled\n")
		} else {
			status.WriteString("  Persistent history: Disabled\n")
		}
		status.WriteString("\n")

		// Environment Information
		status.WriteString("🔧 Environment:\n")
		if m.config != nil && m.config.LogLevel != "" {
			status.WriteString(fmt.Sprintf("  Log level: %s\n", m.config.LogLevel))
		}

		// Check for AGENTS.md
		if _, err := os.Stat("AGENTS.md"); err == nil {
			status.WriteString("  Repository context: AGENTS.md loaded\n")
		} else {
			status.WriteString("  Repository context: Not initialized (run /init)\n")
		}

		return aiResponse{
			content: status.String(),
		}
	}
}

// clearHistory clears the command history
func (m *ChatModel) clearHistory() tea.Cmd {
	return func() tea.Msg {
		// Clear in-memory history
		m.inputHistory = []string{}
		m.historyIndex = -1
		m.currentInput = ""

		// Clear persistent history
		if m.historyManager != nil {
			if err := m.historyManager.Clear(); err != nil {
				return aiResponse{
					err: fmt.Errorf("failed to clear history: %w", err),
				}
			}
		}

		return aiResponse{
			content: "Command history cleared successfully.",
		}
	}
}
