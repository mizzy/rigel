package agent

import (
	"context"
	"testing"
)

// MockLLMProvider for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Return simple mock response for testing
	return `[{"intent":"none","filepath":"","content":""}]`, nil
}

func TestNewPromptAnalyzer(t *testing.T) {
	mockProvider := &MockLLMProvider{}
	analyzer := NewPromptAnalyzer(mockProvider)
	
	if analyzer == nil {
		t.Error("NewPromptAnalyzer should not return nil")
	}
}

func TestIntentToString(t *testing.T) {
	testCases := []struct {
		intent   FileOperationIntent
		expected string
	}{
		{IntentRead, "read"},
		{IntentWrite, "write"},
		{IntentList, "list"},
		{IntentExists, "exists"},
		{IntentDelete, "delete"},
		{IntentNone, "none"},
	}

	for _, tc := range testCases {
		result := IntentToString(tc.intent)
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}