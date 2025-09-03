package termflow

import (
	"sort"
	"strings"
)

// CompletionItem represents a single completion suggestion
type CompletionItem struct {
	Text        string // The completion text
	Description string // Optional description
}

// CompletionProvider provides completion suggestions
type CompletionProvider struct {
	commands []CompletionItem
	prefixFn func(string) []CompletionItem
}

// NewCompletionProvider creates a new completion provider
func NewCompletionProvider() *CompletionProvider {
	return &CompletionProvider{
		commands: []CompletionItem{},
	}
}

// AddCommand adds a command to the completion list
func (cp *CompletionProvider) AddCommand(text, description string) {
	cp.commands = append(cp.commands, CompletionItem{
		Text:        text,
		Description: description,
	})
}

// SetDynamicProvider sets a function to provide dynamic completions based on input
func (cp *CompletionProvider) SetDynamicProvider(fn func(string) []CompletionItem) {
	cp.prefixFn = fn
}

// GetCompletions returns completion suggestions for the given input
func (cp *CompletionProvider) GetCompletions(input string) []CompletionItem {
	var completions []CompletionItem

	// Get static command completions
	for _, cmd := range cp.commands {
		if strings.HasPrefix(cmd.Text, input) {
			completions = append(completions, cmd)
		}
	}

	// Get dynamic completions if provider is set
	if cp.prefixFn != nil {
		dynamicCompletions := cp.prefixFn(input)
		completions = append(completions, dynamicCompletions...)
	}

	// Sort completions by text
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})

	// Remove duplicates
	seen := make(map[string]bool)
	unique := []CompletionItem{}
	for _, completion := range completions {
		if !seen[completion.Text] {
			seen[completion.Text] = true
			unique = append(unique, completion)
		}
	}

	return unique
}

// GetCompletionStrings returns just the completion text strings
func (cp *CompletionProvider) GetCompletionStrings(input string) []string {
	completions := cp.GetCompletions(input)
	result := make([]string, len(completions))
	for i, completion := range completions {
		result[i] = completion.Text
	}
	return result
}

// CompletionSelector manages completion selection with keyboard navigation
type CompletionSelector struct {
	items         []CompletionItem
	selectedIndex int
	visible       bool
}

// NewCompletionSelector creates a new completion selector
func NewCompletionSelector() *CompletionSelector {
	return &CompletionSelector{
		selectedIndex: -1,
		visible:       false,
	}
}

// SetItems sets the completion items
func (cs *CompletionSelector) SetItems(items []CompletionItem) {
	cs.items = items
	cs.selectedIndex = 0
	cs.visible = len(items) > 0
}

// Clear clears the completion selector
func (cs *CompletionSelector) Clear() {
	cs.items = []CompletionItem{}
	cs.selectedIndex = -1
	cs.visible = false
}

// IsVisible returns whether completions are currently visible
func (cs *CompletionSelector) IsVisible() bool {
	return cs.visible
}

// GetSelectedItem returns the currently selected completion item
func (cs *CompletionSelector) GetSelectedItem() *CompletionItem {
	if !cs.visible || cs.selectedIndex < 0 || cs.selectedIndex >= len(cs.items) {
		return nil
	}
	return &cs.items[cs.selectedIndex]
}

// MoveUp moves the selection up
func (cs *CompletionSelector) MoveUp() bool {
	if !cs.visible || len(cs.items) == 0 {
		return false
	}

	if cs.selectedIndex > 0 {
		cs.selectedIndex--
		return true
	}
	return false
}

// MoveDown moves the selection down
func (cs *CompletionSelector) MoveDown() bool {
	if !cs.visible || len(cs.items) == 0 {
		return false
	}

	if cs.selectedIndex < len(cs.items)-1 {
		cs.selectedIndex++
		return true
	}
	return false
}

// RenderCompletions renders the completion list for display
func (cs *CompletionSelector) RenderCompletions(maxItems int) string {
	if !cs.visible || len(cs.items) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("\nCompletions:\n")

	start := 0
	end := len(cs.items)

	if maxItems > 0 && end > maxItems {
		// Center the view around the selected item
		start = cs.selectedIndex - maxItems/2
		if start < 0 {
			start = 0
		}
		end = start + maxItems
		if end > len(cs.items) {
			end = len(cs.items)
			start = end - maxItems
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		item := cs.items[i]
		marker := "  "
		if i == cs.selectedIndex {
			marker = "▶ "
		}

		result.WriteString(marker)
		result.WriteString(item.Text)

		if item.Description != "" {
			result.WriteString(" - ")
			result.WriteString(item.Description)
		}

		result.WriteString("\n")
	}

	return result.String()
}
