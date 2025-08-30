# Rigel - AI Coding Agent

A Go-based AI coding assistant that helps developers write, review, and improve code through natural language interactions.

## Features

- **Code Generation**: Generate code from natural language descriptions
- **Code Review**: Analyze existing code for improvements and potential issues
- **Refactoring**: Suggest and implement code refactoring
- **Documentation**: Generate documentation from code
- **Multi-language Support**: Support for multiple programming languages
- **Context Awareness**: Understand project structure and dependencies
- **Tool Integration**: Integrate with LSP, DAP, Tree-sitter and other development tools

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

```bash
# Start interactive mode
rigel

# Run specific command
rigel generate "Create a function to calculate fibonacci numbers"

# Analyze a file
rigel analyze main.go

# Refactor code
rigel refactor --file main.go --target "improve error handling"
```

## Architecture

```
rigel/
├── cmd/
│   └── rigel/         # CLI entry point
├── internal/
│   ├── agent/         # AI agent implementations
│   ├── llm/           # LLM provider integrations
│   ├── tools/         # Code manipulation tools
│   ├── parser/        # Language-specific parsers
│   ├── analyzer/      # Code analysis engine
│   ├── server/        # HTTP server
│   └── config/        # Configuration management
├── pkg/
│   ├── api/           # Public API
│   ├── types/         # Common type definitions
│   └── utils/         # Utility functions
├── tests/             # Test files
├── docs/              # Documentation
└── examples/          # Example usage
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

## Supported Languages

- JavaScript/TypeScript
- Python
- Java
- Go
- Rust
- C/C++
- Ruby
- PHP

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

- [ ] VSCode extension
- [ ] IntelliJ plugin
- [ ] Web-based interface
- [ ] Team collaboration features
- [ ] Custom model fine-tuning
- [ ] Offline mode support
- [ ] Integration with popular CI/CD pipelines

## Support

For issues and questions, please use the [GitHub Issues](https://github.com/mizzy/rigel/issues) page.
