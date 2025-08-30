# Rigel - AI Coding Agent

A Go-based AI coding assistant that helps developers write, review, and improve code through natural language interactions.

## Features

- 💬 **AI Chat Assistant**: Powered by multiple LLM providers (Anthropic, OpenAI, Ollama)
- 🔍 **Repository Analysis**: Automatic codebase analysis with `/init` command
- 📄 **AGENTS.md Generation**: Creates AI-friendly documentation of your codebase
- 🖥️ **Code Generation**: Generate code from natural language descriptions
- 💡 **Syntax Highlighting**: Beautiful code rendering with syntax colors
- 🎨 **Clean Interface**: Simple, distraction-free chat interface
- 📝 **Multiline Support**: Natural code and text input
- 🌙 **Tokyo Night Theme**: Modern dark theme for comfortable viewing

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
# Choose a provider: ollama, anthropic, openai, google, azure
PROVIDER=anthropic

# AI Model API Keys (required based on provider)
OPENAI_API_KEY=your_openai_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key
GOOGLE_API_KEY=your_google_api_key
AZURE_OPENAI_API_KEY=your_azure_api_key

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
- 💡 **Syntax Highlighting**: Beautiful code rendering with syntax colors
- 📝 **Multiline Input**: Natural code and text input
- 🎨 **Tokyo Night Theme**: Modern dark theme for comfortable viewing

#### Commands

| Command/Shortcut | Action |
|-----------------|--------|
| `/init` | Analyze repository and generate AGENTS.md |
| `Ctrl+Enter` | Send message |
| `Ctrl+I` | Quick /init command |
| `Ctrl+L` | Clear chat |
| `Ctrl+H` or `?` | Show help |
| `Ctrl+C` or `Ctrl+Q` | Quit |

#### Example Session

```
═══════════════════════════════════════════════════════════════════
                    🤖 Rigel AI Assistant
═══════════════════════════════════════════════════════════════════

assistant 12:34
Welcome to Rigel! Type your message or use /init to analyze the repository.

user 12:35
/init

assistant 12:35
🔍 Analyzing repository structure...

assistant 12:36
✅ Repository analyzed successfully! AGENTS.md has been created.

user 12:36
How do I read a file in Go?

assistant 12:36
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

───────────────────────────────────────────────────────────────────
Type your message or /init to analyze repository... (Ctrl+Enter to send)
───────────────────────────────────────────────────────────────────
/init: Analyze repo | Ctrl+Enter: Send | Ctrl+L: Clear | Ctrl+C: Quit | ?: Help
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
│   ├── agent/         # AI agent core
│   ├── llm/           # LLM provider integrations
│   ├── tui/           # Terminal UI components
│   │   ├── model.go   # Bubbletea model
│   │   ├── components.go # UI components
│   │   ├── theme.go   # Color themes
│   │   └── syntax.go  # Syntax highlighting
│   ├── tools/         # Code manipulation tools
│   └── config/        # Configuration management
├── examples/          # Example usage
└── integration_test.go # Integration tests
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

- **Anthropic** (Claude models)
- **OpenAI** (GPT models) - Coming soon
- **Ollama** (Local models)
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

- [x] Rich TUI interface with Bubbletea
- [x] Syntax highlighting for code blocks
- [x] File explorer integration
- [x] Command palette
- [x] Session management
- [ ] VSCode extension
- [ ] Web-based interface
- [ ] Plugin system for custom tools
- [ ] Integration with LSP servers
- [ ] Git integration in TUI
- [ ] Multi-tab support
- [ ] Custom model fine-tuning

## Support

For issues and questions, please use the [GitHub Issues](https://github.com/mizzy/rigel/issues) page.
