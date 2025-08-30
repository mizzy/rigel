package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/tools"
)

type Agent struct {
	provider llm.Provider
	tools    []tools.Tool
	memory   *Memory
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
		tools: []tools.Tool{},
	}
}

func (a *Agent) RegisterTool(tool tools.Tool) {
	a.tools = append(a.tools, tool)
}

func (a *Agent) Execute(ctx context.Context, task string) (string, error) {
	systemPrompt := a.buildSystemPrompt()

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		Temperature:  0.7,
	}

	userPrompt := a.buildUserPrompt(task)

	response, err := a.provider.GenerateWithOptions(ctx, userPrompt, opts)
	if err != nil {
		return "", fmt.Errorf("failed to execute task: %w", err)
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
