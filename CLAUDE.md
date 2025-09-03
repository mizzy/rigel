# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build and Run
```bash
# Build the binary with version info
make build                    # Output to ./bin/rigel with version/git info
make deps                     # Download and tidy dependencies first
go build -o rigel cmd/rigel/main.go  # Direct build (no version info)

# Run in development mode
make run                      # Run without building (go run cmd/rigel/main.go)
make dev                      # Download deps + build + run
PROVIDER=ollama ./bin/rigel   # Run with specific provider

# Run with different UI modes
./bin/rigel --termflow        # Use termflow UI (preserves scrollback)
./bin/rigel                   # Use bubbletea UI (default)

# Security options
./bin/rigel --sandbox         # Force enable sandbox mode (macOS)
./bin/rigel --no-sandbox      # Disable sandbox mode

# Install globally
make install                  # Install to $GOPATH/bin with version info
go install ./cmd/rigel
```

### Testing
```bash
# Run all tests
make test                     # Verbose test output
go test ./...                 # Standard test run

# Run tests for specific packages
go test ./internal/llm -v              # Test LLM package
go test ./internal/ui/terminal -v      # Test Terminal UI
go test ./internal/ui/termflow -v      # Test Termflow UI
go test ./lib/termflow -v              # Test Termflow library

# Test with coverage
make test-coverage            # Coverage report
go test -cover ./...          # Direct coverage

# Run specific tests
go test -run TestAnthropicProvider ./internal/llm
go test -run TestCtrlCBehavior ./internal/ui/termflow
go test -run TestHistoryPersistence ./internal/ui/termflow

# Integration and UI tests (requires RIGEL_TEST_MODE=1)
RIGEL_TEST_MODE=1 PROVIDER=ollama go test ./internal/ui/termflow -v
RIGEL_TEST_MODE=1 PROVIDER=ollama go test ./lib/termflow/uitest -v

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

### Git Branch Management

**After PR Merge - Always Update Main Branch:**
```bash
# Switch to main and pull latest changes
git checkout main
git pull

# Clean up local branches (delete merged feature branch)
git branch -D feature/branch-name

# Remove stale remote branch references
git remote prune origin
```

**Create Feature Branch Workflow:**
```bash
# Start from updated main
git checkout main
git pull

# Create and switch to new feature branch
git checkout -b feature/descriptive-name

