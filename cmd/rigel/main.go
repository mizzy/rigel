package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "rigel",
	Short: "AI Coding Agent - Your intelligent coding assistant",
	Long: `Rigel is an AI-powered coding assistant that helps developers write,
review, and improve code through natural language interactions.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.Load("")
		if err != nil {
			log.Printf("Warning: Failed to load config: %v", err)
		}

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize LLM provider: %v", err)
		}

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("Failed to read from stdin: %v", err)
			}

			prompt := strings.TrimSpace(string(input))
			if prompt == "" {
				log.Fatal("No input provided")
			}

			response, err := provider.Generate(cmd.Context(), prompt)
			if err != nil {
				log.Fatalf("Failed to generate response: %v", err)
			}

			fmt.Print(response)
		} else {
			runInteractiveMode(cmd, provider)
		}
	},
}

func runInteractiveMode(cmd *cobra.Command, provider llm.Provider) {
	fmt.Println("Welcome to Rigel interactive mode!")
	fmt.Println("Type 'exit' or 'quit' to end the session")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		response, err := provider.Generate(cmd.Context(), input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

func init() {
}
