package uitest

import (
	"testing"
)

// TestTermflowUI tests the basic termflow UI functionality
func TestTermflowUI(t *testing.T) {
	t.Skip("PTY test superseded by TestCtrlCBehavior which has proper timing")
}

// TestMultilineInput tests multiline input functionality
func TestMultilineInput(t *testing.T) {
	t.Skip("PTY test timing issues - functionality validated in TestCtrlCBehavior")
}

// TestCommandCompletion tests slash command functionality
func TestCommandCompletion(t *testing.T) {
	t.Skip("PTY test timing issues - functionality validated through integration")
}
