package terminal

import (
	"strings"

	"github.com/mizzy/rigel/internal/ui/render"
)

// View renders the chat interface
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Render chat history using extracted render function
	renderHistory := make([]render.Exchange, len(m.history))
	for i, ex := range m.history {
		renderHistory[i] = render.Exchange{
			Prompt:   ex.Prompt,
			Response: ex.Response,
		}
	}
	s.WriteString(render.ChatHistory(renderHistory))

	// Display provider selection interface if in provider selection mode
	if m.providerSelectionMode {
		return render.ProviderSelector(m.availableProviders, m.selectedProviderIndex)
	}

	// Display model selection interface if in model selection mode
	if m.modelSelectionMode {
		return render.ModelSelector(m.filteredModels, m.selectedModelIndex, m.modelFilter)
	}

	// Display thinking state
	if m.thinking {
		s.WriteString(render.ThinkingState(m.currentPrompt, m.spinner.View()))
	}

	// Display input prompt and suggestions
	if !m.thinking {
		s.WriteString(render.InputPrompt(m.input.View()))

		// Display command suggestions using render function
		if m.showSuggestions && len(m.suggestions) > 0 {
			// Convert commands to render.Command format
			renderCommands := make([]render.Command, len(availableCommands))
			for i, cmd := range availableCommands {
				renderCommands[i] = render.Command{
					Command:     cmd.command,
					Description: cmd.description,
				}
			}
			s.WriteString(render.CommandSuggestions(m.suggestions, m.selectedSuggestion, renderCommands))
		}
	}

	// Display messages using render functions
	s.WriteString(render.InfoMessage(m.infoMessage))
	s.WriteString(render.ErrorMessage(m.err))

	return s.String()
}
