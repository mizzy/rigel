package command

import (
	"context"

	"github.com/mizzy/rigel/internal/llm"
)

// Result represents the result of command execution
type Result struct {
	Type    string // "response", "model_selector", "provider_selector", "status", "quit", "clear", "request"
	Error   error
	Content string // Text response content
	Prompt  string // For "request" type - the prompt to send to LLM

	// Type-specific data (only one should be set based on Type)
	ModelSelector    *ModelSelectorMsg
	ProviderSelector *ProviderSelectorMsg
	StatusInfo       *StatusInfo
}

// CommandContext interface defines what the command handler needs from the UI
type CommandContext interface {
	// State accessors for command implementations
	GetProviderName() string
	GetCurrentModel() string
	ListModels(context.Context) ([]llm.Model, error)
	GetLogLevel() string
	GetStatusInfo() StatusInfo
	GetInputHistory() []string

	// State mutators for command implementations
	ClearChatHistory()
	ClearInputHistory()
	ClearPersistentHistory() error
}

// ModelSelectorMsg represents a model selection request
type ModelSelectorMsg struct {
	CurrentModel string
	Models       []llm.Model
	Error        error
}

// ProviderSelectorMsg represents a provider selection request
type ProviderSelectorMsg struct {
	CurrentProvider string
	Providers       []string
}

// StatusInfo represents session status information
type StatusInfo struct {
	Provider              string
	Model                 string
	MessageCount          int
	UserTokens            int
	AssistantTokens       int
	TotalTokens           int
	CommandsCount         int
	PersistenceEnabled    bool
	LogLevel              string
	RepositoryInitialized bool
}
