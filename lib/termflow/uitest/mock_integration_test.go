package uitest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mizzy/rigel/lib/termflow"
)

// MockIO provides controllable I/O for testing termflow components
type MockIO struct {
	input  *strings.Reader
	output *bytes.Buffer
}

// NewMockIO creates a new mock I/O for testing
func NewMockIO(input string) *MockIO {
	return &MockIO{
		input:  strings.NewReader(input),
		output: &bytes.Buffer{},
	}
}

// Read implements io.Reader
func (m *MockIO) Read(p []byte) (n int, err error) {
	return m.input.Read(p)
}

// Write implements io.Writer
func (m *MockIO) Write(p []byte) (n int, err error) {
	return m.output.Write(p)
}

// GetOutput returns the written output
func (m *MockIO) GetOutput() string {
	return m.output.String()
}

// MockClient creates a termflow client with mock I/O
func MockClient(input string) (*termflow.Client, *MockIO) {
	mockIO := NewMockIO(input)
	client := termflow.New()

	// Note: This would require making termflow.Client's I/O fields configurable
	// For now, this demonstrates the concept
	return client, mockIO
}

// TestTermflowComponents tests individual termflow components
func TestTermflowComponents(t *testing.T) {
	t.Run("Basic Output", func(t *testing.T) {
		client, mockIO := MockClient("hello\n")

		// Test PrintChat functionality
		client.PrintChat("test input", "test response")

		output := mockIO.GetOutput()
		if !strings.Contains(output, "✦") {
			t.Error("Expected prompt symbol in output")
		}
		if !strings.Contains(output, "test input") {
			t.Error("Expected user input in output")
		}
		if !strings.Contains(output, "test response") {
			t.Error("Expected AI response in output")
		}
	})

	t.Run("Colors and Formatting", func(t *testing.T) {
		client, mockIO := MockClient("")

		client.PrintResponse("Hello World")

		output := mockIO.GetOutput()
		// Check for ANSI color codes (color 252 for response text)
		if !strings.Contains(output, "\033[38;5;252m") {
			t.Error("Expected color formatting in response")
		}
		if !strings.Contains(output, "Hello World") {
			t.Error("Expected response text")
		}
		if !strings.Contains(output, "\033[0m") {
			t.Error("Expected color reset sequence")
		}
	})
}

// TestSpinnerComponents tests spinner functionality in isolation
func TestSpinnerComponents(t *testing.T) {
	t.Run("Spinner Animation", func(t *testing.T) {
		client, mockIO := MockClient("")

		// This would require making spinner testable
		// For demonstration purposes:
		_ = client
		_ = mockIO

		// In a real implementation, we'd test:
		// - Spinner starts with correct first frame
		// - Spinner cycles through frames
		// - Spinner stops and clears correctly
		// - ANSI escape sequences are correct
	})
}

// AssertContains checks if output contains expected string
func AssertContains(t *testing.T, output, expected string) {
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
	}
}

// AssertNotContains checks if output does not contain unwanted string
func AssertNotContains(t *testing.T, output, unwanted string) {
	if strings.Contains(output, unwanted) {
		t.Errorf("Expected output to NOT contain %q, got:\n%s", unwanted, output)
	}
}

// AssertColorCode checks if output contains specific ANSI color code
func AssertColorCode(t *testing.T, output, colorCode string) {
	expected := "\033[" + colorCode + "m"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected color code %q in output, got:\n%s", expected, output)
	}
}

// StripANSI removes ANSI escape sequences from text
func StripANSI(text string) string {
	// Simple ANSI stripping - could be more sophisticated
	result := text
	for strings.Contains(result, "\033[") {
		start := strings.Index(result, "\033[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}
