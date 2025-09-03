// Package testing provides UI testing capabilities for termflow applications
package uitest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
)

// TerminalTest represents a terminal testing session
type TerminalTest struct {
	cmd    *exec.Cmd
	ptmx   *os.File
	output *bytes.Buffer
	t      *testing.T
}

// NewTerminalTest creates a new terminal test session
func NewTerminalTest(t *testing.T, command string, args ...string) (*TerminalTest, error) {
	cmd := exec.Command(command, args...)

	// Inherit environment variables from test process
	cmd.Env = os.Environ()

	// Create a pseudo-terminal with size
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}

	tt := &TerminalTest{
		cmd:    cmd,
		ptmx:   ptmx,
		output: &bytes.Buffer{},
		t:      t,
	}

	// Start capturing output
	go tt.captureOutput()

	return tt, nil
}

// captureOutput continuously reads from the terminal and stores output
func (tt *TerminalTest) captureOutput() {
	buf := make([]byte, 1024)
	for {
		n, err := tt.ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				tt.t.Logf("PTY read error: %v", err)
			}
			return
		}
		if n > 0 {
			tt.t.Logf("PTY read %d bytes: %q", n, string(buf[:n]))
			tt.output.Write(buf[:n])
		}
	}
}

// SendKeys sends keystrokes to the terminal
func (tt *TerminalTest) SendKeys(keys string) error {
	_, err := tt.ptmx.Write([]byte(keys))
	return err
}

// SendCtrlC sends Ctrl+C to the terminal
func (tt *TerminalTest) SendCtrlC() error {
	return tt.SendKeys("\x03")
}

// SendCtrlJ sends Ctrl+J to the terminal
func (tt *TerminalTest) SendCtrlJ() error {
	return tt.SendKeys("\x0A")
}

// SendEnter sends Enter key to the terminal
func (tt *TerminalTest) SendEnter() error {
	return tt.SendKeys("\r\n")
}

// Type sends text followed by Enter
func (tt *TerminalTest) Type(text string) error {
	if err := tt.SendKeys(text); err != nil {
		return err
	}
	return tt.SendEnter()
}

// Wait waits for specified duration to let output settle
func (tt *TerminalTest) Wait(duration time.Duration) {
	time.Sleep(duration)
}

// GetOutput returns the current terminal output
func (tt *TerminalTest) GetOutput() string {
	return tt.output.String()
}

// GetVisibleOutput returns the output with ANSI escape sequences stripped
func (tt *TerminalTest) GetVisibleOutput() string {
	raw := tt.output.String()

	// Process carriage returns correctly - simulate overwriting
	processed := tt.processCarriageReturns(raw)

	// Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	clean := ansiRegex.ReplaceAllString(processed, "")

	return clean
}

