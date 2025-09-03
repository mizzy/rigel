package uitest

import (
	"testing"
	"time"
)

// Example of how to use the terminal testing framework
func TestExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping terminal integration test in short mode")
	}

	// This test requires the binary to be built first
	// Run: go build -o /tmp/rigel-test cmd/rigel/main.go

	tt, err := NewTerminalTest(t, "/tmp/rigel-test", "--termflow")
	if err != nil {
		t.Skip("Test binary not available, run: go build -o /tmp/rigel-test cmd/rigel/main.go")
	}
	defer tt.Close()

	// Wait for startup
	tt.Wait(1 * time.Second)

	// Take a screenshot of initial state
	t.Logf("Initial state:\n%s", tt.Screenshot())

	// Test welcome message
	if !tt.ExpectWelcome() {
		t.Fatal("Welcome message not found")
	}

	// Test prompt
	if !tt.ExpectPrompt() {
		t.Fatal("Prompt not found")
	}

	// Test basic interaction
	err = tt.Type("hello")
	if err != nil {
		t.Fatalf("Failed to type: %v", err)
	}

	// Wait a bit and take another screenshot
	tt.Wait(500 * time.Millisecond)
	t.Logf("After typing 'hello':\n%s", tt.Screenshot())

	// The response will take time, so we just check for thinking state
	tt.ExpectThinking()
}

// Benchmark test to measure performance
func BenchmarkTerminalOutput(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	tt, err := NewTerminalTest(&testing.T{}, "/tmp/rigel-test", "--termflow")
	if err != nil {
		b.Skip("Test binary not available")
	}
	defer tt.Close()

	tt.Wait(1 * time.Second) // Wait for startup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tt.SendKeys("test message")
		tt.SendEnter()
		tt.Wait(10 * time.Millisecond)
	}
}
