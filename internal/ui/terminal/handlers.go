package terminal

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/llm"
)

// requestResponse sends a request to the LLM provider with conversation history
func (m *Model) requestResponse(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Build message history from chat exchanges
		history := m.chatState.GetHistory()
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
		provider := m.llmState.GetCurrentProvider()
		if provider == nil {
			return aiResponse{err: fmt.Errorf("no provider available")}
		}
		response, err := provider.GenerateWithHistory(ctx, messages, llm.GenerateOptions{})
		if err != nil {
			return aiResponse{err: err}
		}
		return aiResponse{content: strings.TrimSpace(response)}
	}
}