# After development, push and create PR
git push -u origin feature/descriptive-name
gh pr create --title "feat: Description" --body "Detailed description"
```

## High-Level Architecture

### Dual UI System

**Termflow UI** (`lib/termflow/` + `internal/ui/termflow/`)
- Custom terminal library that preserves scrollback buffer
- Raw terminal mode with advanced line editing capabilities
- Multiline input support with `...` syntax
- History navigation (Up/Down arrows) with file persistence
- Tab completion for commands and models
- Two-press Ctrl+C exit pattern (first press cancels input, second exits)
- PTY-based testing framework (`lib/termflow/uitest/`)

**Terminal UI** (`internal/ui/terminal/`)
- Traditional Bubbletea-based interface with inline mode
- Modal interfaces for provider/model selection
- Spinner animations and structured output
- No alternate screen mode (preserves terminal history)

### Core Systems

**Agent System** (`internal/agent/`)
- Context-aware AI agent with tool integration
- Task detection and automatic execution
- Progress tracking with UI feedback
- Conversation memory management and state persistence

**Tool Integration** (`internal/tools/`)
- `file_tool.go`: File operations (read, write, list, exists, delete)
- `code_tool.go`: Code analysis and manipulation tools
- Automatic tool selection based on prompt analysis
- Tool execution results integrated into AI responses

**LLM Provider System** (`internal/llm/`)
- `provider.go`: Common interface for all LLM providers
- `anthropic.go`: Anthropic Claude models with API model listing
- `ollama.go`: Local Ollama model support
- `agents_loader.go`: Loads AGENTS.md context for repository understanding
- Provider switching happens dynamically at runtime

**Command System** (`internal/command/`)
- Centralized command processing with tab completion
- Available commands: `/init`, `/model`, `/provider`, `/status`, `/help`, `/clear`, `/clearhistory`, `/exit`, `/quit`
- Async command execution with progress feedback
- Command history persistence

**State Management**
- `internal/state/chat.go`: Chat history and session state
- `internal/state/llm.go`: LLM configuration and provider selection
- `internal/history/`: Persistent command and chat history
- Cross-session state preservation

**Security System** (`internal/sandbox/`)
- macOS sandbox support using `sandbox-exec`
- Restricts file operations to current directory and `.rigel` folder
- Auto-enabled on macOS, can be disabled with `--no-sandbox`

### Key Design Patterns

**UI Mode Selection**:
- `--termflow` flag enables termflow UI (recommended for scrollback preservation)
- Default uses bubbletea UI for compatibility
- Both UIs share same agent/tool backend systems

**Provider and Model Management**:
- Default provider: `ollama` for out-of-box experience
- Runtime provider switching without application restart
- Model discovery through provider APIs with fallback lists
- Configuration validation ensures required API keys are present

**Tool Integration Pattern**:
- Tools automatically selected based on prompt analysis
- File operations sandboxed to current directory
- Tool results seamlessly integrated into conversation flow
- Progress feedback during long-running operations

**Testing Architecture**:
- PTY-based testing for terminal UI interactions
- Mock providers for isolated LLM testing
- `RIGEL_TEST_MODE=1` environment variable enables test mode
- Comprehensive test coverage for UI behaviors and tool integration

### Configuration

**Environment Variables** (`.env` file or shell):
```
PROVIDER=ollama|anthropic        # Default: ollama
ANTHROPIC_API_KEY=xxx           # Required for Anthropic provider
MODEL=claude-sonnet-4-20250514  # Optional, uses provider defaults
OLLAMA_BASE_URL=http://localhost:11434  # Default Ollama endpoint
RIGEL_TEST_MODE=1               # Enable test mode for UI testing
```

**Provider Defaults**:
- **Ollama** (default): `gpt-oss:20b`
- **Anthropic**: `claude-sonnet-4-20250514`
- **OpenAI** (planned): `gpt-4-turbo-preview`

**Command Line Flags**:
```bash
--termflow      # Use termflow UI (recommended)
--sandbox       # Force enable sandbox mode
--no-sandbox    # Disable sandbox mode
--version       # Show version information
```

### Repository Analysis and Agent Context

The `/init` command generates `AGENTS.md` by:
1. Analyzing repository structure, dependencies, and code patterns
2. Using LLM to generate comprehensive codebase documentation
3. Automatically prepending AGENTS.md content to system prompts
4. Enabling context-aware responses about the codebase

This provides the AI agent with deep understanding of:
- Project architecture and design patterns
- Build and test workflows
- Available tools and frameworks
- Code conventions and best practices

### Tool Integration System

**Automatic Tool Detection**:
- Agent analyzes user prompts to determine required tools
- File operations, code analysis, and repository queries handled automatically
- Progress feedback shown during tool execution
- Tool results integrated seamlessly into conversation

**Available Tools**:
- **File Tools**: Read, write, list, delete files (sandboxed to current directory)
- **Code Tools**: Code analysis, pattern detection, refactoring assistance
- **Repository Tools**: Git integration, dependency analysis, structure mapping

## Important Implementation Details

### Terminal UI Specifics
- **Termflow Mode**: Preserves scrollback, multiline with `...`, history navigation
- **Bubbletea Mode**: Modal interfaces, spinner feedback, structured output
- **Colors**: Rigel theme with blue (#5793ff) highlights
- **Input**: Tab completion, Alt+Enter for multiline (bubbletea only)

### Security and Sandboxing
- **macOS Sandbox**: Auto-enabled, restricts writes to current directory + `.rigel/`
- **API Security**: Keys loaded from environment, never stored in code
- **Tool Sandboxing**: File operations restricted to project directory

### Testing Infrastructure
- **Mock Testing**: LLM providers mocked with `httptest` for isolation
- **PTY Testing**: Terminal interactions tested with pseudoterminals
- **Integration Testing**: Full agent workflows with `RIGEL_TEST_MODE=1`
- **UI Testing**: Both termflow and bubbletea UIs comprehensively tested

### Performance and Error Handling
- **Provider Fallbacks**: API failures gracefully fall back to hardcoded model lists
- **State Persistence**: Chat and command history saved across sessions
- **Memory Management**: Conversation context managed for optimal performance
- **Async Operations**: Long-running tasks with progress feedback and cancellation
