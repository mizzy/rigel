package handlers

import (
	"github.com/charmbracelet/bubbles/textarea"
)

// HistoryNavigationState holds the state needed for history navigation
type HistoryNavigationState struct {
	InputHistory []string
	HistoryIndex int
	CurrentInput string
}

// NavigateHistory navigates through input history and updates the input field
func NavigateHistory(direction int, input *textarea.Model, state *HistoryNavigationState) {
	if len(state.InputHistory) == 0 {
		return
	}

	// Save current input if we're starting to navigate history
	if state.HistoryIndex == -1 {
		state.CurrentInput = input.Value()
	}

	if direction < 0 {
		// Going up (backward) in history
		if state.HistoryIndex == -1 {
			// Start from the most recent item
			state.HistoryIndex = 0
		} else if state.HistoryIndex < len(state.InputHistory)-1 {
			state.HistoryIndex++
		}

		if state.HistoryIndex < len(state.InputHistory) {
			historyPos := len(state.InputHistory) - 1 - state.HistoryIndex
			input.SetValue(state.InputHistory[historyPos])
		}
	} else {
		// Going down (forward) in history
		if state.HistoryIndex > 0 {
			state.HistoryIndex--
			historyPos := len(state.InputHistory) - 1 - state.HistoryIndex
			input.SetValue(state.InputHistory[historyPos])
		} else if state.HistoryIndex == 0 {
			// Return to current input
			state.HistoryIndex = -1
			input.SetValue(state.CurrentInput)
		}
	}
}
