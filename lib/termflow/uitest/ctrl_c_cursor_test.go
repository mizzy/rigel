package uitest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestCtrlCCursorPosition verifies the cursor remains at the exact
// input position after pressing Ctrl+C during multiline editing.
func TestCtrlCCursorPosition(t *testing.T) {
	// Skip unless explicitly enabled (pre-commit and CI should not run PTY by default)
	if os.Getenv("RIGEL_TEST_MODE") != "1" {
		t.Skip("set RIGEL_TEST_MODE=1 to run PTY cursor tests")
	}

	// Allow skipping in environments without the built binary
	bin := os.Getenv("RIGEL_BINARY")
	if bin == "" {
		// Resolve path relative to this test file to avoid cwd issues
		_, thisFile, _, _ := runtime.Caller(0)
		base := filepath.Dir(thisFile)
		repoRoot := filepath.Clean(filepath.Join(base, "..", "..", ".."))
		bin = filepath.Join(repoRoot, "bin", "rigel")
	}
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("rigel binary not found at %s; set RIGEL_BINARY or build first", bin)
	}

	tt, err := NewTerminalTest(t, bin, "--termflow")
	if err != nil {
		t.Fatalf("failed to start rigel: %v", err)
	}
	defer tt.Close()

	// Wait for welcome/prompt
	tt.Wait(700 * time.Millisecond)
	tt.ExpectWelcome()
	tt.ExpectPrompt()

	// Type multiline: "aaaa" then newline, then "bbbb"
	if err := tt.SendKeys("aaaa"); err != nil {
		t.Fatalf("send aaaa: %v", err)
	}
	if err := tt.SendCtrlJ(); err != nil {
		t.Fatalf("ctrl-j: %v", err)
	}
	if err := tt.SendKeys("bbbb"); err != nil {
		t.Fatalf("send bbbb: %v", err)
	}

	// Press Ctrl+C to show exit hint and restore cursor to input line
	if err := tt.SendCtrlC(); err != nil {
		t.Fatalf("ctrl-c: %v", err)
	}
	tt.Wait(120 * time.Millisecond)

	// Type a marker character; it must appear immediately after "bbbb"
	if err := tt.SendKeys("X"); err != nil {
		t.Fatalf("send X: %v", err)
	}
	tt.Wait(120 * time.Millisecond)

	// Expect the continuation line to show exactly two spaces + bbbbX (no prompt offset)
	if ok := tt.ExpectPattern(`(?m)^\s{2}bbbbX\s*$`); !ok {
		t.Logf("Full output for debugging:\n%s", tt.Screenshot())
	}
}
