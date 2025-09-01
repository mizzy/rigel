# AGENTS.md

> **⚠️** This file is generated automatically from repository metadata.
> **If you notice inaccuracies, please open an issue or modify the source files accordingly.**

---

## 1. Repository Overview

| Item | Description |
|------|-------------|
| **Project name** | **Rigel** |
| **Type** | Command‑line tool written in Go |
| **Purpose** | A lightweight, extensible LLM‑powered command executor that lets users build and run *agents* in a sandboxed environment. It is designed to help developers prototype “AI‑powered assistants” that can read code, run shell commands, and manage state across multiple turns. |

> Rigel is open‑source and aims to lower the barrier to creating and experimenting with autonomous agents that interact with a filesystem, LLM, and external tools.

---

## 2. Main Components

| Package | Purpose | Key responsibilities |
|---------|---------|----------------------|
| `cmd/rigel` | CLI entry point | Parses flags, loads config, initializes the agent engine, and starts the REPL or one‑off commands |
| `internal/agent` | Core agent logic | Maintains the agent’s memory, orchestrates planning → execution cycles, and exposes the public `Agent` interface |
| `internal/analyzer` | Contextual analyzer | Extracts useful information from files, directories, or LLM responses to provide richer prompts |
| `internal/command` | Command registration & execution | Holds all available commands, their signatures, help text, and the dispatch logic |
| `internal/config` | Configuration management | Loads YAML/JSON config files, handles defaults, and provides helpers to validate the schema |
| `internal/git` | Git utilities | Offers helpers to read the current repo state, diff files, and inspect commit history |
| `internal/history` | Interaction history | Stores chat and command history per agent instance, enabling replay or rollback |
| `internal/llm` | LLM abstraction layer | Provides `Provider` interface, concrete back‑ends (`Anthropic`, `Ollama`, etc.), and utilities for prompt formatting |
| `internal/sandbox` | Execution sandbox | Runs commands inside isolated processes, captures stdout/stderr, and enforces resource limits |
| `internal/state` | Runtime state | Persists the agent’s chat log, variables, and LLM metadata across turns |
| `internal/tools` | External tools | Implements domain‑specific utilities (e.g., `code_tool.go` for generating or refactoring code) |
| `internal/command/types.go` | Type definitions | Defines structures used across command handling (arguments, results, error types) |
| `examples/` | Usage examples | Demonstrates how to write an agent, interact with the CLI, and run tests |

---

## 3. Key Files

| File | Why it matters |
|------|----------------|
| `cmd/rigel/main.go` | The main entry point for the CLI. Parses global flags, loads config, and boots the `AgentEngine`. |
| `internal/agent/agent.go` | Implements the `Agent` struct – the heart of the agent lifecycle. |
| `internal/analyzer/analyzer.go` | Turns raw file contents into structured data for LLM prompts. |
| `internal/command/commands.go` | Registers built‑in commands and their handlers. |
| `internal/command/handler.go` | Dispatches commands based on user input. |
| `internal/config/config.go` | Loads and validates the user’s configuration file. |
| `internal/git/git.go` | Provides repo introspection, used by commands like `git diff` or context-aware code generation. |
| `internal/history/history.go` | Stores chat and command history for replay/debugging. |
| `internal/llm/agents_loader.go` | Loads the LLM provider (Anthropic, Ollama, etc.) at runtime based on config. |
| `internal/llm/anthropic.go`, `internal/llm/ollama.go` | Concrete implementations of the `Provider` interface. |
| `internal/llm/provider.go` | Defines the LLM interface expected by the agent. |
| `internal/sandbox/sandbox.go` | Isolates command execution to avoid side effects. |
| `internal/state/chat.go`, `internal/state/llm.go` | Persistent state of the agent’s conversation and LLM context. |
| `internal/tools/code_tool.go` | Example of a custom tool that can be invoked by the agent to modify code. |
| `examples/test_agents_context.go` | Shows how to construct an agent context programmatically for testing. |
| `scripts/` | Helper scripts (e.g., `generate.sh`, `lint.sh`). |

---

## 4. Development Information

### Prerequisites

| Item | Version |
|------|---------|
| Go | 1.22 or newer |
| Git | Latest |
| (Optional) Docker | For running Ollama locally |

### Building

```bash
# Build the CLI binary
go build -o bin/rigel ./cmd/rigel
```

The binary will be placed in `bin/rigel`.

### Running

```bash
# Start the interactive agent REPL
./bin/rigel

# Run a single command
./bin/rigel run --name "list_files" --args "./"
```

A default `config.yaml` is expected in the current working directory. Example config:

```yaml
llm:
  provider: ollama
  model: llama2
  endpoint: http://localhost:11434/api/generate

sandbox:
  timeout: 30s
```

