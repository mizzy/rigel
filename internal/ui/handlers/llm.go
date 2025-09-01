package handlers

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
