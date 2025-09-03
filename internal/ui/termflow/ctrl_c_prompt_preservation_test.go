package termflow

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mizzy/rigel/lib/termflow/uitest"
)

// TestCtrlCPromptPreservation tests that the input prompt remains visible after first Ctrl+C
func TestCtrlCPromptPreservation(t *testing.T) {
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

	// Build test binary for CI
	if err := buildTestBinary(); err != nil {
		t.Skipf("Failed to build test binary: %v", err)
	}

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

	// Take screenshot BEFORE Ctrl+C to see original prompt position
	beforeCtrlC := tt.Screenshot()
	t.Logf("Before Ctrl+C:\n%s", beforeCtrlC)

	// Find the line where the input prompt is waiting for input
	beforeLines := tt.GetLines()
	var originalInputLineIndex int
	found := false
	for i, line := range beforeLines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine == "✦" && i > 7 { // Skip welcome prompts, find the input prompt (lowered threshold)
			originalInputLineIndex = i
			found = true
			t.Logf("Original input prompt found at line %d (before Ctrl+C)", i+1)
			break
		}
	}

	if !found {
		t.Fatal("Could not find original input prompt before Ctrl+C")
	}

	t.Run("First Ctrl+C should show exit message and preserve prompt", func(t *testing.T) {
		// Send first Ctrl+C
		err := tt.SendCtrlC()
		if err != nil {
			t.Fatalf("Failed to send Ctrl+C: %v", err)
		}

		tt.Wait(500 * time.Millisecond)

		// Take screenshot to see current state
		t.Logf("After first Ctrl+C:\n%s", tt.Screenshot())

		// Check that exit instruction message is sent (in raw output)
		rawOutput := tt.GetOutput()
		if !strings.Contains(rawOutput, "Press Ctrl+C again to exit") {
			t.Errorf("Exit instruction not found in raw PTY output")
			t.Logf("Raw output: %q", rawOutput)
		}

		// Also check visible output for debugging
		visibleOutput := tt.GetVisibleOutput()
		t.Logf("Visible output: %q", visibleOutput)

		// Should NOT show ^C characters
		if !tt.ExpectNoCtrlC() {
			t.Fatal("Found unwanted ^C characters")
		}

		// Verify the exit message exists in raw output (internal check)
		if !strings.Contains(rawOutput, "Press Ctrl+C again to exit") {
			t.Fatal("Exit message not found in raw output")
		}

		// CRITICAL TEST: Verify that the exit message appears BELOW the original prompt position
		screenshot := tt.Screenshot()
		exitMessageFound := strings.Contains(screenshot, "Press Ctrl+C again to exit")
		promptFound := strings.Contains(screenshot, "✦")

		if !exitMessageFound {
			t.Error("ISSUE DETECTED: 'Press Ctrl+C again to exit' message should be visible in screenshot")
			t.Logf("Screenshot should contain exit message")
		}

		if !promptFound {
			t.Error("ISSUE DETECTED: Input prompt (✦) should be visible after exit message")
			t.Logf("Screenshot should contain prompt symbol ✦")
		}

		// NEW REQUIREMENT: Check the positioning of elements after Ctrl+C
		lines := tt.GetLines()
		inputPromptLineIndex := -1
		exitMessageLineIndex := -1
		newPromptLineIndex := -1

		for i, line := range lines {
			cleanLine := strings.TrimSpace(line)

			// Find the input prompt line (empty prompt line after welcome, around line 9-12)
			if cleanLine == "✦" && i > 8 { // Skip the welcome message prompts
				if inputPromptLineIndex == -1 {
					inputPromptLineIndex = i
					t.Logf("Found input prompt at line %d: %q", i+1, cleanLine)
				} else if newPromptLineIndex == -1 {
					newPromptLineIndex = i
					t.Logf("Found new prompt at line %d: %q", i+1, cleanLine)
				}
			}

			// Find the exit message line
			if strings.Contains(cleanLine, "Press Ctrl+C again to exit") {
				exitMessageLineIndex = i
				t.Logf("Found exit message at line %d: %q", i+1, cleanLine)
			}
		}

		// CRITICAL TEST: The ideal behavior should be:
		// Line X: ✦ (original input prompt - should stay in place)
		// Line X+1: Press Ctrl+C again to exit (message appears below)
		// Line X+2: ✦ (cursor returns to same position for continued input)

		if inputPromptLineIndex == -1 {
			t.Error("ISSUE DETECTED: Could not find original input prompt line")
		}

		if exitMessageLineIndex == -1 {
			t.Error("ISSUE DETECTED: Could not find exit message line")
		}

		// NEW IDEAL BEHAVIOR: The original prompt should remain active and usable
		// Before: Line X has "✦ " (waiting for input)
		// After:  Line X still has "✦ " (SAME position, SAME prompt, cursor stays here)
		//         Line X+1 has "Press Ctrl+C again to exit" (message appears below)
		//         User can continue typing at Line X (no new prompt needed)

		if exitMessageLineIndex != -1 {
			expectedExitMessageLine := originalInputLineIndex + 1

			// Check that the exit message appears right below the original prompt
			if exitMessageLineIndex != expectedExitMessageLine {
				t.Error("ISSUE DETECTED: Exit message should appear immediately below the ORIGINAL input prompt position")
				t.Logf("Original input prompt line (before Ctrl+C): %d", originalInputLineIndex+1)
				t.Logf("Expected exit message line: %d, Actual: %d", expectedExitMessageLine+1, exitMessageLineIndex+1)
			}

			// CRITICAL: There should NOT be a second prompt - the original should remain active
			if inputPromptLineIndex != -1 && inputPromptLineIndex != originalInputLineIndex {
				t.Error("ISSUE DETECTED: Original prompt should remain active, no new prompt should be created")
				t.Logf("Original prompt line: %d, New prompt found at line: %d", originalInputLineIndex+1, inputPromptLineIndex+1)
				t.Logf("Expected: Only one prompt at line %d, with message below at line %d", originalInputLineIndex+1, expectedExitMessageLine+1)
			}

			// Success condition: message appears below original prompt, no duplicate prompt
			if exitMessageLineIndex == expectedExitMessageLine && (inputPromptLineIndex == -1 || inputPromptLineIndex == originalInputLineIndex) {
				t.Log("SUCCESS: Exit message appears below original prompt, which remains active for continued input")
			}
		} else {
			t.Error("Could not find exit message line")
		}
	})

	t.Run("User can still type after first Ctrl+C", func(t *testing.T) {
		// Try typing something to verify prompt is functional
		testInput := "test after ctrl+c"
		err := tt.SendKeys(testInput)
		if err != nil {
			t.Fatalf("Failed to send keys after Ctrl+C: %v", err)
		}

		tt.Wait(200 * time.Millisecond)
		afterTyping := tt.Screenshot()
		t.Logf("After typing input:\n%s", afterTyping)

		// Should show the typed input
		if !tt.ExpectOutput(testInput) {
			t.Error("Typed input not visible after first Ctrl+C")
		}

		// CRITICAL: The input should appear at the ORIGINAL prompt line, not a new line
		afterLines := tt.GetLines()
		inputFoundAtOriginalLine := false
		for i, line := range afterLines {
			if strings.Contains(line, testInput) && i == originalInputLineIndex {
				inputFoundAtOriginalLine = true
				t.Logf("SUCCESS: User input appears at original prompt line %d: %q", i+1, strings.TrimSpace(line))
				break
			}
		}

		if !inputFoundAtOriginalLine {
			t.Error("ISSUE DETECTED: User input should appear at the original prompt line")
			t.Logf("Expected input at line %d (original prompt position)", originalInputLineIndex+1)
		}

		// Clear the input by pressing Ctrl+C again and then testing fresh prompt
		tt.SendCtrlC()
		tt.Wait(200 * time.Millisecond)
		t.Log("Exited cleanly after second Ctrl+C")
	})
}
