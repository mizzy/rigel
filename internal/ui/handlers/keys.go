package handlers

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleSpecialKeys handles special key combinations like Ctrl+C
func HandleSpecialKeys(msg tea.KeyMsg, thinking bool, ctrlCPressed *bool, quitting *bool) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyCtrlC:
		if !*ctrlCPressed {
			*ctrlCPressed = true
			return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
				return "reset_ctrl_c"
			}), true
		}
		*quitting = true
		return tea.Quit, true

	case tea.KeyEsc:
		if thinking {
			return nil, true // Don't allow escape during thinking
		}
		return nil, false
	}

	return nil, false
}

// HandleNavigationKeys handles up/down arrow keys for navigation
func HandleNavigationKeys(msg tea.KeyMsg, selectedIndex *int, maxIndex int) bool {
	switch msg.Type {
	case tea.KeyUp:
		if *selectedIndex > 0 {
			*selectedIndex--
		}
		return true

	case tea.KeyDown:
		if *selectedIndex < maxIndex-1 {
			*selectedIndex++
		}
		return true
	}

	return false
}

// HandleFilterInput handles text input for filtering
func HandleFilterInput(input string, oldFilter string, items []interface{}, filterFunc func([]interface{}, string) []interface{}) ([]interface{}, bool) {
	if input != oldFilter {
		filtered := filterFunc(items, input)
		return filtered, true
	}
	return nil, false
}
