package terminal

import "github.com/mizzy/rigel/internal/llm"

// modelSelectorMsg is sent when model selection is requested
type modelSelectorMsg struct {
	models []llm.Model
	err    error
}

// providerSelectorMsg is sent when provider selection is requested
type providerSelectorMsg struct {
	providers       []llm.Provider
	currentProvider llm.Provider
	err             error
}
