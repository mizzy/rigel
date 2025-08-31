package chat

// navigateHistory navigates through input history
func (m *Model) navigateHistory(direction int) {
	if len(m.inputHistory) == 0 {
		return
	}

	// Save current input if we're starting to navigate history
	if m.historyIndex == -1 {
		m.currentInput = m.input.Value()
	}

	if direction < 0 {
		// Going up (backward) in history
		if m.historyIndex == -1 {
			// Start from the most recent item
			m.historyIndex = 0
		} else if m.historyIndex < len(m.inputHistory)-1 {
			m.historyIndex++
		}

		if m.historyIndex < len(m.inputHistory) {
			historyPos := len(m.inputHistory) - 1 - m.historyIndex
			m.input.SetValue(m.inputHistory[historyPos])
		}
	} else {
		// Going down (forward) in history
		if m.historyIndex > 0 {
			m.historyIndex--
			historyPos := len(m.inputHistory) - 1 - m.historyIndex
			m.input.SetValue(m.inputHistory[historyPos])
		} else if m.historyIndex == 0 {
			// Return to current input
			m.historyIndex = -1
			m.input.SetValue(m.currentInput)
		}
	}
}
