package termflow

import (
	"fmt"
	"io"
	"strings"
)

// LineEditor provides line editing with history navigation
type LineEditor struct {
	client       *Client
	keyboard     *KeyboardReader
	history      []string
	historyIndex int
	line         string
	cursor       int
	prompt       string
}

// NewLineEditor creates a new line editor
func NewLineEditor(client *Client) (*LineEditor, error) {
	keyboard, err := NewKeyboardReader()
	if err != nil {
		return nil, err
	}

	return &LineEditor{
		client:       client,
		keyboard:     keyboard,
		history:      []string{},
		historyIndex: -1,
		line:         "",
		cursor:       0,
		prompt:       client.prompt,
	}, nil
}

// SetHistory sets the command history for navigation
func (le *LineEditor) SetHistory(history []string) {
	le.history = append([]string{}, history...) // Copy
	le.historyIndex = -1
}

// ReadLineWithHistory reads a line with arrow key history navigation
func (le *LineEditor) ReadLineWithHistory() (string, error) {
	// Enable raw mode for key-by-key input
	if err := le.keyboard.EnableRawMode(); err != nil {
		// Fall back to regular ReadLine if raw mode fails
		return le.client.ReadLine()
	}
	defer le.keyboard.DisableRawMode()

	// Initialize line state
	le.line = ""
	le.cursor = 0
	le.historyIndex = -1

	// Show initial prompt
	fmt.Fprint(le.client.output, le.prompt)

	for {
		key, err := le.keyboard.ReadKey()
		if err != nil {
			return "", err
		}

		switch key.Type {
		case KeyEnter:
			// Finish input
			fmt.Fprint(le.client.output, "\n")
			result := le.line

			// Add to history if not empty and different from last entry
			if strings.TrimSpace(result) != "" {
				le.addToHistory(result)
			}

			return result, nil

		case KeyCtrlC:
			// Clear current line and move to next line cleanly
			fmt.Fprint(le.client.output, "\r\033[K\n")
			return "", fmt.Errorf("interrupted")

		case KeyCtrlD:
			if len(le.line) == 0 {
				fmt.Fprint(le.client.output, "\n")
				return "", fmt.Errorf("EOF")
			}
			// If there's content, Ctrl+D does nothing

		case KeyBackspace:
			le.handleBackspace()

		case KeyDelete:
			le.handleDelete()

		case KeyArrowLeft:
			le.moveCursorLeft()

		case KeyArrowRight:
			le.moveCursorRight()

		case KeyArrowUp:
			le.navigateHistory(-1)

		case KeyArrowDown:
			le.navigateHistory(1)

		case KeyTab:
			// TODO: Implement tab completion
			continue

		case KeyRune:
			le.insertRune(key.Rune)

		default:
			// Ignore unknown keys
			continue
		}

		// Refresh the display
		le.refreshDisplay()
	}
}

// handleBackspace removes character before cursor
func (le *LineEditor) handleBackspace() {
	if le.cursor > 0 {
		le.line = le.line[:le.cursor-1] + le.line[le.cursor:]
		le.cursor--
	}
}

// handleDelete removes character at cursor
func (le *LineEditor) handleDelete() {
	if le.cursor < len(le.line) {
		le.line = le.line[:le.cursor] + le.line[le.cursor+1:]
	}
}

// insertRune inserts a rune at the cursor position
func (le *LineEditor) insertRune(r rune) {
	le.line = le.line[:le.cursor] + string(r) + le.line[le.cursor:]
	le.cursor++
}

// moveCursorLeft moves cursor one position left
func (le *LineEditor) moveCursorLeft() {
	if le.cursor > 0 {
		le.cursor--
	}
}

// moveCursorRight moves cursor one position right
func (le *LineEditor) moveCursorRight() {
	if le.cursor < len(le.line) {
		le.cursor++
	}
}

// navigateHistory navigates through command history
func (le *LineEditor) navigateHistory(direction int) {
	if len(le.history) == 0 {
		return
	}

	newIndex := le.historyIndex + direction

	if direction < 0 { // Up arrow
		if le.historyIndex == -1 {
			// First time pressing up, go to most recent
			newIndex = len(le.history) - 1
		} else if newIndex < 0 {
			// Already at oldest, don't go further
			return
		}
	} else { // Down arrow
		if newIndex >= len(le.history) {
			// Go back to empty line
			le.historyIndex = -1
			le.line = ""
			le.cursor = 0
			return
		}
	}

	if newIndex >= 0 && newIndex < len(le.history) {
		le.historyIndex = newIndex
		le.line = le.history[newIndex]
		le.cursor = len(le.line)
	}
}

