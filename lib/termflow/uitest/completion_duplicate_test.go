package uitest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// countInTail counts occurrences of the needle within the last N visible lines
func countInTail(tt *TerminalTest, needle string, tailLines int) int {
	lines := tt.GetLines()
	if tailLines > len(lines) {
		tailLines = len(lines)
	}
	tail := lines[len(lines)-tailLines:]
	count := 0
	for _, ln := range tail {
		if strings.Contains(ln, needle) {
			count++
		}
	}
	return count
}

// TestCompletion_NoDuplicateBlocks_OnNavigation ensures the completion header doesn't duplicate
func TestCompletion_NoDuplicateBlocks_OnNavigation(t *testing.T) {
	if os.Getenv("RIGEL_TEST_MODE") != "1" {
		t.Skip("set RIGEL_TEST_MODE=1 to run PTY completion duplication test")
	}

	rigelPath, _ := filepath.Abs(filepath.FromSlash("../../bin/rigel"))
	if _, err := os.Stat(rigelPath); err != nil {
		t.Skipf("rigel binary not found at %s; build first", rigelPath)
	}

	tt, err := NewTerminalTest(t, rigelPath)
	if err != nil {
		t.Fatalf("failed to start terminal session: %v", err)
	}
	defer tt.Close()

	tt.Wait(400 * time.Millisecond)
	if !tt.ExpectWelcome() {
		t.Fatalf("welcome not visible")
	}

	// Trigger completion menu
	if err := tt.SendKeys("/"); err != nil {
		t.Fatalf("send '/' failed: %v", err)
	}
	tt.Wait(120 * time.Millisecond)
	if err := tt.SendKeys("\t"); err != nil { // Tab
		t.Fatalf("send Tab failed: %v", err)
	}
	tt.Wait(200 * time.Millisecond)

	// Navigate a bit: Down, Down, Up, Down
	seq := []string{"\x1b[B", "\x1b[B", "\x1b[A", "\x1b[B"}
	for _, s := range seq {
		if err := tt.SendKeys(s); err != nil {
			t.Fatalf("send arrow failed: %v", err)
		}
		tt.Wait(120 * time.Millisecond)
	}

	// In the last 60 lines, we should have exactly one Completions: header
	headers := countInTail(tt, "Completions:", 60)
	if headers != 1 {
		t.Fatalf("expected exactly 1 completion header in tail, got %d\n%s", headers, tt.Screenshot())
	}
}

// TestCompletion_MenuClosesOnApply ensures the menu is cleared after apply
func TestCompletion_MenuClosesOnApply(t *testing.T) {
	if os.Getenv("RIGEL_TEST_MODE") != "1" {
		t.Skip("set RIGEL_TEST_MODE=1 to run PTY completion apply test")
	}

	rigelPath, _ := filepath.Abs(filepath.FromSlash("../../bin/rigel"))
	if _, err := os.Stat(rigelPath); err != nil {
		t.Skipf("rigel binary not found at %s; build first", rigelPath)
	}

	tt, err := NewTerminalTest(t, rigelPath)
	if err != nil {
		t.Fatalf("failed to start terminal session: %v", err)
	}
	defer tt.Close()

	tt.Wait(400 * time.Millisecond)
	if !tt.ExpectWelcome() {
		t.Fatalf("welcome not visible")
	}

	// Show menu
	if err := tt.SendKeys("/"); err != nil {
		t.Fatalf("send '/' failed: %v", err)
	}
	tt.Wait(120 * time.Millisecond)
	if err := tt.SendKeys("\t"); err != nil {
		t.Fatalf("send Tab failed: %v", err)
	}
	tt.Wait(200 * time.Millisecond)

	// Apply with Tab
	if err := tt.SendKeys("\t"); err != nil {
		t.Fatalf("send Tab apply failed: %v", err)
	}
	tt.Wait(200 * time.Millisecond)

	// After apply, in the last 40 lines there should be 0 headers
	headers := countInTail(tt, "Completions:", 40)
	if headers != 0 {
		t.Fatalf("expected 0 completion headers after apply, got %d\n%s", headers, tt.Screenshot())
	}
}
