# Rigel - AI Coding Agent

A Go-based AI coding assistant that helps developers write, review, and improve code through natural language interactions.

## Features

- 💬 **AI Chat Assistant**: Powered by multiple LLM providers (Anthropic, Ollama)
- 🔍 **Repository Analysis**: Automatic codebase analysis with `/init` command
- 📄 **AGENTS.md Generation**: Creates AI-friendly documentation of your codebase
- 🖥️ **Code Generation**: Generate code from natural language descriptions
- 🎨 **Clean Interface**: Simple, distraction-free chat interface
- 📝 **Multiline Support**: Natural code and text input with Alt+Enter
- ✨ **Command Autocomplete**: Tab completion for slash commands
- 🎯 **Rigel Theme**: Blue-white star themed interface

## Requirements

- Go 1.25 or higher
- Git

## Installation

```bash
# Clone the repository
git clone https://github.com/mizzy/rigel.git
cd rigel

# Download dependencies
go mod download

# Build
go build -o rigel cmd/rigel/main.go

# Install (optional)
go install ./cmd/rigel

# Set up environment variables
cp .env.example .env
# Edit .env with your API keys
```

## Configuration

By default, Rigel uses Ollama with the `gpt-oss:20b` model. No API keys are required for the default configuration.

### Default Configuration (Ollama)

The application works out of the box with Ollama running locally:
- Provider: `ollama`
- Model: `gpt-oss:20b`
- Base URL: `http://localhost:11434`

Make sure Ollama is installed and running locally:
```bash
# Install Ollama (if not already installed)
curl -fsSL https://ollama.com/install.sh | sh

# Pull the default model
ollama pull gpt-oss:20b

# Start Ollama server (if not already running)
ollama serve
```

### Custom Configuration

Create a `.env` file to use different providers or models:

```bash
# Choose a provider: ollama, anthropic
PROVIDER=anthropic

# AI Model API Keys (required based on provider)
ANTHROPIC_API_KEY=your_anthropic_api_key
# OPENAI_API_KEY=your_openai_api_key        # Coming soon
# GOOGLE_API_KEY=your_google_api_key        # Coming soon
# AZURE_OPENAI_API_KEY=your_azure_api_key   # Coming soon

# Custom model (optional, defaults based on provider)
MODEL=claude-3-5-sonnet-20241022

# Ollama configuration (when using Ollama)
OLLAMA_BASE_URL=http://localhost:11434

# Logging
RIGEL_LOG_LEVEL=info
```

## Usage

### Interactive Chat Mode

Rigel features a clean and simple chat interface for AI-assisted coding:

```bash
# Start Rigel
rigel
```

#### Features

- 💬 **Simple Chat Interface**: Clean, distraction-free chat with AI
- 🔍 **Repository Analysis**: Use `/init` to analyze your codebase and generate AGENTS.md
- 📝 **Multiline Input**: Natural code and text input with Alt+Enter
- ✨ **Command Autocomplete**: Tab completion and navigation for slash commands
- 🎯 **Rigel Theme**: Blue-white star themed interface

#### Commands

| Command/Shortcut | Action |
|-----------------|--------|
| `/init` | Analyze repository and generate AGENTS.md |
| `/help` | Show available commands |
| `/clear` | Clear chat history |
| `/exit` or `/quit` | Exit the application |
| `Enter` | Send message |
| `Alt+Enter` | New line |
| `Tab` | Complete command |
| `↑/↓` | Navigate suggestions |
| `Ctrl+C` (twice) | Exit |

#### Example Session

```
✦ /init

✅ Repository analyzed successfully! AGENTS.md has been created.

The file contains:
• Repository structure and overview
• Key components and their responsibilities
• File purposes and dependencies
• Testing and configuration information

✦ How do I read a file in Go?

To read a file in Go, you have several options. Here's the most common approach:

```go
import (
    "os"
    "io"
)

func readFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}
```

✦ █ Type a message or / for commands (Alt+Enter for new line)
```

### Non-Interactive Mode

You can also use Rigel with pipes and scripts:

```bash
# Pipe input
echo "Write a hello world in Python" | rigel

# Use with heredocs
rigel << EOF
Explain this code:
$(cat main.go)
EOF

# Read from file
cat prompt.txt | rigel
```

## Architecture

```
rigel/
├── cmd/
│   └── rigel/         # CLI entry point
├── internal/
│   ├── analyzer/      # Repository analysis
│   ├── llm/           # LLM provider integrations
│   │   ├── anthropic.go  # Anthropic Claude integration
│   │   └── ollama.go     # Ollama local models
│   ├── tui/           # Terminal UI components
│   │   ├── chat.go       # Main chat model
│   │   ├── commands.go   # Command handling
│   │   ├── suggestions.go # Autocomplete logic
│   │   └── styles.go     # Color scheme
│   └── config/        # Configuration management
```

## Development

### Setup Pre-commit Hooks

```bash
# Install pre-commit (if not already installed)
# macOS
brew install pre-commit

# Linux/Windows (via pip)
pip install pre-commit

# Install git hooks
pre-commit install

# Run hooks manually on all files
pre-commit run --all-files
```

### Development Commands

```bash
# Run in development mode
go run cmd/rigel/main.go

# Run tests
go test ./...

# Test coverage
go test -cover ./...

# Benchmark tests
go test -bench=. ./...

# Static analysis
staticcheck ./...

# Build
make build

# Build Docker image
docker build -t rigel:latest .
```

## Supported LLM Providers

### Currently Supported
- **Anthropic** (Claude models) - Full support
- **Ollama** (Local models) - Full support

### Planned
- **OpenAI** (GPT models) - Coming soon
- **Google** (Gemini models) - Coming soon
- **Azure OpenAI** - Coming soon

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

- Inspired by Claude Code and GitHub Copilot
- Built with modern AI models and language processing tools

## Roadmap

### Completed
- [x] Interactive chat interface with Bubbletea
- [x] Repository analysis with `/init` command
- [x] Command autocomplete with Tab
- [x] Multiline input support
- [x] Multiple LLM provider support (Anthropic, Ollama)

### In Progress
- [ ] Syntax highlighting for code blocks
- [ ] Session management (save/load conversations)

### Planned
- [ ] Additional LLM providers (OpenAI, Google, Azure)
- [ ] File operations commands (/read, /write, /edit)
- [ ] Git integration in TUI
- [ ] Custom themes
- [ ] Plugin system for custom tools
- [ ] Web-based interface
- [ ] VSCode extension

## Support

For issues and questions, please use the [GitHub Issues](https://github.com/mizzy/rigel/issues) page.
