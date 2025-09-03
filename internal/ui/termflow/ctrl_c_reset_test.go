package termflow

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mizzy/rigel/lib/termflow/uitest"
)

func TestCtrlCResetAfterOneSecond(t *testing.T) {
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

	// Wait for startup - rigel takes ~12 seconds to start
	tt.Wait(15 * time.Second)

	// Verify initial state
	if !tt.ExpectWelcome() {
		t.Fatal("Welcome message not found")
	}
	if !tt.ExpectPrompt() {
		t.Fatal("Initial prompt not found")
	}

	t.Run("Ctrl+C should show message then reset after 1 second", func(t *testing.T) {
		// Send first Ctrl+C
		err := tt.SendCtrlC()
		if err != nil {
			t.Fatalf("Failed to send Ctrl+C: %v", err)
		}

		tt.Wait(200 * time.Millisecond)

		// Take screenshot to see current state
		t.Logf("After first Ctrl+C:\n%s", tt.Screenshot())

		// Should show the exit instruction
		if !tt.ExpectOutput("Press Ctrl+C again to exit") {
			t.Fatal("Exit instruction not found")
		}

		// Should NOT show ^C characters
		if !tt.ExpectNoCtrlC() {
			t.Fatal("Found unwanted ^C characters")
		}

		// Wait for 1 second + a small buffer
		tt.Wait(1200 * time.Millisecond)

		// Take screenshot to see state after 1 second
		t.Logf("After 1 second wait:\n%s", tt.Screenshot())

		// The exit message should be cleared (no longer visible)
		lines := tt.GetLines()
		for i, line := range lines {
			cleanLine := strings.TrimSpace(line)
			if strings.Contains(cleanLine, "Press Ctrl+C again to exit") {
				t.Errorf("Exit message should be cleared after 1 second but found on line %d: %q", i+1, cleanLine)
			}
		}

		// Now try pressing Ctrl+C again - it should NOT exit (should reset to first press behavior)
		err = tt.SendCtrlC()
		if err != nil {
			t.Fatalf("Failed to send second Ctrl+C: %v", err)
		}

		tt.Wait(200 * time.Millisecond)

		// Should show exit message again (first press behavior, not exit)
		if !tt.ExpectOutput("Press Ctrl+C again to exit") {
			t.Fatal("Exit instruction should appear again after reset")
		}

		// Should NOT show goodbye message (would indicate it exited)
		screenshot := tt.Screenshot()
		if strings.Contains(screenshot, "Goodbye!") {
			t.Fatal("Should NOT exit - Ctrl+C state should have reset after 1 second")
		}

		// Now immediately send another Ctrl+C (within the 1 second window) - this should exit
		err = tt.SendCtrlC()
		if err != nil {
			t.Fatalf("Failed to send final Ctrl+C: %v", err)
		}

		tt.Wait(200 * time.Millisecond)

		// Should show goodbye message (exit)
		if !tt.ExpectOutput("Goodbye!") {
			t.Fatal("Should exit on second Ctrl+C within the window")
		}
	})
}
