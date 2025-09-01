package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/tools"
)

type Agent struct {
	provider        llm.Provider
	tools           []tools.Tool
	memory          *Memory
	promptAnalyzer  *PromptAnalyzer
	autoToolEnabled bool
	progressDisplay ProgressDisplay
}

type Memory struct {
	conversationHistory []Message
	context             map[string]interface{}
}

type Message struct {
	Role    string
	Content string
}

func New(provider llm.Provider) *Agent {
	return &Agent{
		provider: provider,
		memory: &Memory{
			conversationHistory: []Message{},
			context:             make(map[string]interface{}),
		},
		tools:           []tools.Tool{},
		promptAnalyzer:  NewPromptAnalyzer(provider),
		autoToolEnabled: true,
		progressDisplay: &ConsoleProgressDisplay{},
	}
}

func (a *Agent) RegisterTool(tool tools.Tool) {
	a.tools = append(a.tools, tool)
}

func (a *Agent) Execute(ctx context.Context, task string) (string, error) {
	var toolResults []ToolExecutionResult
	var finalResponse strings.Builder

	// Phase 1: Analyze prompt for file operations with conversation history
	if a.autoToolEnabled {
		matches := a.promptAnalyzer.AnalyzePromptWithHistory(task, a.memory.conversationHistory)
		if len(matches) > 0 {
			// Phase 2: Create structured tasks from intents
			tasks := CreateTasksFromMatches(matches)

			finalResponse.WriteString("I'll help you with that. Here's what I need to do:\n\n")

			// Show task list to user
			for i, taskItem := range tasks {
				finalResponse.WriteString(fmt.Sprintf("%d. %s\n", i+1, taskItem.Description))
			}
			finalResponse.WriteString("\nExecuting tasks...\n\n")

			// Phase 3: Process tasks and generate content if needed
			for i, taskItem := range tasks {
				if taskItem.Match.Intent == IntentWrite && taskItem.Match.Content == "<GENERATE_TEXT>" {
					// Generate content using LLM with conversation context
					contentPrompt := a.buildContentGenerationPrompt(task, a.memory.conversationHistory)
					generatedContent, err := a.provider.Generate(ctx, contentPrompt)
					if err == nil {
						tasks[i].Match.Content = strings.TrimSpace(generatedContent)
					} else {
						tasks[i].Match.Content = "Sample text generated for user request."
					}
				}
			}

			// Phase 4: Execute tasks with progress tracking
			if uiDisplay, ok := a.progressDisplay.(*UIProgressDisplay); ok {
				// Execute with progress tracking
				toolResults = a.ExecuteTasksWithProgress(ctx, tasks, a.progressDisplay)

				// Add progress messages
				progressMessages := uiDisplay.GetAllMessages()
				if len(progressMessages) > 0 {
					finalResponse.WriteString(strings.Join(progressMessages, "\n"))
					finalResponse.WriteString("\n\n")
				}
			} else {
				// Execute without UI progress display
				toolResults = a.ExecuteTasksWithProgress(ctx, tasks, a.progressDisplay)
			}

			// Build response with tool results (detailed output)
			finalResponse.WriteString("Results:\n")
			for _, result := range toolResults {
				if result.Error == nil && result.Output != "" {
					finalResponse.WriteString(fmt.Sprintf("📄 %s output:\n%s\n\n", result.Tool, result.Output))
				}
			}
		}
	}

	// Generate AI response
	systemPrompt := a.buildSystemPrompt()
	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		Temperature:  0.7,
	}

	// Include tool results in the prompt if available
	var userPrompt string
	if len(toolResults) > 0 {
		toolContext := a.buildToolContext(toolResults)
		userPrompt = fmt.Sprintf("%s\n\nTool execution results:\n%s", a.buildUserPrompt(task), toolContext)
	} else {
		userPrompt = a.buildUserPrompt(task)
	}

	response, err := a.provider.GenerateWithOptions(ctx, userPrompt, opts)
	if err != nil {
		return "", fmt.Errorf("failed to execute task: %w", err)
	}

	// Combine tool results and AI response
	if finalResponse.Len() > 0 {
		finalResponse.WriteString("---\n\n")
		finalResponse.WriteString(response)
		response = finalResponse.String()
	}

	a.memory.conversationHistory = append(a.memory.conversationHistory,
		Message{Role: "user", Content: task},
		Message{Role: "assistant", Content: response},
	)

	return response, nil
}

