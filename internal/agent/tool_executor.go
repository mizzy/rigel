package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mizzy/rigel/internal/tools"
)

// ToolExecutionResult represents the result of tool execution
type ToolExecutionResult struct {
	Tool      string
	Input     string
	Output    string
	Error     error
	Duration  time.Duration
	StartTime time.Time
}

// ProgressDisplay interface for showing tool execution progress
type ProgressDisplay interface {
	ShowProgress(toolName, operation string)
	ShowResult(result ToolExecutionResult)
}

// ConsoleProgressDisplay implements ProgressDisplay for console output
type ConsoleProgressDisplay struct{}

func (c *ConsoleProgressDisplay) ShowProgress(toolName, operation string) {
	fmt.Printf("🔧 Executing %s: %s...\n", toolName, operation)
}

func (c *ConsoleProgressDisplay) ShowResult(result ToolExecutionResult) {
	duration := result.Duration.Round(time.Millisecond)
	if result.Error != nil {
		fmt.Printf("❌ %s failed (%v): %v\n", result.Tool, duration, result.Error)
	} else {
		fmt.Printf("✅ %s completed (%v)\n", result.Tool, duration)
	}
}

// UIProgressDisplay implements ProgressDisplay that returns formatted strings for UI integration
type UIProgressDisplay struct {
	progressMessages []string
	resultMessages   []string
}

func NewUIProgressDisplay() *UIProgressDisplay {
	return &UIProgressDisplay{
		progressMessages: make([]string, 0),
		resultMessages:   make([]string, 0),
	}
}

// Initialize creates and initializes a UIProgressDisplay
func (u *UIProgressDisplay) Initialize() {
	if u.progressMessages == nil {
		u.progressMessages = make([]string, 0)
	}
	if u.resultMessages == nil {
		u.resultMessages = make([]string, 0)
	}
}

func (u *UIProgressDisplay) ShowProgress(toolName, operation string) {
	msg := fmt.Sprintf("🔧 Executing %s: %s...", toolName, operation)
	u.progressMessages = append(u.progressMessages, msg)
}

func (u *UIProgressDisplay) ShowResult(result ToolExecutionResult) {
	duration := result.Duration.Round(time.Millisecond)
	var msg string
	if result.Error != nil {
		msg = fmt.Sprintf("❌ %s failed (%v): %v", result.Tool, duration, result.Error)
	} else {
		msg = fmt.Sprintf("✅ %s completed (%v)", result.Tool, duration)
	}
	u.resultMessages = append(u.resultMessages, msg)
}

func (u *UIProgressDisplay) GetProgressMessages() []string {
	return u.progressMessages
}

func (u *UIProgressDisplay) GetResultMessages() []string {
	return u.resultMessages
}

func (u *UIProgressDisplay) GetAllMessages() []string {
	var all []string
	all = append(all, u.progressMessages...)
	all = append(all, u.resultMessages...)
	return all
}

// FileOperationIntent represents different types of file operations
type FileOperationIntent int

const (
	IntentRead FileOperationIntent = iota
	IntentWrite
	IntentList
	IntentExists
	IntentDelete
	IntentNone
)

// PromptAnalyzer analyzes user prompts using LLM to determine if file operations are needed
type PromptAnalyzer struct {
	llmProvider interface {
		Generate(ctx context.Context, prompt string) (string, error)
	}
}

// NewPromptAnalyzer creates a new LLM-based prompt analyzer
func NewPromptAnalyzer(llmProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}) *PromptAnalyzer {
	return &PromptAnalyzer{
		llmProvider: llmProvider,
	}
}

