package termflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/mizzy/rigel/internal/agent"
	"github.com/mizzy/rigel/internal/command"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/git"
	"github.com/mizzy/rigel/internal/history"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/state"
	"github.com/mizzy/rigel/internal/tools"
	"github.com/mizzy/rigel/lib/termflow"
)

// ChatSession represents a termflow-based chat session
type ChatSession struct {
	client         *termflow.InteractiveClient
	historyManager *history.Manager
	chatState      *state.ChatState
	llmState       *state.LLMState
	agent          *agent.Agent
	config         *config.Config
	gitInfo        *git.Info
	ctrlCPressed   bool // Track Ctrl+C presses for 2-press exit
	waitingForExit bool // Waiting for second Ctrl+C, don't show prompt
}

// NewChatSession creates a new termflow chat session
func NewChatSession(provider llm.Provider, cfg *config.Config) (*ChatSession, error) {
	client, err := termflow.NewInteractive()
	if err != nil {
		return nil, fmt.Errorf("failed to create termflow client: %w", err)
	}

	// Initialize history manager
	histManager, err := history.NewManager()
	if err != nil {
		histManager = nil // Continue without history if it fails
	} else {
		_ = histManager.Load() // Load existing history
	}

	// Initialize LLM state
	llmState := state.NewLLMState()
	llmState.SetCurrentProvider(provider)

	// Create intelligent agent with file tools
	intelligentAgent := agent.New(provider)
	fileTool := tools.NewFileTool()
	intelligentAgent.RegisterTool(fileTool)

	// Use UI progress display for termflow mode
	uiProgress := agent.NewUIProgressDisplay()
	intelligentAgent.SetProgressDisplay(uiProgress)

	session := &ChatSession{
		client:         client,
		historyManager: histManager,
		chatState:      state.NewChatState(),
		llmState:       llmState,
		agent:          intelligentAgent,
		config:         cfg,
		gitInfo:        git.GetRepoInfo(),
	}

	// Set up command completion
	session.setupCompletion()

	// Load persistent history into the client
	if histManager != nil {
		history := histManager.GetCommands()
		client.SetHistory(history)
	}

	return session, nil
}

// setupCompletion sets up tab completion for commands
func (cs *ChatSession) setupCompletion() {
	provider := termflow.NewCompletionProvider()

	// Add available commands
	for _, cmd := range command.AvailableCommands {
		provider.AddCommand(cmd.Command, cmd.Description)
	}

	cs.client.SetCompletionFunc(func(input string) []string {
		return provider.GetCompletionStrings(input)
	})
}

// Run starts the chat session
func (cs *ChatSession) Run() error {
	defer cs.client.Close()
	defer func() {
		if cs.historyManager != nil {
			cs.historyManager.Save()
		}
	}()

	// Show welcome message
	cs.showWelcome()

	// Main chat loop
	for {
		var input string
		var err error

		// If waiting for exit, use a different input method that doesn't show prompt
		if cs.waitingForExit {
			// Reset the waiting flag and use raw keyboard input
			cs.waitingForExit = false
			input, err = cs.readRawInput()
		} else {
			// Use ReadLineOrMultiLine to support both single and multi-line input
			input, err = cs.client.ReadLineOrMultiLine()
		}

		if err != nil {
			// Handle interruption with 2-press exit behavior
			if err.Error() == "interrupted" {
				if cs.ctrlCPressed {
					// Second Ctrl+C - exit
					cs.client.ShowInfo("Goodbye!")
					break
				}
				// First Ctrl+C - show message and set flag
				cs.ctrlCPressed = true
				cs.waitingForExit = true
				cs.client.ShowInfo("Press Ctrl+C again to exit")
				// Continue the loop but don't show prompt - wait for second Ctrl+C
				continue
			}
			cs.client.ShowError(err)
			break
		}

		// Reset Ctrl+C flag when user provides input
		cs.ctrlCPressed = false

		// Handle empty input
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Add to history
		if cs.historyManager != nil {
			cs.historyManager.Add(input)
		}

		// Handle quit commands
		if input == "/quit" || input == "/exit" {
			cs.client.ShowInfo("Goodbye!")
			break
		}

		// Process the input
		if err := cs.processInput(input); err != nil {
			cs.client.ShowError(err)
		}
	}

	return nil
}

// showWelcome displays the welcome message
func (cs *ChatSession) showWelcome() {
	cs.client.Printf("\n\033[1;38;5;87m✦\033[0m \033[1mRigel - AI Coding Agent\033[0m\n")
	if cs.gitInfo != nil {
		cs.client.Printf("  \033[38;2;87;147;255m%s\033[0m \033[38;5;117m(%s)\033[0m\n", cs.gitInfo.RepoName, cs.gitInfo.Branch)
	}
	cs.client.Printf("  Using termflow UI - terminal scrollback is preserved!\n")
	cs.client.Printf("  \033[90mInput:\033[0m Single line or end with '...' for multi-line\n")
	cs.client.Printf("  \033[90mCommands:\033[0m Type / for commands (Ctrl+C to exit)\n\n")
}