// processCarriageReturns simulates how carriage returns work in real terminals
func (tt *TerminalTest) processCarriageReturns(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if strings.Contains(line, "\r") {
			// Handle sequences like "\r\033[K..." (clear line and rewrite)
			if strings.Contains(line, "\r\033[K") {
				parts := strings.Split(line, "\r\033[K")
				// Keep everything before the first \r\033[K, then the last part overwrites
				if len(parts) > 1 {
					result = append(result, parts[len(parts)-1])
				} else {
					result = append(result, line)
				}
			} else {
				// Regular \r without clear - just remove \r
				cleaned := strings.ReplaceAll(line, "\r", "")
				result = append(result, cleaned)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// ExpectOutput checks if the terminal contains expected text
func (tt *TerminalTest) ExpectOutput(expected string) bool {
	output := tt.GetVisibleOutput()
	contains := strings.Contains(output, expected)
	if !contains {
		tt.t.Errorf("Expected output to contain: %q\nActual output:\n%s", expected, output)
	}
	return contains
}

// ExpectPattern checks if the terminal output matches a regex pattern
func (tt *TerminalTest) ExpectPattern(pattern string) bool {
	output := tt.GetVisibleOutput()
	matched, err := regexp.MatchString(pattern, output)
	if err != nil {
		tt.t.Errorf("Invalid regex pattern: %s", pattern)
		return false
	}
	if !matched {
		tt.t.Errorf("Expected output to match pattern: %q\nActual output:\n%s", pattern, output)
	}
	return matched
}

// ExpectPrompt checks if the prompt symbol is visible
func (tt *TerminalTest) ExpectPrompt() bool {
	return tt.ExpectOutput("✦")
}

// ExpectWelcome checks if the welcome message is displayed
func (tt *TerminalTest) ExpectWelcome() bool {
	return tt.ExpectOutput("Rigel - AI Coding Agent")
}

// ExpectSpinner checks if spinner animation is present
func (tt *TerminalTest) ExpectSpinner() bool {
	// Look for common spinner characters
	output := tt.GetOutput() // Keep ANSI for spinner detection
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for _, char := range spinnerChars {
		if strings.Contains(output, char) {
			return true
		}
	}
	tt.t.Error("No spinner animation detected")
	return false
}

// ExpectThinking checks if "Thinking..." message is displayed
func (tt *TerminalTest) ExpectThinking() bool {
	return tt.ExpectOutput("Thinking...")
}

// ExpectNoCtrlC checks that ^C characters are not visible
func (tt *TerminalTest) ExpectNoCtrlC() bool {
	output := tt.GetVisibleOutput()
	if strings.Contains(output, "^C") {
		tt.t.Error("Found unwanted ^C characters in output")
		return false
	}
	return true
}

// GetLines returns the terminal output as separate lines
func (tt *TerminalTest) GetLines() []string {
	output := tt.GetVisibleOutput()
	return strings.Split(output, "\n")
}

// ExpectLineCount checks if output has expected number of non-empty lines
func (tt *TerminalTest) ExpectLineCount(count int) bool {
	lines := tt.GetLines()
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines != count {
		tt.t.Errorf("Expected %d non-empty lines, got %d", count, nonEmptyLines)
		return false
	}
	return true
}

// Close closes the terminal test session
func (tt *TerminalTest) Close() error {
	if tt.ptmx != nil {
		tt.ptmx.Close()
	}
	if tt.cmd != nil && tt.cmd.Process != nil {
		tt.cmd.Process.Kill()
		tt.cmd.Wait()
	}
	return nil
}

// GetPID returns the process ID
func (tt *TerminalTest) GetPID() int {
	if tt.cmd != nil && tt.cmd.Process != nil {
		return tt.cmd.Process.Pid
	}
	return -1
}

// GetProcessState returns the process state
func (tt *TerminalTest) GetProcessState() string {
	if tt.cmd == nil {
		return "no command"
	}
	if tt.cmd.Process == nil {
		return "no process"
	}
	if tt.cmd.ProcessState != nil {
		return tt.cmd.ProcessState.String()
	}
	return "running"
}

// Screenshot returns a formatted view of the current terminal state
func (tt *TerminalTest) Screenshot() string {
	lines := tt.GetLines()
	var result strings.Builder
	result.WriteString("=== Terminal Screenshot ===\n")
	for i, line := range lines {
		// Show ALL lines including empty ones to get accurate terminal representation
		result.WriteString(fmt.Sprintf("%2d: %s\n", i+1, line))
	}
	result.WriteString("========================\n")
	return result.String()
}

// ScreenshotNonEmpty returns a formatted view showing only non-empty lines (original behavior)
func (tt *TerminalTest) ScreenshotNonEmpty() string {
	lines := tt.GetLines()
	var result strings.Builder
	result.WriteString("=== Terminal Screenshot (Non-Empty) ===\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			result.WriteString(fmt.Sprintf("%2d: %s\n", i+1, line))
		}
	}
	result.WriteString("========================\n")
	return result.String()
}