// AnalyzePrompt analyzes a user prompt using LLM and returns file operation intents
func (pa *PromptAnalyzer) AnalyzePrompt(prompt string) []FileOperationMatch {
	ctx := context.Background()

	systemPrompt := `You are a file operation intent analyzer. Analyze the given user prompt and determine if it contains any file operation requests.

Respond with a JSON array containing file operations. Each operation should have:
- "intent": one of "read", "write", "list", "exists", "delete", "none"
- "filepath": the target file path (use "sample.txt" if not specified for write operations)
- "content": the content to write (only for write operations). Use "<GENERATE_TEXT>" when the user asks to generate content like "sample text", "dummy content", "適当な文章", "some text", etc.

Examples:
User: "read config.json"
Response: [{"intent":"read","filepath":"config.json","content":""}]

User: "create a file with hello world"
Response: [{"intent":"write","filepath":"sample.txt","content":"hello world"}]

User: "適当な文章をファイルに書き出して"
Response: [{"intent":"write","filepath":"sample.txt","content":"<GENERATE_TEXT>"}]

User: "list files"
Response: [{"intent":"list","filepath":".","content":""}]

User: "how are you today?"
Response: [{"intent":"none","filepath":"","content":""}]

Only respond with the JSON array, nothing else.`

	fullPrompt := fmt.Sprintf("%s\n\nUser: %s\nResponse:", systemPrompt, prompt)

	response, err := pa.llmProvider.Generate(ctx, fullPrompt)
	if err != nil {
		// Fallback to no matches if LLM fails
		return []FileOperationMatch{}
	}

	// Parse JSON response
	var rawMatches []struct {
		Intent   string `json:"intent"`
		FilePath string `json:"filepath"`
		Content  string `json:"content"`
	}

	// Clean response (remove markdown code blocks if present)
	cleanResponse := strings.TrimSpace(response)
	if strings.HasPrefix(cleanResponse, "```json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)
	}

	if err := json.Unmarshal([]byte(cleanResponse), &rawMatches); err != nil {
		// If JSON parsing fails, return no matches
		return []FileOperationMatch{}
	}

	// Convert to FileOperationMatch
	var matches []FileOperationMatch
	for _, raw := range rawMatches {
		if raw.Intent == "none" {
			continue
		}

		var intent FileOperationIntent
		switch raw.Intent {
		case "read":
			intent = IntentRead
		case "write":
			intent = IntentWrite
		case "list":
			intent = IntentList
		case "exists":
			intent = IntentExists
		case "delete":
			intent = IntentDelete
		default:
			continue
		}

		matches = append(matches, FileOperationMatch{
			Intent:   intent,
			FilePath: raw.FilePath,
			Content:  raw.Content,
		})
	}

	return matches
}

// FileOperationMatch represents a matched file operation from the prompt
type FileOperationMatch struct {
	Intent   FileOperationIntent
	FilePath string
	Content  string
}

// ExecuteFileOperations executes file operations based on analyzed intents
func (a *Agent) ExecuteFileOperations(ctx context.Context, matches []FileOperationMatch) []ToolExecutionResult {
	return a.ExecuteFileOperationsWithProgress(ctx, matches, &ConsoleProgressDisplay{})
}

// ExecuteFileOperationsWithProgress executes file operations with custom progress display
func (a *Agent) ExecuteFileOperationsWithProgress(ctx context.Context, matches []FileOperationMatch, progressDisplay ProgressDisplay) []ToolExecutionResult {
	var results []ToolExecutionResult

	// Find the file tool
	var fileTool tools.Tool
	for _, tool := range a.tools {
		if tool.Name() == "file_operations" {
			fileTool = tool
			break
		}
	}

	if fileTool == nil {
		result := ToolExecutionResult{
			Tool:      "file_operations",
			Error:     fmt.Errorf("file tool not registered"),
			StartTime: time.Now(),
		}
		result.Duration = time.Since(result.StartTime)
		results = append(results, result)
		return results
	}

	for _, match := range matches {
		var input string
		var operation string
		var operationDesc string

		switch match.Intent {
		case IntentRead:
			operation = "read"
			operationDesc = fmt.Sprintf("Reading file '%s'", match.FilePath)
			input = fmt.Sprintf("read %s", match.FilePath)
		case IntentWrite:
			operation = "write"
			operationDesc = fmt.Sprintf("Writing to file '%s'", match.FilePath)
			input = fmt.Sprintf("write %s %s", match.FilePath, match.Content)
		case IntentList:
			operation = "list"
			if match.FilePath == "" || match.FilePath == "." {
				operationDesc = "Listing current directory"
				input = "list ."
			} else {
				operationDesc = fmt.Sprintf("Listing directory '%s'", match.FilePath)
				input = fmt.Sprintf("list %s", match.FilePath)
			}
		case IntentExists:
			operation = "exists"
			operationDesc = fmt.Sprintf("Checking existence of '%s'", match.FilePath)
			input = fmt.Sprintf("exists %s", match.FilePath)
		case IntentDelete:
			operation = "delete"
			operationDesc = fmt.Sprintf("Deleting file '%s'", match.FilePath)
			input = fmt.Sprintf("delete %s", match.FilePath)
		default:
			continue
		}

		// Show progress before execution
		progressDisplay.ShowProgress(operation, operationDesc)

		// Execute with timing
		startTime := time.Now()
		output, err := fileTool.Execute(ctx, input)
		duration := time.Since(startTime)

		result := ToolExecutionResult{
			Tool:      operation,
			Input:     input,
			Output:    output,
			Error:     err,
			Duration:  duration,
			StartTime: startTime,
		}

		// Show result after execution
		progressDisplay.ShowResult(result)

		results = append(results, result)
	}

	return results
}

// IntentToString converts FileOperationIntent to string
func IntentToString(intent FileOperationIntent) string {
	switch intent {
	case IntentRead:
		return "read"
	case IntentWrite:
		return "write"
	case IntentList:
		return "list"
	case IntentExists:
		return "exists"
	case IntentDelete:
		return "delete"
	default:
		return "none"
	}
}
