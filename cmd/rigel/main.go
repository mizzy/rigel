package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mizzy/rigel/internal/config"
	"github.com/mizzy/rigel/internal/llm"
	"github.com/mizzy/rigel/internal/sandbox"
	"github.com/mizzy/rigel/internal/ui/terminal"
	"github.com/mizzy/rigel/internal/version"
	"github.com/spf13/cobra"
)

var (
	cfg           *config.Config
	sandboxFlag   bool
	noSandboxFlag bool
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
		// Handle sandbox mode (default enabled on macOS)
		if !noSandboxFlag && (sandboxFlag || shouldEnableSandboxByDefault()) {
			if !sandbox.IsSandboxed() {
				if err := sandbox.EnableSandbox("."); err != nil {
					log.Printf("Warning: Failed to enable sandbox: %v", err)
					log.Println("Running without sandbox restrictions.")
				}
				// If EnableSandbox succeeds, it will re-exec and exit
			}
		}

		// Show sandbox status
		if sandbox.IsSandboxed() {
			fmt.Fprintln(os.Stderr, "🔒 Sandbox enabled: File writes restricted to current directory")
		} else if noSandboxFlag {
			fmt.Fprintln(os.Stderr, "⚠️  Running without sandbox. File operations are unrestricted.")
		}
		var err error
		cfg, err = config.Load("")
		if err != nil {
			log.Printf("Warning: Failed to load config: %v", err)
		}

		provider, err := llm.NewProvider(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize LLM provider: %v", err)
		}

		// Check if input is piped
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Handle piped input - no interactive commands in pipe mode
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("Failed to read from stdin: %v", err)
			}

			prompt := strings.TrimSpace(string(input))
			if prompt == "" {
				log.Fatal("No input provided")
			}

			// In pipe mode, slash commands are not supported
			if strings.HasPrefix(prompt, "/") {
				fmt.Fprintf(os.Stderr, "Slash commands like %s are only available in interactive mode.\n", prompt)
				fmt.Fprintf(os.Stderr, "Run 'rigel' without piping input to use interactive mode.\n")
				os.Exit(1)
			}

			// Generate response for the prompt
			response, err := provider.Generate(cmd.Context(), prompt)
			if err != nil {
				log.Fatalf("Failed to generate response: %v", err)
			}

			fmt.Print(response)
		} else {
			// Run interactive chat mode (inline, no alternate screen)
			runChatMode(provider)
		}
	},
}

func runChatMode(provider llm.Provider) {
	model := terminal.NewModel(provider, cfg)
	p := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running chat: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolVar(&sandboxFlag, "sandbox", false, "Force enable sandbox mode (default on macOS)")
	rootCmd.Flags().BoolVar(&noSandboxFlag, "no-sandbox", false, "Disable sandbox mode explicitly")
}

func shouldEnableSandboxByDefault() bool {
	// Enable sandbox by default on macOS
	// Can be expanded to other platforms in the future
	return runtime.GOOS == "darwin"
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of rigel",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.String())
	},
}
