package uitest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestTermflowUI tests the basic termflow UI functionality
func TestTermflowUI(t *testing.T) {
	// Build the test binary
	testBinary := buildTestBinary(t)
	defer os.Remove(testBinary)

	// Start terminal test
	tt, err := NewTerminalTest(t, testBinary, "--termflow")
	if err != nil {
		t.Fatalf("Failed to start terminal test: %v", err)
	}
	defer tt.Close()

	// Wait for application to start
	tt.Wait(500 * time.Millisecond)

	t.Run("Welcome Message", func(t *testing.T) {
		tt.ExpectWelcome()
		tt.ExpectOutput("Using termflow UI")
		tt.ExpectPrompt()
	})

	t.Run("Basic Input", func(t *testing.T) {
		// Type a simple message
		err := tt.Type("hello")
		if err != nil {
			t.Fatalf("Failed to type input: %v", err)
		}

		tt.Wait(100 * time.Millisecond)

		// Should show thinking state
		tt.ExpectThinking()
		// Should show spinner
		tt.ExpectSpinner()

		// Wait for response
		tt.Wait(3 * time.Second)

		// Should return to prompt
		tt.ExpectPrompt()
	})

	t.Run("Ctrl+C Handling", func(t *testing.T) {
		// First Ctrl+C
		err := tt.SendCtrlC()
		if err != nil {
			t.Fatalf("Failed to send Ctrl+C: %v", err)
		}

		tt.Wait(100 * time.Millisecond)

		// Should show exit message
		tt.ExpectOutput("Press Ctrl+C again to exit")
		// Should not show ^C characters
		tt.ExpectNoCtrlC()
		// Should still have prompt
		tt.ExpectPrompt()

		// Second Ctrl+C
		err = tt.SendCtrlC()
		if err != nil {
			t.Fatalf("Failed to send second Ctrl+C: %v", err)
		}

		tt.Wait(100 * time.Millisecond)

		// Should show goodbye message
		tt.ExpectOutput("Goodbye!")
	})
}

// TestMultilineInput tests multiline input functionality
func TestMultilineInput(t *testing.T) {
	testBinary := buildTestBinary(t)
	defer os.Remove(testBinary)

	tt, err := NewTerminalTest(t, testBinary, "--termflow")
	if err != nil {
		t.Fatalf("Failed to start terminal test: %v", err)
	}
	defer tt.Close()

	tt.Wait(500 * time.Millisecond)

	t.Run("Multiline Trigger", func(t *testing.T) {
		// Type input ending with ...
		err := tt.SendKeys("Write a function...")
		if err != nil {
			t.Fatalf("Failed to type input: %v", err)
		}

		err = tt.SendEnter()
		if err != nil {
			t.Fatalf("Failed to send enter: %v", err)
		}

		tt.Wait(100 * time.Millisecond)

		// Should show multiline continuation prompt
		tt.ExpectOutput("Continue typing")
		tt.ExpectPattern(`\d+>`) // Should show line numbers like "2>"
	})
}

// TestCommandCompletion tests slash command functionality
func TestCommandCompletion(t *testing.T) {
	testBinary := buildTestBinary(t)
	defer os.Remove(testBinary)

	tt, err := NewTerminalTest(t, testBinary, "--termflow")
	if err != nil {
		t.Fatalf("Failed to start terminal test: %v", err)
	}
	defer tt.Close()

	tt.Wait(500 * time.Millisecond)

	t.Run("Help Command", func(t *testing.T) {
		err := tt.Type("/help")
		if err != nil {
			t.Fatalf("Failed to type command: %v", err)
		}

		tt.Wait(100 * time.Millisecond)

		// Should show help output
		tt.ExpectOutput("Available commands")
	})
}

// buildTestBinary builds a test version of the rigel binary
func buildTestBinary(t *testing.T) string {
	// Find the project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Go up directories to find cmd/rigel/main.go
	projectRoot := findProjectRoot(wd)
	if projectRoot == "" {
		t.Fatalf("Could not find project root")
	}

	mainPath := filepath.Join(projectRoot, "cmd", "rigel", "main.go")
	testBinary := filepath.Join(os.TempDir(), "rigel-test")

	// Build the binary
	if err := runCommand("go", "build", "-o", testBinary, mainPath); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	return testBinary
}

// findProjectRoot searches for the project root directory
func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// runCommand executes a shell command
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
