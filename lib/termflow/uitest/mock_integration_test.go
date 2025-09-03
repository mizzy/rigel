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
	t.Skip("Mock testing requires refactoring termflow Client I/O architecture")
}

// TestSpinnerComponents tests spinner functionality in isolation
func TestSpinnerComponents(t *testing.T) {
	t.Skip("Spinner testing requires refactoring spinner I/O architecture")
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
