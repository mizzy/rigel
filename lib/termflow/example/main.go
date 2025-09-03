// Package main provides an example usage of the termflow library
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/mizzy/rigel/lib/termflow"
)

func main() {
	// Create a new interactive termflow client
	client, err := termflow.NewInteractive()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Set up history management
	history := termflow.NewHistoryManager(".termflow_example_history")
	if err := history.Load(); err != nil {
		fmt.Printf("Warning: Could not load history: %v\n", err)
	}
	defer history.Save()

	// Set up command completion
	provider := termflow.NewCompletionProvider()
	provider.AddCommand("/help", "Show available commands")
	provider.AddCommand("/history", "Show command history")
	provider.AddCommand("/multiline", "Test multiline input")
	provider.AddCommand("/clear", "Clear the screen")
	provider.AddCommand("/quit", "Exit the application")
	provider.AddCommand("/exit", "Exit the application")

	client.SetCompletionFunc(func(input string) []string {
		return provider.GetCompletionStrings(input)
	})

	// Welcome message
	client.ShowInfo("Welcome to termflow example!")
	client.Printf("This demonstrates how termflow preserves scrollback history.\n")
	client.Printf("Try typing some commands:\n")
	client.Printf("  /help - Show help\n")
	client.Printf("  /history - Show input history\n")
	client.Printf("  /multiline - Test multiline input\n")
	client.Printf("  /quit - Exit\n")
	client.Printf("\nFor multiline input, end your first line with '...'\n")
	client.Printf("Example: 'Write a function...' then continue on next lines\n\n")

	// Main loop
	for {
		input, err := client.ReadLineOrMultiLine()
		if err != nil {
			client.ShowError(err)
			break
		}

		// Add to history
		history.Add(input)

		// Handle commands
		switch {
		case input == "/quit" || input == "/exit":
			client.ShowInfo("Goodbye!")
			break

		case input == "/help":
			client.Printf("\nAvailable commands:\n")
			client.Printf("  /help      - Show this help\n")
			client.Printf("  /history   - Show command history\n")
			client.Printf("  /multiline - Test multiline input\n")
			client.Printf("  /clear     - Clear the screen\n")
			client.Printf("  /quit      - Exit the application\n")
			client.Printf("\nFor multiline input, end your first line with '...'\n\n")

		case input == "/history":
			entries := history.GetLatest(10)
			if len(entries) == 0 {
				client.ShowInfo("No history available")
			} else {
				client.Printf("\nRecent command history:\n")
				for i, entry := range entries {
					client.Printf("  %d. %s\n", i+1, entry)
				}
				client.Printf("\n")
			}

		case input == "/multiline":
			client.Printf("\nTesting multiline input:\n")
			multiInput, err := client.ReadMultiLine()
			if err != nil {
				client.ShowError(err)
			} else {
				client.Printf("\nYou entered:\n%s\n\n", multiInput)
				history.Add(fmt.Sprintf("/multiline: %s", multiInput))
			}

		case input == "/clear":
			// This would clear the screen, but we'll just show a message
			client.ShowInfo("Screen cleared (in a real app, this would clear the terminal)")

		case strings.HasPrefix(input, "/"):
			client.ShowError(fmt.Errorf("unknown command: %s", input))

		case strings.TrimSpace(input) == "":
			// Empty input, do nothing

		default:
			// Echo the input as a chat response
			client.PrintChat(input, fmt.Sprintf("You said: %s", input))
		}
	}
}
