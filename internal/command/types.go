package command

import (
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

// ModelSelectorMsg represents a model selection request
type ModelSelectorMsg struct {
	CurrentModel string
	Models       []llm.Model
	Error        error
}

// ProviderSelectorMsg represents a provider selection request
type ProviderSelectorMsg struct {
	CurrentProvider llm.Provider
	Providers       []llm.Provider
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
