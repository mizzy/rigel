package termflow

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// InteractiveClient provides advanced input features like history navigation and tab completion
type InteractiveClient struct {
	*Client
	rawMode    bool
	oldState   *term.State
	lineEditor *LineEditor
}

// NewInteractive creates a new interactive termflow client with advanced input features
func NewInteractive() (*InteractiveClient, error) {
	baseClient := New()

	lineEditor, err := NewLineEditor(baseClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create line editor: %w", err)
	}

	client := &InteractiveClient{
		Client:     baseClient,
		lineEditor: lineEditor,
	}

	return client, nil
}

// EnableRawMode enables raw terminal mode for advanced input handling
func (ic *InteractiveClient) EnableRawMode() error {
	if ic.rawMode {
		return nil
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}

	ic.oldState = oldState
	ic.rawMode = true
	return nil
}

// DisableRawMode disables raw terminal mode
func (ic *InteractiveClient) DisableRawMode() error {
	if !ic.rawMode {
		return nil
	}

	fd := int(os.Stdin.Fd())
	err := term.Restore(fd, ic.oldState)
	ic.rawMode = false
	return err
}

// ReadLineInteractive reads input with history navigation and tab completion support
func (ic *InteractiveClient) ReadLineInteractive() (string, error) {
	// Set history in line editor
	ic.lineEditor.SetHistory(ic.history)

	// Use line editor for input with cursor key support
	line, err := ic.lineEditor.ReadLineWithHistory()
	if err != nil {
		// For interruption, return error immediately without fallback
		if err.Error() == "interrupted" {
			return "", err
		}
		// Fall back to regular ReadLine if line editor fails
		return ic.Client.ReadLine()
	}

	// Add to history if not empty
	if strings.TrimSpace(line) != "" {
		ic.addToHistory(line)
	}

	return line, nil
}

// ReadLineOrMultiLine reads input with cursor key support and multiline detection
func (ic *InteractiveClient) ReadLineOrMultiLine() (string, error) {
	// Set history in line editor
	ic.lineEditor.SetHistory(ic.history)

	// Use line editor for first line input
	line, err := ic.lineEditor.ReadLineWithHistory()
	if err != nil {
		// For interruption, return error immediately without fallback
		if err.Error() == "interrupted" {
			return "", err
		}
		// Fall back to regular ReadLineOrMultiLine if line editor fails
		return ic.Client.ReadLineOrMultiLine()
	}

	// Check if user wants multiline input
	if strings.HasSuffix(line, "...") {
		// Remove the "..." marker and start multiline input
		firstLine := strings.TrimSuffix(line, "...")

		var lines []string
		if strings.TrimSpace(firstLine) != "" {
			lines = append(lines, firstLine)
		}

		ic.Printf("Continue typing (type '.' on empty line or Ctrl+D to finish):\n")
		lineNum := 2

		for {
			// Show continuation prompt - use regular reader for continuation lines
			fmt.Fprintf(ic.output, "%2d> ", lineNum)

			nextLine, err := ic.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", err
			}

			// Clean up the line
			nextLine = strings.TrimSuffix(nextLine, "\n")
			nextLine = strings.TrimSuffix(nextLine, "\r")

			// Check for end marker
			if nextLine == "." {
				break
			}

			lines = append(lines, nextLine)
			lineNum++
		}

		result := strings.Join(lines, "\n")

		// Add to history if not empty
		if strings.TrimSpace(result) != "" {
			ic.addToHistory(result)
		}

		return result, nil
	}

	// Single line input - add to history
	if strings.TrimSpace(line) != "" {
		ic.addToHistory(line)
	}

	return line, nil
}

// ReadLineOrMultiLineWithoutPrompt reads input without showing initial prompt (for resuming after Ctrl+C)
func (ic *InteractiveClient) ReadLineOrMultiLineWithoutPrompt() (string, error) {
	// Set history in line editor
	ic.lineEditor.SetHistory(ic.history)

	// Use line editor for first line input without initial prompt
	line, err := ic.lineEditor.ReadLineWithoutPrompt()
	if err != nil {
		// For interruption, return error immediately without fallback
		if err.Error() == "interrupted" {
			return "", err
		}
		// Fall back to regular ReadLineOrMultiLine if line editor fails
		return ic.Client.ReadLineOrMultiLine()
	}

	// Check if user wants multiline input (same logic as ReadLineOrMultiLine)
	if strings.HasSuffix(line, "...") {
		// Remove the "..." marker and start multiline input
		firstLine := strings.TrimSuffix(line, "...")

		var lines []string
		if strings.TrimSpace(firstLine) != "" {
			lines = append(lines, firstLine)
		}

		ic.Printf("Continue typing (type '.' on empty line or Ctrl+D to finish):\n")
		lineNum := 2

		for {
			// Show continuation prompt - use regular reader for continuation lines
			fmt.Fprintf(ic.output, "%2d> ", lineNum)

			nextLine, err := ic.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", err
			}

			// Clean up the line
			nextLine = strings.TrimSuffix(nextLine, "\n")
			nextLine = strings.TrimSuffix(nextLine, "\r")

			// Check for end marker
			if nextLine == "." {
				break
			}

			lines = append(lines, nextLine)
			lineNum++
		}

		result := strings.Join(lines, "\n")

		// Add to history if not empty
		if strings.TrimSpace(result) != "" {
			ic.addToHistory(result)
		}

		return result, nil
	}

	// Single line input - add to history
	if strings.TrimSpace(line) != "" {
		ic.addToHistory(line)
	}

	return line, nil
}

// Close cleans up the interactive client
func (ic *InteractiveClient) Close() error {
	if ic.rawMode {
		return ic.DisableRawMode()
	}
	return nil
}

// SetHistory sets the command history for the interactive client
func (ic *InteractiveClient) SetHistory(history []string) {
	ic.history = append([]string{}, history...) // Copy the history
}

// ShowCompletions displays available completions
func (ic *InteractiveClient) ShowCompletions(input string, completions []string) {
	if len(completions) == 0 {
		return
	}

	ic.Printf("\nCompletions:\n")
	for i, completion := range completions {
		marker := "  "
		if i == 0 {
			marker = "▶ "
		}
		ic.Printf("%s%s\n", marker, completion)
	}
	ic.Printf("\n")
}

// ShowError displays an error message
func (ic *InteractiveClient) ShowError(err error) {
	ic.Printf("\n\n\033[38;5;196mError: %v\033[0m", err)
}

// ShowInfo displays an info message
func (ic *InteractiveClient) ShowInfo(message string) {
	ic.Printf("\n\n\033[38;5;240m%s\033[0m\n", message)
}

// ShowInfoInline displays an info message while preserving current cursor position
// This is used for Ctrl+C to show the exit message below the current prompt
func (ic *InteractiveClient) ShowInfoInline(message string) {
	// Save current cursor position
	fmt.Fprintf(ic.output, "\0337") // Save cursor position (ESC 7)

	// Move to next line and show the message
	fmt.Fprintf(ic.output, "\n\033[38;5;240m%s\033[0m", message)

	// Restore cursor to original position
	fmt.Fprintf(ic.output, "\0338") // Restore cursor position (ESC 8)
}

// ShowThinking displays a thinking indicator
func (ic *InteractiveClient) ShowThinking(message string) {
	ic.Printf("\n\033[3;38;5;117m%s\033[0m\n", message)
}

// ShowThinkingWithSpinner displays a thinking indicator with animated spinner
func (ic *InteractiveClient) ShowThinkingWithSpinner(message string) *ThinkingSpinner {
	ts := NewThinkingSpinner(ic, message)
	ts.Start()
	return ts
}
