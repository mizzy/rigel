package termflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizzy/rigel/lib/termflow/uitest"
)

// buildTestBinary builds the test binary if it doesn't exist
func buildTestBinary() error {
	// Check if binary already exists
	if _, err := os.Stat("/tmp/rigel-test"); err == nil {
		return nil
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "/tmp/rigel-test", "cmd/rigel/main.go")
	cmd.Env = append(os.Environ(), "PROVIDER=ollama")
	return cmd.Run()
}

// TestHistoryPersistenceAcrossSessions tests that commands from one session are available in the next
func TestHistoryPersistenceAcrossSessions(t *testing.T) {
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

	// Create a temporary directory for this test's history
	tmpDir, err := os.MkdirTemp("", "rigel_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set HOME to our temp directory so history goes there
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("First session - add commands to history", func(t *testing.T) {
		// Build test binary for CI
		if err := buildTestBinary(); err != nil {
			t.Skipf("Failed to build test binary: %v", err)
		}

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

		// Add a unique command to history
		testCommand := "test history persistence"
		t.Logf("Typing command: %s", testCommand)
		err = tt.Type(testCommand)
		if err != nil {
			t.Fatalf("Failed to type command: %v", err)
		}

		// Wait for command to be processed and return to prompt
		tt.Wait(5 * time.Second)

		// Add another command
		testCommand2 := "/status"
		t.Logf("Typing command: %s", testCommand2)
		err = tt.Type(testCommand2)
		if err != nil {
			t.Fatalf("Failed to type command: %v", err)
		}

		// Wait for status command to complete
		tt.Wait(2 * time.Second)

		// Verify we have a prompt
		if !tt.ExpectPrompt() {
			t.Logf("No prompt found after commands, taking screenshot")
			t.Logf("Current state:\n%s", tt.Screenshot())
		}

		// Send Ctrl+C twice to exit cleanly
		tt.SendCtrlC()
		tt.Wait(100 * time.Millisecond)
		tt.SendCtrlC()
		tt.Wait(500 * time.Millisecond)

		t.Log("First session completed")
	})

	// Verify history file was created
	historyPath := filepath.Join(tmpDir, ".rigel", "history")
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}
	t.Logf("History file exists at: %s", historyPath)

	// Read and log history file content for debugging
	historyContent, err := os.ReadFile(historyPath)
	if err != nil {
		t.Logf("Could not read history file: %v", err)
	} else {
		t.Logf("History file content:\n%s", string(historyContent))
	}

	t.Run("Second session - verify history is available", func(t *testing.T) {
		// Ensure test binary exists
		if err := buildTestBinary(); err != nil {
			t.Skipf("Failed to build test binary: %v", err)
		}

		tt, err := uitest.NewTerminalTest(t, "/tmp/rigel-test", "--termflow")
		if err != nil {
			t.Skip("Test binary not available")
		}
		defer tt.Close()

		// Wait for startup
		tt.Wait(15 * time.Second)

		// Verify initial state
		if !tt.ExpectWelcome() {
			t.Fatal("Welcome message not found in second session")
		}
		if !tt.ExpectPrompt() {
			t.Fatal("Initial prompt not found in second session")
		}

		// Test up arrow to get most recent command from previous session
		t.Log("Testing up arrow for history from previous session")
		err = tt.SendKeys("\033[A") // Up arrow
		if err != nil {
			t.Fatalf("Failed to send up arrow: %v", err)
		}

		tt.Wait(500 * time.Millisecond)
		t.Logf("After up arrow:\n%s", tt.Screenshot())

		// Should show the most recent command (/status) from previous session
		if !tt.ExpectOutput("/status") {
			t.Error("Up arrow should show most recent command '/status' from previous session")
			t.Logf("Current output: %q", tt.GetVisibleOutput())
		}

		// Test another up arrow to get the earlier command
		t.Log("Testing second up arrow for earlier command")
		err = tt.SendKeys("\033[A") // Up arrow again
		if err != nil {
			t.Fatalf("Failed to send second up arrow: %v", err)
		}

		tt.Wait(500 * time.Millisecond)
		t.Logf("After second up arrow:\n%s", tt.Screenshot())

		// Should show "test history persistence" from previous session
		if !tt.ExpectOutput("test history persistence") {
			t.Error("Second up arrow should show 'test history persistence' from previous session")
			t.Logf("Current output: %q", tt.GetVisibleOutput())
		}

		// Send Ctrl+C twice to exit cleanly
		tt.SendCtrlC()
		tt.Wait(100 * time.Millisecond)
		tt.SendCtrlC()
		tt.Wait(500 * time.Millisecond)

		t.Log("Second session completed - history persistence verified!")
	})
}
