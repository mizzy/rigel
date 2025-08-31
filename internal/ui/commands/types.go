package commands

import "github.com/mizzy/rigel/internal/llm"

// AIResponse represents a response from the AI
type AIResponse struct {
	Content string
	Err     error
}

// ModelSelectorMsg is sent when model selection is requested
type ModelSelectorMsg struct {
	Models       []llm.Model
	CurrentModel string
}

// ProviderSelectorMsg is sent when provider selection is requested
type ProviderSelectorMsg struct {
	Providers []string
}

// ProviderSwitchResponse represents a provider switch response
type ProviderSwitchResponse struct {
	Provider     llm.Provider
	ProviderName string
}
