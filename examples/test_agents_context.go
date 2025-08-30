package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mizzy/rigel/internal/llm"
)

// This example demonstrates how AGENTS.md content is automatically included
// in the LLM context when making requests.
func main() {
	// Check if AGENTS.md exists
	if _, err := os.Stat("AGENTS.md"); os.IsNotExist(err) {
		fmt.Println("AGENTS.md not found in current directory")
		fmt.Println("The LLM will proceed without repository context")
	} else {
		fmt.Println("AGENTS.md found - it will be included in the LLM context")
	}

	// Initialize provider (example with Anthropic)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	provider, err := llm.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-20241022")
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Make a request - AGENTS.md will be automatically included
	ctx := context.Background()
	response, err := provider.GenerateWithOptions(ctx,
		"What is the main purpose of this repository based on the context?",
		llm.GenerateOptions{
			SystemPrompt: "You are a helpful coding assistant.",
			Temperature:  0.7,
		})

	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	fmt.Println("\nLLM Response:")
	fmt.Println(response)
}