// processInput processes user input (commands or chat messages)
func (cs *ChatSession) processInput(input string) error {
	cs.chatState.SetCurrentPrompt(input)

	// Reset Ctrl+C flag when processing input
	cs.ctrlCPressed = false

	if strings.HasPrefix(input, "/") {
		// Handle commands
		return cs.handleCommand(input)
	} else {
		// Handle chat message
		return cs.handleChatMessage(input)
	}
}

// handleCommand processes slash commands
func (cs *ChatSession) handleCommand(input string) error {
	result := command.HandleCommand(
		input,
		cs.llmState,
		cs.chatState,
		cs.config,
		cs.historyManager,
		cs.getInputHistory(),
	)

	// Handle different result types
	switch result.Type {
	case "response":
		if result.Error != nil {
			return result.Error
		}
		cs.client.PrintResponse(result.Content)
		cs.chatState.AddExchange(input, result.Content)

	case "async":
		if result.AsyncFn != nil {
			// Show animated processing spinner
			spinner := cs.client.ShowThinkingWithSpinner("Processing...")
			asyncResult := result.AsyncFn()
			spinner.Stop()
			if asyncResult.Error != nil {
				return asyncResult.Error
			}
			cs.client.PrintResponse(asyncResult.Content)
			cs.chatState.AddExchange(input, asyncResult.Content)
		}

	case "clear":
		// Clear chat state
		cs.chatState.ClearHistory()
		cs.client.ShowInfo("Chat history cleared")

	case "clear_input_history":
		if cs.historyManager != nil {
			if err := cs.historyManager.Clear(); err != nil {
				return fmt.Errorf("failed to clear history: %w", err)
			}
		}
		cs.client.ShowInfo("Command history cleared")

	case "status":
		if result.StatusInfo != nil {
			content := cs.formatStatusInfo(result.StatusInfo)
			cs.client.PrintResponse(content)
			cs.chatState.AddExchange(input, content)
		}

	case "request":
		// Handle normal prompts using intelligent agent
		return cs.handleChatMessage(result.Prompt)

	default:
		if result.Content != "" {
			cs.client.PrintResponse(result.Content)
			cs.chatState.AddExchange(input, result.Content)
		}
	}

	cs.chatState.ClearCurrentPrompt()
	return nil
}

// handleChatMessage processes regular chat messages
func (cs *ChatSession) handleChatMessage(input string) error {
	// Show animated thinking spinner
	spinner := cs.client.ShowThinkingWithSpinner("Thinking...")
	defer spinner.Stop()

	// Use the intelligent agent to generate response
	response, err := cs.agent.Execute(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	// Display only the AI response (user input is already visible)
	cs.client.PrintResponse(response)

	// Add to chat state
	cs.chatState.AddExchange(input, response)
	cs.chatState.ClearCurrentPrompt()

	return nil
}

// formatStatusInfo formats status information for display
func (cs *ChatSession) formatStatusInfo(status *command.StatusInfo) string {
	return fmt.Sprintf("✦ Rigel Session Status\n\n"+
		"🤖 LLM Configuration\n"+
		"  Provider: %s\n"+
		"  Model: %s\n\n"+
		"💬 Chat History\n"+
		"  Messages: %d\n"+
		"  User tokens: ~%d\n"+
		"  Assistant tokens: ~%d\n"+
		"  Total tokens: ~%d\n\n"+
		"📝 Command History\n"+
		"  Commands saved: %d\n"+
		"  Persistence: %s\n\n"+
		"🔧 Environment\n"+
		"  UI Mode: termflow\n"+
		"  Log level: %s\n"+
		"  Repository context: %s\n",
		status.Provider, status.Model,
		status.MessageCount,
		status.UserTokens, status.AssistantTokens, status.TotalTokens,
		status.CommandsCount,
		map[bool]string{true: "✓ Enabled", false: "✗ Disabled"}[status.PersistenceEnabled],
		status.LogLevel,
		map[bool]string{true: "✓ AGENTS.md loaded", false: "✗ Not initialized (run /init)"}[status.RepositoryInitialized])
}

// readRawInput reads input without showing a prompt (for Ctrl+C waiting state)
func (cs *ChatSession) readRawInput() (string, error) {
	// Use the line editor directly but without showing prompt initially
	lineEditor, err := termflow.NewLineEditor(cs.client.Client)
	if err != nil {
		// Fall back to basic input reading
		return cs.client.Client.ReadLine()
	}

	// Set history
	lineEditor.SetHistory(cs.getInputHistory())

	// Read without initial prompt display
	return lineEditor.ReadLineWithoutPrompt()
}

// getInputHistory returns the current input history for commands
func (cs *ChatSession) getInputHistory() []string {
	if cs.historyManager != nil {
		return cs.historyManager.GetCommands()
	}
	return []string{}
}
