package uitest

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"
)

// TestBackspaceVisualErase verifies that pressing Backspace removes the character visually
func TestBackspaceVisualErase(t *testing.T) {
	// Run only in explicit UI test mode to avoid CI hiccups
	if os.Getenv("RIGEL_TEST_MODE") != "1" {
		t.Skip("set RIGEL_TEST_MODE=1 to run PTY backspace test")
	}
	// Path to the built rigel binary from this package directory
	// lib/termflow/uitest -> ../../bin/rigel
	rigelPath, _ := filepath.Abs(filepath.FromSlash("../../bin/rigel"))
	if rigelPath == "" {
		t.Fatal("could not resolve rigel binary path")
	}

	// Ensure the binary exists before starting the PTY session
	if _, err := os.Stat(rigelPath); err != nil {
		t.Skipf("rigel binary not found at %s; build first or set RIGEL_BINARY", rigelPath)
	}

	tt, err := NewTerminalTest(t, rigelPath, "--termflow")
	if err != nil {
		t.Fatalf("failed to start terminal session: %v", err)
	}
	defer tt.Close()

	// Allow the app to start and render prompt
	tt.Wait(400 * time.Millisecond)

	// Expect welcome then prompt
	if !tt.ExpectWelcome() {
		t.Fatalf("welcome not visible")
	}
	if !tt.ExpectPrompt() {
		t.Fatalf("prompt not visible")
	}

	// Type characters and then backspace one char
	if err := tt.SendKeys("abc"); err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}
	tt.Wait(120 * time.Millisecond)
	if err := tt.SendKeys("\x7f"); err != nil { // DEL as backspace
		t.Fatalf("failed to send backspace: %v", err)
	}
	tt.Wait(200 * time.Millisecond)

	// The last prompt line should end with "ab" (character erased)
	visible := tt.GetVisibleOutput()
	// Match a line ending with prompt followed by ab
	// The prompt symbol is
	re := regexp.MustCompile(`(?m)^✦\s+ab$`)
	if !re.MatchString(visible) {
		t.Fatalf("expected prompt line to end with 'ab' after backspace.\nOutput:\n%s", visible)
	}

	// Exit cleanly: send double Ctrl+C
	if err := tt.SendCtrlC(); err != nil {
		t.Logf("warn: failed to send first Ctrl+C: %v", err)
	}
	tt.Wait(100 * time.Millisecond)
	if err := tt.SendCtrlC(); err != nil {
		t.Logf("warn: failed to send second Ctrl+C: %v", err)
	}

	// Give it a moment to exit (especially on macOS CI)
	if runtime.GOOS == "darwin" {
		tt.Wait(300 * time.Millisecond)
	} else {
		tt.Wait(150 * time.Millisecond)
	}
}
