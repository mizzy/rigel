# Termflow UI Testing Framework

A Playwright-style automated testing framework for terminal applications.

## Overview

This framework provides tools for automated UI testing of termflow-based applications. It simulates real terminal environments, allowing you to simulate user interactions and validate outputs.

## Features

### 1. **TerminalTest** - Virtual Terminal Testing
- Uses actual PTY (pseudo-terminal)
- Real keyboard input simulation
- ANSI escape sequence support
- Spinner animation validation

### 2. **MockIO** - Lightweight Unit Testing
- Mock I/O for component-level testing
- Fast execution
- Output format validation

## Usage

### Basic Testing

```go
func TestMyApp(t *testing.T) {
    // Start application
    tt, err := NewTerminalTest(t, "./my-app", "--termflow")
    if err != nil {
        t.Fatalf("Failed to start: %v", err)
    }
    defer tt.Close()

    // Wait for startup
    tt.Wait(500 * time.Millisecond)

    // Verify welcome message
    tt.ExpectWelcome()
    tt.ExpectPrompt()

    // Simulate user input
    tt.Type("hello world")

    // Wait for response and validate
    tt.Wait(1 * time.Second)
    tt.ExpectOutput("Hello!")
}
```

### Advanced Testing

```go
func TestSpinnerAndCtrlC(t *testing.T) {
    tt, err := NewTerminalTest(t, "./app", "--termflow")
    if err != nil {
        t.Fatalf("Failed to start: %v", err)
    }
    defer tt.Close()

    tt.Wait(500 * time.Millisecond)

    // Trigger long-running task
    tt.Type("complex task")

    // Verify spinner animation
    tt.ExpectSpinner()
    tt.ExpectThinking()

    // Test Ctrl+C behavior
    tt.SendCtrlC()
    tt.ExpectOutput("Press Ctrl+C again to exit")
    tt.ExpectNoCtrlC() // Verify ^C characters are not displayed

    // Second Ctrl+C
    tt.SendCtrlC()
    tt.ExpectOutput("Goodbye!")
}
```

### Multiline Input Testing

```go
func TestMultilineInput(t *testing.T) {
    tt, err := NewTerminalTest(t, "./app", "--termflow")
    if err != nil {
        t.Fatalf("Failed to start: %v", err)
    }
    defer tt.Close()

    tt.Wait(500 * time.Millisecond)

    // Type first line, then insert a newline with Ctrl+J
    tt.SendKeys("def hello():")
    tt.SendCtrlJ() // Insert newline without submitting
    tt.SendKeys("    print('Hello')")
    tt.SendEnter() // Submit the multiline input

    tt.ExpectThinking()
}
```

## Available Methods

### Input Operations
- `SendKeys(string)` - Send key input
- `Type(string)` - Type text + Enter
- `SendEnter()` - Send Enter key
- `SendCtrlC()` - Send Ctrl+C
- `Wait(duration)` - Wait for specified duration

### Output Validation
- `ExpectOutput(string)` - Verify text is present in output
- `ExpectPattern(regex)` - Verify regex pattern matches
- `ExpectPrompt()` - Verify prompt (✦) is present
- `ExpectWelcome()` - Verify welcome message
- `ExpectSpinner()` - Verify spinner animation
- `ExpectThinking()` - Verify "Thinking..." message
- `ExpectNoCtrlC()` - Verify ^C characters are not displayed

### Debugging
- `GetOutput()` - Get raw output (including ANSI codes)
- `GetVisibleOutput()` - Get visible text (ANSI codes removed)
- `Screenshot()` - Format and display current screen state
- `GetLines()` - Get output as array of lines

## Running Tests

```bash
# Basic tests
go test ./lib/termflow/uitest

# Integration tests (requires binary build)
go build -o /tmp/rigel-test cmd/rigel/main.go
go test ./lib/termflow/uitest -v

# Benchmarks
go test -bench=. ./lib/termflow/uitest

# Short tests only (skip integration tests)
go test -short ./lib/termflow/uitest
```

## Architecture

```
┌─────────────────────────────────────┐
│           Test Code                 │
├─────────────────────────────────────┤
│      TerminalTest Framework         │
├─────────────────────────────────────┤
│    PTY (Pseudo Terminal)            │
├─────────────────────────────────────┤
│    Target Application               │
│    (Rigel with --termflow)          │
└─────────────────────────────────────┘
```

## Limitations

- Requires PTY support on Linux/macOS
- Limited Windows support (WSL recommended)
- Application build required
- Timing-dependent tests may be unstable

## Best Practices

1. **Proper Wait Times**: Use `Wait()` with sufficient duration
2. **Use Screenshots**: Use `Screenshot()` for debugging
3. **Step-by-Step Validation**: Break large operations into small verification steps
4. **Environment Initialization**: Use new TerminalTest instance for each test
5. **Error Handling**: Always call Close() with defer

This framework enables automated quality assurance for termflow applications through comprehensive testing.
