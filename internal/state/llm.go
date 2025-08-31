package state

import (
	"strings"

	"github.com/mizzy/rigel/internal/llm"
)

// LLMState manages LLM configuration and selection
type LLMState struct {
	// Current LLM configuration
	currentProvider string
	currentModel    string
	provider        llm.Provider

	// Model selection
	modelSelectionActive bool
	availableModels      []llm.Model
	filteredModels       []llm.Model
	selectedModelIndex   int
	modelFilter          string

	// Provider selection
	providerSelectionActive bool
	availableProviders      []string
	selectedProviderIndex   int
}

// NewLLMState creates a new LLM state manager
func NewLLMState() *LLMState {
	return &LLMState{}
}

// Model Selection Methods

// Current Configuration Methods

// SetCurrentProvider updates the current provider
func (ls *LLMState) SetCurrentProvider(providerName string, provider llm.Provider) {
	ls.currentProvider = providerName
	ls.provider = provider
	if provider != nil {
		ls.currentModel = provider.GetCurrentModel()
	}
}

// GetCurrentProvider returns the current provider name
func (ls *LLMState) GetCurrentProvider() string {
	return ls.currentProvider
}

// GetCurrentModel returns the current model name
func (ls *LLMState) GetCurrentModel() string {
	return ls.currentModel
}

// GetProvider returns the current LLM provider
func (ls *LLMState) GetProvider() llm.Provider {
	return ls.provider
}

// SetCurrentModel updates the current model name
func (ls *LLMState) SetCurrentModel(modelName string) {
	ls.currentModel = modelName
}

// Model Selection Methods

// IsModelSelectionActive returns whether model selection mode is active
func (ls *LLMState) IsModelSelectionActive() bool {
	return ls.modelSelectionActive
}

// ActivateModelSelection enters model selection mode
func (ls *LLMState) ActivateModelSelection(models []llm.Model) {
	ls.modelSelectionActive = true
	ls.availableModels = models
	ls.filteredModels = models
	ls.selectedModelIndex = 0
	ls.modelFilter = ""
}

// DeactivateModelSelection exits model selection mode
func (ls *LLMState) DeactivateModelSelection() {
	ls.modelSelectionActive = false
	ls.availableModels = nil
	ls.filteredModels = nil
	ls.selectedModelIndex = 0
	ls.modelFilter = ""
}

// GetFilteredModels returns currently filtered models
func (ls *LLMState) GetFilteredModels() []llm.Model {
	return ls.filteredModels
}

// GetSelectedModelIndex returns the currently selected model index
func (ls *LLMState) GetSelectedModelIndex() int {
	return ls.selectedModelIndex
}

// GetSelectedModel returns the currently selected model
func (ls *LLMState) GetSelectedModel() (llm.Model, bool) {
	if !ls.modelSelectionActive || len(ls.filteredModels) == 0 || ls.selectedModelIndex >= len(ls.filteredModels) {
		return llm.Model{}, false
	}
	return ls.filteredModels[ls.selectedModelIndex], true
}

// GetModelFilter returns the current model filter
func (ls *LLMState) GetModelFilter() string {
	return ls.modelFilter
}

// SetModelFilter updates the model filter and re-filters
func (ls *LLMState) SetModelFilter(filter string) {
	ls.modelFilter = filter
	ls.filterModels()
	ls.selectedModelIndex = 0 // Reset selection when filter changes
}

// MoveModelSelectionUp moves model selection up
func (ls *LLMState) MoveModelSelectionUp() {
	if ls.selectedModelIndex > 0 {
		ls.selectedModelIndex--
	}
}

// MoveModelSelectionDown moves model selection down
func (ls *LLMState) MoveModelSelectionDown() {
	if ls.selectedModelIndex < len(ls.filteredModels)-1 {
		ls.selectedModelIndex++
	}
}

// HasFilteredModels returns whether there are any filtered models
func (ls *LLMState) HasFilteredModels() bool {
	return len(ls.filteredModels) > 0
}

// filterModels applies the current filter to available models
func (ls *LLMState) filterModels() {
	if ls.modelFilter == "" {
		ls.filteredModels = ls.availableModels
		return
	}

	filter := strings.ToLower(ls.modelFilter)
	ls.filteredModels = ls.filteredModels[:0] // Clear but keep capacity

	for _, model := range ls.availableModels {
		if strings.Contains(strings.ToLower(model.Name), filter) {
			ls.filteredModels = append(ls.filteredModels, model)
		}
	}
}

// Provider Selection Methods

// IsProviderSelectionActive returns whether provider selection mode is active
func (ls *LLMState) IsProviderSelectionActive() bool {
	return ls.providerSelectionActive
}

// ActivateProviderSelection enters provider selection mode
func (ls *LLMState) ActivateProviderSelection(providers []string, currentProvider string) {
	ls.providerSelectionActive = true
	ls.availableProviders = providers
	ls.selectedProviderIndex = 0

	// Find current provider index
	for i, p := range providers {
		if p == currentProvider {
			ls.selectedProviderIndex = i
			break
		}
	}
}

// DeactivateProviderSelection exits provider selection mode
func (ls *LLMState) DeactivateProviderSelection() {
	ls.providerSelectionActive = false
	ls.availableProviders = nil
	ls.selectedProviderIndex = 0
}

// GetAvailableProviders returns available providers
func (ls *LLMState) GetAvailableProviders() []string {
	return ls.availableProviders
}

// GetSelectedProviderIndex returns the currently selected provider index
func (ls *LLMState) GetSelectedProviderIndex() int {
	return ls.selectedProviderIndex
}

// GetSelectedProvider returns the currently selected provider
func (ls *LLMState) GetSelectedProvider() (string, bool) {
	if !ls.providerSelectionActive || len(ls.availableProviders) == 0 || ls.selectedProviderIndex >= len(ls.availableProviders) {
		return "", false
	}
	return ls.availableProviders[ls.selectedProviderIndex], true
}

// MoveProviderSelectionUp moves provider selection up
func (ls *LLMState) MoveProviderSelectionUp() {
	if ls.selectedProviderIndex > 0 {
		ls.selectedProviderIndex--
	}
}

// MoveProviderSelectionDown moves provider selection down
func (ls *LLMState) MoveProviderSelectionDown() {
	if ls.selectedProviderIndex < len(ls.availableProviders)-1 {
		ls.selectedProviderIndex++
	}
}
