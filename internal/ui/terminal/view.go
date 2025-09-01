package terminal

import (
	"strings"

	"github.com/mizzy/rigel/internal/command"
	"github.com/mizzy/rigel/internal/ui/render"
)

// View renders the chat interface
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Render chat history using extracted render function
	history := m.chatState.GetHistory()
	renderHistory := make([]render.Exchange, len(history))
	for i, ex := range history {
		renderHistory[i] = render.Exchange{
			Prompt:   ex.Prompt,
			Response: ex.Response,
		}
	}
	s.WriteString(render.ChatHistory(renderHistory))

	// Display provider selection interface if in provider selection mode
	if m.llmState.IsProviderSelectionActive() {
		s.WriteString(render.ProviderSelector(m.llmState.GetAvailableProviders(), m.llmState.GetSelectedProviderIndex()))
		s.WriteString(render.InfoMessage(m.infoMessage))
		s.WriteString(render.ErrorMessage(m.chatState.GetError()))
		return s.String()
	}

	// Display model selection interface if in model selection mode
	if m.llmState.IsModelSelectionActive() {
		s.WriteString(render.ModelSelector(m.llmState.GetFilteredModels(), m.llmState.GetSelectedModelIndex(), m.llmState.GetModelFilter()))
		s.WriteString(render.InfoMessage(m.infoMessage))
		s.WriteString(render.ErrorMessage(m.chatState.GetError()))
		return s.String()
	}

	// Display thinking state
	if m.chatState.IsThinking() {
		s.WriteString(render.ThinkingState(m.chatState.GetCurrentPrompt(), m.spinner.View()))
		s.WriteString(render.InfoMessage(m.infoMessage))
		s.WriteString(render.ErrorMessage(m.chatState.GetError()))
		return s.String()
	}

	// Display input prompt and suggestions
	if !m.chatState.IsThinking() {
		// Display repository information above input prompt if available
		if m.gitInfo != nil {
			s.WriteString(render.RepoInfo(m.gitInfo.RepoName, m.gitInfo.Branch))
		}
		s.WriteString(render.InputPrompt(m.input.View()))

		// Display command completions using render function
		if m.showCompletions && len(m.completions) > 0 {
			// Convert commands to render.Command format
			renderCommands := make([]render.Command, len(command.AvailableCommands))
			for i, cmd := range command.AvailableCommands {
				renderCommands[i] = render.Command{
					Command:     cmd.Command,
					Description: cmd.Description,
				}
			}
			s.WriteString(render.CommandSuggestions(m.completions, m.selectedCompletion, renderCommands))
		}
	}

	// Display messages using render functions
	s.WriteString(render.InfoMessage(m.infoMessage))
	s.WriteString(render.ErrorMessage(m.chatState.GetError()))

	return s.String()
}
