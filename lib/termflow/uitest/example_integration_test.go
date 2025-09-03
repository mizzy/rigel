package uitest

import (
	"testing"
)

// Example of how to use the terminal testing framework
func TestExample(t *testing.T) {
	t.Skip("PTY test requires longer startup times than practical for CI - use TestCtrlCBehavior as main PTY validation")
}

// Benchmark test to measure performance
func BenchmarkTerminalOutput(b *testing.B) {
	b.Skip("PTY benchmark requires stable timing environment")
}
