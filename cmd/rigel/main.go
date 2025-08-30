package main

import (
	"log"
	"os"

	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			log.Printf("Warning: Failed to load config: %v", err)
		}
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate [prompt]",
	Short: "Generate code from a natural language description",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := args[0]
		for i := 1; i < len(args); i++ {
			prompt += " " + args[i]
		}

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize LLM provider: %v", err)
		}

		response, err := provider.Generate(cmd.Context(), prompt)
		if err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}

		os.Stdout.WriteString(response)
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze a code file for improvements and potential issues",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		content, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize LLM provider: %v", err)
		}

		prompt := "Please analyze the following code for improvements, potential issues, and best practices:\n\n```\n" + string(content) + "\n```"

		response, err := provider.Generate(cmd.Context(), prompt)
		if err != nil {
			log.Fatalf("Failed to analyze code: %v", err)
		}

		os.Stdout.WriteString(response)
	},
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start an interactive session with the AI agent",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting interactive mode...")
		log.Println("Type 'exit' or 'quit' to end the session")
		log.Println("Interactive mode not yet implemented")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(interactiveCmd)
}