// refreshDisplay redraws the current line
func (le *LineEditor) refreshDisplay() {
	// Clear the current line
	fmt.Fprint(le.client.output, "\r\033[K")

	// Print prompt and current line
	fmt.Fprint(le.client.output, le.prompt)
	fmt.Fprint(le.client.output, le.line)

	// Position cursor correctly
	if le.cursor < len(le.line) {
		// Move cursor to correct position
		fmt.Fprintf(le.client.output, "\033[%dD", len(le.line)-le.cursor)
	}
}

// refreshDisplayWithoutPrompt redraws the current line without showing the prompt
func (le *LineEditor) refreshDisplayWithoutPrompt() {
	// Clear the current line
	fmt.Fprint(le.client.output, "\r\033[K")

	// Print only the current line (no prompt)
	fmt.Fprint(le.client.output, le.line)

	// Position cursor correctly
	if le.cursor < len(le.line) {
		// Move cursor to correct position
		fmt.Fprintf(le.client.output, "\033[%dD", len(le.line)-le.cursor)
	}
}

// ReadLineWithoutPrompt reads input without showing the initial prompt
func (le *LineEditor) ReadLineWithoutPrompt() (string, error) {
	// Enable raw mode for key-by-key input
	if err := le.keyboard.EnableRawMode(); err != nil {
		// Fall back to regular ReadLine if raw mode fails
		return le.client.ReadLine()
	}
	defer le.keyboard.DisableRawMode()

	// Initialize line state
	le.line = ""
	le.cursor = 0
	le.historyIndex = -1

	// Don't show initial prompt - this is the key difference

	for {
		key, err := le.keyboard.ReadKey()
		if err != nil {
			return "", err
		}

		switch key.Type {
		case KeyEnter:
			// Finish input
			fmt.Fprint(le.client.output, "\n")
			result := le.line

			// Add to history if not empty and different from last entry
			if strings.TrimSpace(result) != "" {
				le.addToHistory(result)
			}

			return result, nil

		case KeyCtrlC:
			// Clear current line and move to next line cleanly
			fmt.Fprint(le.client.output, "\r\033[K\n")
			return "", fmt.Errorf("interrupted")

		case KeyCtrlD:
			if le.line == "" {
				// EOF on empty line
				fmt.Fprint(le.client.output, "\n")
				return "", io.EOF
			}
			// Otherwise ignore Ctrl+D when there's text

		case KeyBackspace:
			if le.cursor > 0 {
				// Remove character before cursor
				le.line = le.line[:le.cursor-1] + le.line[le.cursor:]
				le.cursor--
				le.refreshDisplayWithoutPrompt()
			}

		case KeyDelete:
			if le.cursor < len(le.line) {
				// Remove character at cursor
				le.line = le.line[:le.cursor] + le.line[le.cursor+1:]
				le.refreshDisplayWithoutPrompt()
			}

		case KeyArrowLeft:
			if le.cursor > 0 {
				le.cursor--
				fmt.Fprint(le.client.output, "\033[1D") // Move cursor left
			}

		case KeyArrowRight:
			if le.cursor < len(le.line) {
				le.cursor++
				fmt.Fprint(le.client.output, "\033[1C") // Move cursor right
			}

		case KeyArrowUp:
			if le.historyIndex < len(le.history)-1 {
				le.historyIndex++
				le.line = le.history[len(le.history)-1-le.historyIndex]
				le.cursor = len(le.line)
				le.refreshDisplayWithoutPrompt()
			}

		case KeyArrowDown:
			if le.historyIndex >= 0 {
				le.historyIndex--
				if le.historyIndex >= 0 {
					le.line = le.history[len(le.history)-1-le.historyIndex]
				} else {
					le.line = ""
				}
				le.cursor = len(le.line)
				le.refreshDisplayWithoutPrompt()
			}

		case KeyRune:
			// Insert character at cursor position
			if le.cursor >= len(le.line) {
				le.line += string(key.Rune)
			} else {
				le.line = le.line[:le.cursor] + string(key.Rune) + le.line[le.cursor:]
			}
			le.cursor++
			le.refreshDisplayWithoutPrompt()
		}
	}
}

// addToHistory adds a command to the history
func (le *LineEditor) addToHistory(command string) {
	// Don't add duplicate of last command
	if len(le.history) > 0 && le.history[len(le.history)-1] == command {
		return
	}

	le.history = append(le.history, command)

	// Keep history size reasonable
	maxHistory := 1000
	if len(le.history) > maxHistory {
		le.history = le.history[1:]
	}
}
