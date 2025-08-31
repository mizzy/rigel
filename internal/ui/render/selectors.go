package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mizzy/rigel/internal/llm"
)

// ModelSelector renders the model selection interface
func ModelSelector(models []llm.Model, selectedIndex int, filter string) string {
	if len(models) == 0 {
		return "No models available"
	}

	var sb strings.Builder
	sb.WriteString("Select a model:\n\n")

	// Show filter if active
	if filter != "" {
		sb.WriteString(fmt.Sprintf("Filter: %s\n\n", filter))
	}

	// Display models
	for i, model := range models {
		displayName := model.Name
		if model.Details.Family != "" {
			displayName = fmt.Sprintf("%s (%s)", model.Name, model.Details.Family)
		}

		if i == selectedIndex {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)
			sb.WriteString(style.Render(fmt.Sprintf("> %s", displayName)))
		} else {
			sb.WriteString(fmt.Sprintf("  %s", displayName))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString("↑/↓: navigate • Enter: select • /: filter • Esc: cancel")

	return sb.String()
}

// ProviderSelector renders the provider selection interface
func ProviderSelector(providers []string, selectedIndex int) string {
	if len(providers) == 0 {
		return "No providers available"
	}

	var sb strings.Builder
	sb.WriteString("Select a provider:\n\n")

	// Display providers
	for i, provider := range providers {
		if i == selectedIndex {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)
			sb.WriteString(style.Render(fmt.Sprintf("> %s", provider)))
		} else {
			sb.WriteString(fmt.Sprintf("  %s", provider))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString("↑/↓: navigate • Enter: select • Esc: cancel")

	return sb.String()
}
