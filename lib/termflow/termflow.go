// Package termflow provides a minimal terminal UI library that preserves scrollback history.
// Unlike full TUI frameworks that take over the entire screen, termflow uses normal
// terminal output for preserving chat history while providing interactive input features.
package termflow

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Client represents a termflow terminal interface client
type Client struct {
	input          io.Reader
	output         io.Writer
	prompt         string
	history        []string
	maxHistory     int
	completionFunc CompletionFunc
	reader         *bufio.Reader
}

// CompletionFunc is called to provide completion suggestions for user input
type CompletionFunc func(input string) []string

// New creates a new termflow client
func New() *Client {
	return &Client{
		input:      os.Stdin,
		output:     os.Stdout,
		prompt:     "\033[1;38;5;87m✦\033[0m ",
		maxHistory: 1000,
		reader:     bufio.NewReader(os.Stdin),
	}
}

// SetPrompt sets the input prompt string
func (c *Client) SetPrompt(prompt string) {
	c.prompt = prompt
}

// SetCompletionFunc sets the tab completion function
func (c *Client) SetCompletionFunc(fn CompletionFunc) {
	c.completionFunc = fn
}

// Print outputs text to the terminal (preserved in scrollback)
func (c *Client) Print(text string) {
	fmt.Fprint(c.output, text)
}

// Printf outputs formatted text to the terminal (preserved in scrollback)
func (c *Client) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.output, format, args...)
}

// PrintChat outputs a chat exchange (user input + AI response) with formatting
func (c *Client) PrintChat(userInput, aiResponse string) {
	// User prompt with ✦ symbol (same as bubbletea)
	c.Printf("\033[1;38;5;87m✦\033[0m \033[38;5;195m%s\033[0m\n\n", userInput)
	// AI response with normal terminal color
	c.Printf("\033[38;5;252m%s\033[0m\n\n", aiResponse)
}

// PrintResponse outputs only the AI response (user input is already visible)
func (c *Client) PrintResponse(response string) {
	// AI response with normal terminal color, preceded by newline for spacing
	c.Printf("\n\033[38;5;252m%s\033[0m\n\n", response)
}

// ReadLine reads a line of input from the user with history support
func (c *Client) ReadLine() (string, error) {
	// Show prompt
	fmt.Fprint(c.output, c.prompt)

	// Read input
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Clean up the input
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	// Add to history if not empty
	if strings.TrimSpace(line) != "" {
		c.addToHistory(line)
	}

	return line, nil
}

// ReadMultiLine reads multiple lines of input until the user enters a line containing only "."
// or presses Ctrl+D. Returns the combined input as a single string.
func (c *Client) ReadMultiLine() (string, error) {
	var lines []string
	lineNum := 1

	c.Printf("Enter multiple lines (type '.' on empty line or Ctrl+D to finish):\n")

	for {
		// Show line number prompt
		fmt.Fprintf(c.output, "%2d> ", lineNum)

		// Read line
		line, err := c.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Ctrl+D pressed, finish input
				break
			}
			return "", err
		}

		// Clean up the line
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		// Check for end marker
		if line == "." {
			break
		}

		lines = append(lines, line)
		lineNum++
	}

	result := strings.Join(lines, "\n")

	// Add to history if not empty
	if strings.TrimSpace(result) != "" {
		c.addToHistory(result)
	}

	return result, nil
}

// ReadLineOrMultiLine reads user input.
// In interactive mode, multiline editing is supported via Ctrl+J.
func (c *Client) ReadLineOrMultiLine() (string, error) {
	// Show prompt
	fmt.Fprint(c.output, c.prompt)

	// Read a line (base client cannot intercept Ctrl+J newlines in canonical mode)
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Clean up the input
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	// Add to history if not empty
	if strings.TrimSpace(line) != "" {
		c.addToHistory(line)
	}

	return line, nil
}

// addToHistory adds an entry to the input history
func (c *Client) addToHistory(entry string) {
	c.history = append(c.history, entry)

	// Limit history size
	if len(c.history) > c.maxHistory {
		c.history = c.history[1:]
	}
}

// GetHistory returns the input history
func (c *Client) GetHistory() []string {
	return append([]string{}, c.history...) // Return a copy
}

// ClearHistory clears the input history
func (c *Client) ClearHistory() {
	c.history = []string{}
}
