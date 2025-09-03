package termflow

import (
	"os"
	"testing"
	"time"

	"github.com/mizzy/rigel/lib/termflow/uitest"
)

// TestSimpleHistoryNavigation tests basic history navigation with /help command only
func TestSimpleHistoryNavigation(t *testing.T) {
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

	// Type /help command (fast command, doesn't require LLM)
	t.Log("Typing /help command")
	err = tt.Type("/help")
	if err != nil {
		t.Fatalf("Failed to type /help: %v", err)
	}

	// Wait for command to complete
	tt.Wait(1 * time.Second)

	// Verify we're back at prompt
	if !tt.ExpectPrompt() {
		t.Log("No prompt found, taking screenshot")
		t.Logf("State after /help:\n%s", tt.Screenshot())
	}

	// Now test up arrow to recall /help
	t.Log("Testing up arrow for history recall")
	err = tt.SendKeys("\033[A") // Up arrow
	if err != nil {
		t.Fatalf("Failed to send up arrow: %v", err)
	}

	tt.Wait(500 * time.Millisecond)
	t.Logf("After up arrow:\n%s", tt.Screenshot())

	// Check if /help appears in the line
	output := tt.GetVisibleOutput()
	t.Logf("Full output after up arrow: %q", output)

	if !tt.ExpectOutput("/help") {
		t.Error("Up arrow should recall /help command")
	}
}
