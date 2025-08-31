# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build and Run
```bash
# Build the binary
make build                    # Output to ./bin/rigel
go build -o rigel cmd/rigel/main.go  # Direct build

# Run in development mode
make run                      # Run without building
go run cmd/rigel/main.go      # Direct run
make dev                      # Build and run

# Install globally
make install                  # Install to $GOPATH/bin
go install ./cmd/rigel
```

### Testing
```bash
# Run all tests
make test                     # Verbose test output
go test ./...                 # Standard test run

# Run tests for specific package
go test ./internal/llm -v    # Test LLM package
go test ./internal/ui -v     # Test UI package

# Test with coverage
make test-coverage            # Coverage report
go test -cover ./...          # Direct coverage

# Run single test
go test -run TestAnthropicProvider ./internal/llm

# Integration tests (requires API keys)
ANTHROPIC_API_KEY=xxx go test ./internal/llm -run TestAnthropicProvider_Generate
```

### Linting and Formatting
```bash
# Run linter (if golangci-lint installed)
make lint
golangci-lint run

# Format code
go fmt ./...

# Static analysis
staticcheck ./...

# Pre-commit hooks (if configured)
pre-commit run --all-files
```

## High-Level Architecture

### Core Components

**Terminal UI System** (`internal/ui/terminal/`)
- `model.go`: Main chat model managing state, input handling, and rendering
- `commands.go`: Slash command handlers (`/init`, `/model`, `/provider`, `/help`)
- Provider and model selection are handled through interactive modal interfaces
- Uses Bubbletea framework for terminal UI with inline mode (no alternate screen)

**LLM Provider System** (`internal/llm/`)
- `provider.go`: Common interface for all LLM providers
- `anthropic.go`: Anthropic implementation with API model listing and fallback
- `ollama.go`: Ollama local model support
- Provider switching happens dynamically through config updates and provider recreation

**Command Flow**
1. User types `/provider` or `/model` command
2. Command handler fetches available options (from API or config)
3. Interactive selector UI is shown (arrow keys to navigate, Enter to select)
4. On selection, provider/model is switched and config updated
5. Chat continues with new provider/model

### Key Design Patterns

**Provider Selection**:
- Runtime provider switching without restart
- Config object passed to ChatModel for dynamic updates
- Provider interface allows transparent switching between Anthropic/Ollama

**Model Discovery**:
- Anthropic: Attempts API call to `/v1/models`, falls back to hardcoded list
- Ollama: Uses local API to list installed models
- Default models: Claude Sonnet 4 for Anthropic, gpt-oss:20b for Ollama

**State Management**:
- ChatModel maintains provider, config, and UI state
- Selection modes (provider/model) are exclusive states
- ESC key properly exits selection modes and resets thinking state

### Configuration

**Environment Variables** (`.env` file or shell):
```
PROVIDER=anthropic|ollama
ANTHROPIC_API_KEY=xxx
MODEL=claude-sonnet-4-20250514  # Optional, uses provider defaults
OLLAMA_BASE_URL=http://localhost:11434
```

**Provider Defaults**:
- Anthropic: `claude-sonnet-4-20250514`
- Ollama: `gpt-oss:20b`
- OpenAI (planned): `gpt-4-turbo-preview`

### Repository Analysis

The `/init` command generates `AGENTS.md` by:
1. Analyzing repository structure and code patterns
2. Using LLM to generate comprehensive documentation
3. Prepending AGENTS.md content to system prompts automatically

This context helps the AI understand the codebase structure when answering questions.

## Important Implementation Details

- **Terminal Colors**: Uses Rigel theme (blue #5793ff for highlights)
- **Input Handling**: Alt+Enter for multiline, Tab for command completion
- **Error Recovery**: API failures fall back to hardcoded model lists
- **Testing**: Most tests mock LLM responses with httptest
- **Security**: No API keys in code, loaded from environment only