### Testing

```bash
# Run all unit tests
go test ./...

# Run only the examples tests
go test ./examples/...
```

Test files are located under `*_test.go`. The repository contains 13 test files covering agent logic, command execution, and LLM interactions.

### Contributing

1. Fork the repository.
2. Create a new branch with a descriptive name.
3. Add tests that cover your changes.
4. Follow Go code style guidelines (go fmt, go vet).
5. Submit a pull request.
6. Ensure CI passes before merging (GitHub Actions are configured).

### Linting & Formatting

```bash
# Format all files
go fmt ./...

# Lint with golangci-lint
golangci-lint run ./...
```

---

## 5. Architecture Overview

```
+-----------------------------+
|          CLI (cmd/rigel)    |
+-----------------------------+
            |
            v
+-----------------------------+
|        AgentEngine          |
|  (initializes Agent, LLM,   |
|  sandbox, command registry)|
+-----------------------------+
            |
            v
+-----------------------------+
|           Agent             |
|  (memory, planner,          |
|  executor, state)           |
+-----------------------------+
     /            \
    v              v
+-----+        +------------+
| LLM |        |  Sandbox   |
+-----+        +------------+
    ^              |
    |              v
+-----------------------------+
|  External Providers        |
|  (Anthropic, Ollama, etc.) |
+-----------------------------+

```

* **CLI** – parses user input and forwards commands to the **AgentEngine**.
* **AgentEngine** – bootstrapper: loads config, sets up LLM provider, sandbox, and registers commands.
* **Agent** – core stateful object that keeps track of the conversation, variables, and the plan.
* **LLM** – an abstraction that any provider must satisfy (`Provider` interface).
* **Sandbox** – isolates execution of commands that touch the filesystem or run binaries.
* **Commands** – registered in `internal/command/commands.go`; each has a signature, handler, and optional help text.
* **Tools** – reusable utilities (e.g., `code_tool.go`) that can be invoked as commands or LLM‑generated code snippets.

---

## 6. Additional Useful Information

### Command Pattern

Every command follows the pattern:

```go
type Command struct {
    Name        string
    Description string
    Args        []Argument
    Handler     func(ctx *CommandContext) error
}
```

Handlers receive a `CommandContext` that provides:

- `Agent`: the current agent instance
- `Sandbox`: sandbox to run commands
- `Config`: current configuration
- `Logger`: structured logging

### Extending with New Commands

1. Add a new file under `internal/command/` or `internal/tools/`.
2. Define a `Command` struct and its handler.
3. Register it in `commands.go` via `RegisterCommand`.
4. Write a test under `*_test.go` to exercise the new command.

### LLM Providers

- **Anthropic** (`internal/llm/anthropic.go`) – interacts with Anthropic’s API.
- **Ollama** (`internal/llm/ollama.go`) – works with local Ollama models.
- **Custom Providers** – implement `Provider` interface in a new file under `internal/llm/`.

### Sandboxing

Sandbox uses `os/exec` with `Timeout` and `Stdin/Stderr` redirection. It can also be configured to run commands as a non‑root user.

### Agent History

History is persisted in memory per session but can be serialized to JSON if needed. The `history/history.go` package provides utilities to load/store history for debugging or replay.

### Configuration Schema

```yaml
llm:
  provider: string   # "anthropic" or "ollama"
  model: string
  endpoint: string   # optional, defaults per provider
sandbox:
  timeout: string    # duration string (e.g., "30s")
  workdir: string    # default working directory
commands:
  - name: string
    help: string
    # Optional: extra config per command
```

---

### Known Issues / TODOs

| Item | Status |
|------|--------|
| Automatic context extraction for large repos | ✅ |
| WebSocket streaming for LLM responses | 🔧 |
| Integration tests for sandbox isolation | ⏳ |
| Dynamic command loading from plugins | 🔧 |

Feel free to open an issue or pull request to help close any of these tickets.

---

## 7. Quick Start

```bash
# 1. Build the CLI
go build -o bin/rigel ./cmd/rigel

# 2. Create a minimal config (config.yaml)
cat <<EOF > config.yaml
llm:
  provider: ollama
  model: llama2
  endpoint: http://localhost:11434/api/generate
sandbox:
  timeout: 30s
EOF

# 3. Run the REPL
./bin/rigel
```

Once inside the REPL, type `help` to see available commands. You can now experiment with the built‑in tools or create your own agents.

---

### Acknowledgements

- The **Ollama** and **Anthropic** LLM wrappers were inspired by the official SDKs.
- The sandbox implementation borrows patterns from Docker‑in‑Docker minimal setups.
- The project structure follows the [Go project layout guidelines](https://github.com/golang-standards/project-layout).

Happy coding! 🚀
