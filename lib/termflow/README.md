# termflow

A minimal terminal UI library for Go that preserves scrollback history.

## Overview

termflow is designed to solve the scrollback history limitation found in full TUI frameworks like bubbletea. Instead of taking over the entire terminal screen, termflow uses normal terminal output for displaying content (which gets preserved in the terminal's scrollback buffer) while providing interactive input features.

## Key Features

- **Scrollback Preservation**: Chat history and output are preserved in the terminal's scrollback buffer
- **Multiline Input**: Support for both single-line and multiline input with intuitive syntax
- **Input History**: Persistent command history with file storage
- **Tab Completion**: Configurable command completion
- **Minimal Dependencies**: Uses only Go standard library plus `golang.org/x/term` for advanced features
- **Cross-Platform**: Works on Unix-like systems and Windows

## Design Philosophy

Unlike full TUI frameworks that:
- Take over the entire terminal screen
- Use alternate screen buffers
- Redraw content on every update
- Lose scrollback history when the app exits

termflow:
- Uses normal terminal output (preserved in scrollback)
- Only manages the input line
- Provides just the interactive features you need
- Maintains all terminal history

## Usage

### Basic Usage

```go
package main

import (
    "github.com/mizzy/rigel/lib/termflow"
)

func main() {
    // Create a basic client
    client := termflow.New()

    // Output text (preserved in scrollback)
    client.Printf("Welcome to my chat app!\n")

    // Read user input (single line)
    input, err := client.ReadLine()
    if err != nil {
        panic(err)
    }

    // Print a chat exchange
    client.PrintChat(input, "AI response here")
}
```

### Multiline Input

termflow supports intuitive multiline input:

```go
// Automatic detection: end first line with "..."
input, err := client.ReadLineOrMultiLine()
// User types: "Write a function..."
// System switches to multiline mode automatically

// Explicit multiline input
multiInput, err := client.ReadMultiLine()
// Always prompts for multiline input
```

**Usage patterns:**
- Single line: `Hello world` → single line input
- Multiline trigger: `Write a function...` → continues on next lines
- End multiline: Type `.` on empty line or press Ctrl+D

### Interactive Features

```go
// Create an interactive client with advanced features
client, err := termflow.NewInteractive()
if err != nil {
    panic(err)
}
defer client.Close()

// Set up persistent history
history := termflow.NewHistoryManager(".myapp_history")
history.Load()
defer history.Save()

// Set up command completion
provider := termflow.NewCompletionProvider()
provider.AddCommand("/help", "Show help")
provider.AddCommand("/quit", "Exit app")

client.SetCompletionFunc(func(input string) []string {
    return provider.GetCompletionStrings(input)
})
```

### History Management

```go
// Create history manager
history := termflow.NewHistoryManager(".myapp_history")

// Load from file
if err := history.Load(); err != nil {
    log.Printf("Could not load history: %v", err)
}

// Add entries
history.Add("user command")

// Get recent entries
recent := history.GetLatest(10)

// Search history
matches := history.Search("search term")

// Save to file
history.Save()
```

## API Reference

### Core Types

- `Client`: Basic terminal client for input/output
- `InteractiveClient`: Advanced client with raw mode support
- `HistoryManager`: Persistent command history management
- `CompletionProvider`: Tab completion functionality

### Key Methods

- `ReadLine()`: Read a line of user input
- `Print()`, `Printf()`: Output text (preserved in scrollback)
- `PrintChat()`: Output formatted chat exchange
- `ShowError()`, `ShowInfo()`: Display formatted messages

## Example

See `example/main.go` for a complete working example that demonstrates:
- Basic input/output
- Command history
- Tab completion
- Command handling

## Comparison with Full TUI Frameworks

| Feature | termflow | bubbletea/tview |
|---------|----------|-----------------|
| Scrollback History | ✅ Preserved | ❌ Lost on exit |
| Complex Layouts | ❌ Not supported | ✅ Full support |
| Input History | ✅ Built-in | ⚠️ Manual implementation |
| Tab Completion | ✅ Built-in | ⚠️ Manual implementation |
| Performance | ✅ Minimal CPU | ⚠️ Continuous rendering |
| Learning Curve | ✅ Simple | ⚠️ Complex |

## When to Use termflow

Choose termflow when:
- You need to preserve terminal scrollback history
- You're building a chat or REPL-style application
- You want simple input/output with some interactive features
- You don't need complex layouts or full-screen TUIs

Choose a full TUI framework when:
- You need complex layouts and widgets
- You're building a full-screen application
- You need real-time updates across the entire screen
- Scrollback history preservation is not important

## License

Same as the parent rigel project.
