package termflow

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mizzy/rigel/lib/termflow/uitest"
)

// TestHistoryNavigation tests cursor key navigation through input history
func TestHistoryNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping terminal integration test in short mode")
	}

	// Skip if no terminal is available (CI environment)
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping terminal integration test in CI environment")
	}

	// Set test mode and provider environment variables
	oldTestEnv := os.Getenv("RIGEL_TEST_MODE")
	oldProvider := os.Getenv("PROVIDER")
	os.Setenv("RIGEL_TEST_MODE", "1")
	os.Setenv("PROVIDER", "ollama")
	defer func() {
		os.Setenv("RIGEL_TEST_MODE", oldTestEnv)
		os.Setenv("PROVIDER", oldProvider)
	}()

	tt, err := uitest.NewTerminalTest(t, "/tmp/rigel-test", "--termflow")
	if err != nil {
		t.Skip("Test binary not available, run: go build -o /tmp/rigel-test cmd/rigel/main.go")
	}
	defer tt.Close()

	// Wait for startup
	tt.Wait(15 * time.Second)

	// Verify initial state
	if !tt.ExpectWelcome() {
		t.Fatal("Welcome message not found")
	}
	if !tt.ExpectPrompt() {
		t.Fatal("Initial prompt not found")
	}

	t.Run("Build command history", func(t *testing.T) {
		// Add some commands to history
		commands := []string{"hello world", "test command", "/status"}

		for i, cmd := range commands {
			t.Logf("Entering command %d: %s", i+1, cmd)

			// Type the command and press Enter
			err := tt.Type(cmd)
			if err != nil {
				t.Fatalf("Failed to type command %q: %v", cmd, err)
			}

			// Wait for command to be processed
			tt.Wait(500 * time.Millisecond)

			// Wait for prompt to return (indicating command completed)
			// For /status this should be quick, for chat commands it takes longer
			if strings.HasPrefix(cmd, "/") {
				tt.Wait(1 * time.Second)
			} else {
				// Chat commands need more time
				tt.Wait(5 * time.Second)
			}

			// Verify prompt returned
			if !tt.ExpectPrompt() {
				t.Logf("Prompt not found after command %q, taking screenshot", cmd)
				t.Logf("Current state:\n%s", tt.Screenshot())
			}
		}

		t.Logf("Command history built. Current state:\n%s", tt.Screenshot())
	})

	t.Run("Navigate history with arrow keys", func(t *testing.T) {
		// Test up arrow to get most recent command
		t.Log("Testing up arrow - should show most recent command")
		err := tt.SendKeys("\033[A") // Up arrow
		if err != nil {
			t.Fatalf("Failed to send up arrow: %v", err)
		}

		tt.Wait(200 * time.Millisecond)
		t.Logf("After up arrow:\n%s", tt.Screenshot())

		// Should show the most recent command (/status)
		if !tt.ExpectOutput("/status") {
			t.Error("Up arrow should show most recent command '/status'")
			t.Logf("Current output: %q", tt.GetVisibleOutput())
		}

		// Test another up arrow to go to previous command
		t.Log("Testing second up arrow - should show previous command")
		err = tt.SendKeys("\033[A") // Up arrow again
		if err != nil {
			t.Fatalf("Failed to send second up arrow: %v", err)
		}

		tt.Wait(200 * time.Millisecond)
		t.Logf("After second up arrow:\n%s", tt.Screenshot())

		// Should show "test command"
		if !tt.ExpectOutput("test command") {
			t.Error("Second up arrow should show 'test command'")
			t.Logf("Current output: %q", tt.GetVisibleOutput())
		}

		// Test down arrow to go forward in history
		t.Log("Testing down arrow - should go forward in history")
		err = tt.SendKeys("\033[B") // Down arrow
		if err != nil {
			t.Fatalf("Failed to send down arrow: %v", err)
		}

		tt.Wait(200 * time.Millisecond)
		t.Logf("After down arrow:\n%s", tt.Screenshot())

		// Should show "/status" again
		if !tt.ExpectOutput("/status") {
			t.Error("Down arrow should show '/status' again")
			t.Logf("Current output: %q", tt.GetVisibleOutput())
		}

		// Test down arrow to go to empty line
		t.Log("Testing final down arrow - should clear to empty line")
		err = tt.SendKeys("\033[B") // Down arrow
		if err != nil {
			t.Fatalf("Failed to send final down arrow: %v", err)
		}

		tt.Wait(200 * time.Millisecond)
		t.Logf("After final down arrow:\n%s", tt.Screenshot())

		// Should be back to empty prompt
		lines := tt.GetLines()
		lastLine := ""
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				lastLine = strings.TrimSpace(lines[i])
				break
			}
		}

		if !strings.HasSuffix(lastLine, "✦") || strings.Contains(lastLine, "/status") {
			t.Errorf("Final down arrow should clear to empty prompt, got: %q", lastLine)
		}
	})
}

// TestHistoryPersistence tests that history persists across sessions
func TestHistoryPersistence(t *testing.T) {
	t.Skip("History persistence test - requires file system testing setup")
}