func (a *Agent) buildSystemPrompt() string {
	prompts := []string{
		"You are Rigel, an intelligent AI coding assistant.",
		"You help developers write clean, efficient, and maintainable code.",
		"You can analyze code, suggest improvements, and generate new code based on requirements.",
		"Always follow best practices and coding standards.",
		"Be concise but thorough in your explanations.",
	}

	if len(a.tools) > 0 {
		prompts = append(prompts, "\nAvailable tools:")
		for _, tool := range a.tools {
			prompts = append(prompts, fmt.Sprintf("- %s: %s", tool.Name(), tool.Description()))
		}
	}

	return strings.Join(prompts, "\n")
}

func (a *Agent) buildUserPrompt(task string) string {
	if len(a.memory.conversationHistory) > 0 {
		var history []string
		for _, msg := range a.memory.conversationHistory {
			history = append(history, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
		}
		return fmt.Sprintf("Previous conversation:\n%s\n\nCurrent task: %s",
			strings.Join(history, "\n"), task)
	}
	return task
}

func (a *Agent) ClearMemory() {
	a.memory.conversationHistory = []Message{}
	a.memory.context = make(map[string]interface{})
}

func (a *Agent) SetContext(key string, value interface{}) {
	a.memory.context[key] = value
}

func (a *Agent) GetContext(key string) (interface{}, bool) {
	val, ok := a.memory.context[key]
	return val, ok
}

// SetAutoToolEnabled enables or disables automatic tool execution
func (a *Agent) SetAutoToolEnabled(enabled bool) {
	a.autoToolEnabled = enabled
}

// IsAutoToolEnabled returns whether automatic tool execution is enabled
func (a *Agent) IsAutoToolEnabled() bool {
	return a.autoToolEnabled
}

// SetProgressDisplay sets a custom progress display implementation
func (a *Agent) SetProgressDisplay(display ProgressDisplay) {
	a.progressDisplay = display
}

// GetProgressDisplay returns the current progress display implementation
func (a *Agent) GetProgressDisplay() ProgressDisplay {
	return a.progressDisplay
}

// buildToolContext builds context from tool execution results
func (a *Agent) buildToolContext(results []ToolExecutionResult) string {
	var context []string
	for _, result := range results {
		if result.Error != nil {
			context = append(context, fmt.Sprintf("Tool %s failed: %v", result.Tool, result.Error))
		} else {
			context = append(context, fmt.Sprintf("Tool %s succeeded: %s", result.Tool, result.Output))
		}
	}
	return strings.Join(context, "\n")
}

// buildContentGenerationPrompt creates a context-aware prompt for content generation
func (a *Agent) buildContentGenerationPrompt(task string, history []Message) string {
	var contextBuilder strings.Builder

	contextBuilder.WriteString("Generate appropriate, realistic content for a file based on this user request.\n")
	contextBuilder.WriteString("The content should be meaningful and ready to use, not placeholder text.\n\n")

	// Add conversation context if available
	if len(history) > 0 {
		contextBuilder.WriteString("Conversation context:\n")
		for _, msg := range history {
			contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		contextBuilder.WriteString("\n")
	}

	contextBuilder.WriteString(fmt.Sprintf("Current request: %s\n\n", task))
	contextBuilder.WriteString("Examples:\n")
	contextBuilder.WriteString("- If creating config.json, generate valid JSON configuration\n")
	contextBuilder.WriteString("- If creating README.md, generate actual documentation\n")
	contextBuilder.WriteString("- If creating sample text, generate meaningful content\n\n")
	contextBuilder.WriteString("Provide ONLY the file content to be written, no explanations, no questions, no markdown code blocks.")

	return contextBuilder.String()
}
