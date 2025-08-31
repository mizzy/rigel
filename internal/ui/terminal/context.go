package terminal

import (
	"context"
	"fmt"
	"os"

	"github.com/mizzy/rigel/internal/command"
	"github.com/mizzy/rigel/internal/llm"
)

// CommandContext interface implementation for Model
// This allows the CommandHandler to interact with the Model

func (m *Model) GetProviderName() string {
	if m.config != nil {
		return m.config.Provider
	}
	return ""
}

func (m *Model) GetCurrentModel() string {
	if m.provider != nil {
		return m.provider.GetCurrentModel()
	}
	return ""
}

func (m *Model) ListModels(ctx context.Context) ([]llm.Model, error) {
	if m.provider == nil {
		return nil, fmt.Errorf("no provider available")
	}

	return m.provider.ListModels(ctx)
}

func (m *Model) GetLogLevel() string {
	if m.config != nil {
		return m.config.LogLevel
	}
	return ""
}

func (m *Model) GetInputHistory() []string {
	return m.inputHistory
}

func (m *Model) ClearPersistentHistory() error {
	if m.historyManager != nil {
		return m.historyManager.Clear()
	}
	return nil
}

func (m *Model) GetStatusInfo() command.StatusInfo {
	provider := m.GetProviderName()
	model := m.GetCurrentModel()
	logLevel := m.GetLogLevel()

	// Calculate token usage
	totalUserChars := 0
	totalAssistantChars := 0
	for _, exchange := range m.chatState.GetHistory() {
		totalUserChars += len(exchange.Prompt)
		totalAssistantChars += len(exchange.Response)
	}
	approxUserTokens := totalUserChars / 4
	approxAssistantTokens := totalAssistantChars / 4
	totalTokens := approxUserTokens + approxAssistantTokens

	// Check if AGENTS.md exists
	repositoryInitialized := false
	if _, err := os.Stat("AGENTS.md"); err == nil {
		repositoryInitialized = true
	}

	return command.StatusInfo{
		Provider:              provider,
		Model:                 model,
		MessageCount:          m.chatState.GetMessageCount(),
		UserTokens:            approxUserTokens,
		AssistantTokens:       approxAssistantTokens,
		TotalTokens:           totalTokens,
		CommandsCount:         len(m.inputHistory),
		PersistenceEnabled:    m.historyManager != nil,
		LogLevel:              logLevel,
		RepositoryInitialized: repositoryInitialized,
	}
}

func (m *Model) ClearChatHistory() {
	m.chatState.ClearHistory()
}

func (m *Model) ClearInputHistory() {
	m.inputHistory = []string{}
	m.historyIndex = -1
	m.currentInput = ""
}
