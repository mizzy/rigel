package termflow

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// LineEditor provides line editing with history navigation
type LineEditor struct {
	client           *Client
	keyboard         *KeyboardReader
	history          []string
	historyIndex     int
	line             string
	cursor           int
	prompt           string
	ctrlCPressed     bool        // Track first Ctrl+C press for two-press exit
	exitMessageShown bool        // Track if exit message is shown below current line
	cursorOnExitLine bool        // Track if cursor is positioned on the line above exit message
	ctrlCTimer       *time.Timer // Timer to reset Ctrl+C state after 1 second
	displayedLines   int         // Track how many lines we've displayed
}

// NewLineEditor creates a new line editor
func NewLineEditor(client *Client) (*LineEditor, error) {
	keyboard, err := NewKeyboardReader()
	if err != nil {
		return nil, err
	}

	return &LineEditor{
		client:         client,
		keyboard:       keyboard,
		history:        []string{},
		historyIndex:   -1,
		line:           "",
		cursor:         0,
		prompt:         client.prompt,
		displayedLines: 0,
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
	le.displayedLines = 0

	// Show initial prompt
	le.refreshDisplay()

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

			// Reset flags when completing input
			le.ctrlCPressed = false
			le.exitMessageShown = false
			le.cursorOnExitLine = false
			le.displayedLines = 0

			// Stop timer if running
			le.stopCtrlCTimer()

			// Add to history if not empty and different from last entry
			if strings.TrimSpace(result) != "" {
				le.addToHistory(result)
			}

			return result, nil

		case KeyCtrlC:
			if le.ctrlCPressed {
				// Second Ctrl+C - return interrupted error to exit
				le.stopCtrlCTimer()
				return "", fmt.Errorf("interrupted")
			}
			// First Ctrl+C: don't redraw the input block to avoid duplication.
			// Simply print the exit hint on the next line and restore the cursor.
			le.ctrlCPressed = true
			le.exitMessageShown = true
			le.cursorOnExitLine = true
			// Show exit message on the next line
			fmt.Fprintf(le.client.output, "\n\r\033[38;5;240m(Press Ctrl+C again to exit)\033[0m")
			// Move cursor back up to the input line
			fmt.Fprintf(le.client.output, "\033[1A")
			// Position cursor depending on whether we're on first or continuation line
			textBeforeCursor := le.line[:le.cursor]
			linesBeforeCursor := strings.Split(textBeforeCursor, "\n")
			currentLineIndex := len(linesBeforeCursor) - 1
			currentColumn := len(linesBeforeCursor[len(linesBeforeCursor)-1])
			if currentLineIndex == 0 {
				fmt.Fprintf(le.client.output, "\r\033[%dC", visibleLength(le.prompt)+currentColumn)
			} else {
				fmt.Fprintf(le.client.output, "\r\033[%dC", 2+currentColumn)
			}

			// Start 1-second timer to reset Ctrl+C state and clear message
			le.startCtrlCTimer()
			continue // Continue input loop instead of returning

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

		case KeyCtrlJ:
			// Insert newline for multiline input
			le.insertRune('\n')
			le.refreshDisplay()
			continue

		case KeyRune:
			le.insertRune(key.Rune)

			// Use smart refresh: simple echo for single-line, no refresh for multiline character input
			if strings.Contains(le.line, "\n") {
				// We're in multiline mode - just echo the character, don't do full refresh
				// This prevents the duplication issue
				fmt.Fprint(le.client.output, string(key.Rune))
			} else {
				// Single line mode - use simple character echo
				fmt.Fprint(le.client.output, string(key.Rune))
			}

		default:
			// Ignore unknown keys
			continue
		}

		// For non-KeyRune cases that need refresh, it's handled in the specific case
		// No automatic refresh here to avoid duplication
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

// refreshDisplay redraws the current line(s) with multiline support
func (le *LineEditor) refreshDisplay() {
	lines := strings.Split(le.line, "\n")

	// Compute cursor location in input text for final positioning
	textBeforeCursor := le.line[:le.cursor]
	linesBeforeCursor := strings.Split(textBeforeCursor, "\n")
	currentLineIndex := len(linesBeforeCursor) - 1
	currentColumn := len(linesBeforeCursor[len(linesBeforeCursor)-1])

	// Move to the top of the previously drawn input block and clear it
	n := le.displayedLines
	if n > 0 {
		if n > 1 {
			fmt.Fprintf(le.client.output, "\033[%dA", n-1)
		}
		fmt.Fprint(le.client.output, "\r")
		for i := 0; i < n; i++ {
			fmt.Fprint(le.client.output, "\033[K")
			if i < n-1 {
				fmt.Fprint(le.client.output, "\033[1B\r")
			}
		}
		if n > 1 {
			fmt.Fprintf(le.client.output, "\033[%dA\r", n-1)
		} else {
			fmt.Fprint(le.client.output, "\r")
		}
	}

	// Draw fresh content (no leading newline; spacer is provided by welcome)
	fmt.Fprint(le.client.output, le.prompt)
	fmt.Fprint(le.client.output, lines[0])
	for i := 1; i < len(lines); i++ {
		fmt.Fprint(le.client.output, "\n\r  ")
		fmt.Fprint(le.client.output, lines[i])
	}

	// Position cursor
	if len(lines) > 1 {
		fmt.Fprintf(le.client.output, "\033[%dA", len(lines)-1)
	}
	fmt.Fprint(le.client.output, "\r")
	if currentLineIndex > 0 {
		fmt.Fprintf(le.client.output, "\033[%dB", currentLineIndex)
		fmt.Fprintf(le.client.output, "\033[%dC", 2+currentColumn)
	} else {
		fmt.Fprintf(le.client.output, "\033[%dC", visibleLength(le.prompt)+currentColumn)
	}

	le.displayedLines = len(lines)
}

// refreshDisplayWithoutPrompt redraws the current line(s) without showing the prompt
func (le *LineEditor) refreshDisplayWithoutPrompt() {
	lines := strings.Split(le.line, "\n")

	// Calculate cursor position
	textBeforeCursor := le.line[:le.cursor]
	linesBeforeCursor := strings.Split(textBeforeCursor, "\n")
	targetLine := len(linesBeforeCursor) - 1
	targetColumn := len(linesBeforeCursor[len(linesBeforeCursor)-1])

	// Move to end of first line (after the prompt and first line content)
	fmt.Fprint(le.client.output, "\r")
	fmt.Fprint(le.client.output, "\033[999C") // Move to end of line

	// Clear all continuation lines aggressively
	maxLinesToClear := 10 // Clear up to 10 lines to be safe
	for i := 0; i < maxLinesToClear; i++ {
		fmt.Fprint(le.client.output, "\n\r\033[K") // Move down, carriage return, then clear line
	}

	// Move back to end of first line
	fmt.Fprintf(le.client.output, "\033[%dA", maxLinesToClear) // Move back up
	fmt.Fprint(le.client.output, "\r\033[999C")                // Move to end of first line

	// Display only continuation lines (skip first line which already has prompt)
	for i := 1; i < len(lines); i++ {
		fmt.Fprint(le.client.output, "\n\r  ") // Newline + carriage return + 2 spaces for alignment
		fmt.Fprint(le.client.output, lines[i])
	}

	// Position cursor correctly
	// Move to beginning of first line
	if len(lines) > 1 {
		fmt.Fprintf(le.client.output, "\033[%dA", len(lines)-1)
	}
	fmt.Fprint(le.client.output, "\r")

	// Move down to target line
	if targetLine > 0 {
		fmt.Fprintf(le.client.output, "\033[%dB", targetLine)
	}

	// Move to target column, accounting for alignment on continuation lines
	if targetLine == 0 {
		// First line - no additional alignment
		fmt.Fprintf(le.client.output, "\033[%dC", targetColumn)
	} else {
		// Subsequent lines have 2-space alignment + target column
		totalColumn := 2 + targetColumn
		if totalColumn > 0 {
			fmt.Fprintf(le.client.output, "\033[%dC", totalColumn)
		}
	}

	// Update displayed lines count for next refresh
	le.displayedLines = len(lines)
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
	le.displayedLines = 0

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

			// Reset flags when completing input
			le.ctrlCPressed = false
			le.exitMessageShown = false
			le.cursorOnExitLine = false
			le.displayedLines = 0

			// Stop timer if running
			le.stopCtrlCTimer()

			// Add to history if not empty and different from last entry
			if strings.TrimSpace(result) != "" {
				le.addToHistory(result)
			}

			return result, nil

		case KeyCtrlC:
			if le.ctrlCPressed {
				// Second Ctrl+C - return interrupted error to exit
				le.stopCtrlCTimer()
				return "", fmt.Errorf("interrupted")
			}
			// First Ctrl+C: avoid redrawing the input to prevent duplicate lines.
			le.ctrlCPressed = true
			le.exitMessageShown = true
			le.cursorOnExitLine = true
			// Show exit message on the next line
			fmt.Fprintf(le.client.output, "\n\r\033[38;5;240m(Press Ctrl+C again to exit)\033[0m")
			// Move cursor back up to the input line
			fmt.Fprintf(le.client.output, "\033[1A")
			// Position cursor depending on whether we're on first or continuation line
			textBeforeCursor := le.line[:le.cursor]
			linesBeforeCursor := strings.Split(textBeforeCursor, "\n")
			currentLineIndex := len(linesBeforeCursor) - 1
			currentColumn := len(linesBeforeCursor[len(linesBeforeCursor)-1])
			if currentLineIndex == 0 {
				fmt.Fprintf(le.client.output, "\r\033[%dC", visibleLength(le.prompt)+currentColumn)
			} else {
				fmt.Fprintf(le.client.output, "\r\033[%dC", 2+currentColumn)
			}

			// Start 1-second timer to reset Ctrl+C state and clear message
			le.startCtrlCTimer()
			continue // Continue input loop instead of returning

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

		case KeyCtrlJ:
			// Insert newline for multiline input
			le.insertRune('\n')
			le.refreshDisplayWithoutPrompt()

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

// visibleLength calculates the visible length of a string by removing ANSI escape sequences
func visibleLength(s string) int {
	// Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	cleaned := ansiRegex.ReplaceAllString(s, "")
	// Count Unicode runes (characters) not bytes
	return utf8.RuneCountInString(cleaned)
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

// startCtrlCTimer starts a 1-second timer to reset Ctrl+C state and clear exit message
func (le *LineEditor) startCtrlCTimer() {
	// Stop any existing timer first
	le.stopCtrlCTimer()

	le.ctrlCTimer = time.AfterFunc(1*time.Second, func() {
		// Reset Ctrl+C state
		le.ctrlCPressed = false
		le.exitMessageShown = false
		le.cursorOnExitLine = false

		// Clear the exit message by redrawing the display
		// Move cursor down to the exit message line and clear it
		fmt.Fprint(le.client.output, "\033[1B")  // Move down 1 line to the exit message
		fmt.Fprint(le.client.output, "\r\033[K") // Clear the line

		// Move back up to the input line and reposition cursor accurately
		fmt.Fprint(le.client.output, "\033[1A") // Move up 1 line
		textBeforeCursor := le.line[:le.cursor]
		linesBeforeCursor := strings.Split(textBeforeCursor, "\n")
		currentLineIndex := len(linesBeforeCursor) - 1
		currentColumn := len(linesBeforeCursor[len(linesBeforeCursor)-1])
		if currentLineIndex == 0 {
			fmt.Fprintf(le.client.output, "\r\033[%dC", visibleLength(le.prompt)+currentColumn)
		} else {
			fmt.Fprintf(le.client.output, "\r\033[%dC", 2+currentColumn)
		}
	})
}

// stopCtrlCTimer stops the Ctrl+C reset timer if it's running
func (le *LineEditor) stopCtrlCTimer() {
	if le.ctrlCTimer != nil {
		le.ctrlCTimer.Stop()
		le.ctrlCTimer = nil
	}
}
