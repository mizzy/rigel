package command

// Command represents a command with its description
type Command struct {
	Command     string
	Description string
}

// AvailableCommands contains all available commands
var AvailableCommands = []Command{
	{"/init", "Analyze repository and generate AGENTS.md"},
	{"/model", "Show current model and select from available models"},
	{"/provider", "Switch between LLM providers (Anthropic, Ollama, etc.)"},
	{"/status", "Show current session status and configuration"},
	{"/help", "Show available commands"},
	{"/clear", "Clear chat history"},
	{"/clearhistory", "Clear command history"},
	{"/exit", "Exit the application"},
	{"/quit", "Exit the application"},
}
