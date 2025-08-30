# AGENTS.md

> **⚠️ Note** – This file is auto‑generated from the current repository snapshot and is meant to aid automated agents and developers in quickly understanding the codebase. It may not be exhaustive; for deeper dives consult the source files and the documentation in the repository.

---

## 1. Repository Overview

| Item | Detail |
|------|--------|
| **Project name** | **Rigel** |
| **Type** | Command‑line LLM‑powered tool for code exploration and debugging |
| **Purpose** | Provide a lightweight, pluggable framework for interacting with large language models (LLMs) via a rich set of tools (file browsing, code analysis, TUI). It is designed to be extended with custom tools, LLM providers, and UI back‑ends. |

Rigel sits at the intersection of two key concepts:

1. **LLM‑driven assistance** – Leverage modern LLMs (Anthropic, Ollama, etc.) to generate code, answer questions, or analyze codebases.
2. **Tool‑based agent architecture** – Expose a set of small, deterministic tools that the LLM can invoke to get reliable data (e.g., read file contents, analyze code complexity).

The command‑line binary (`cmd/rigel`) offers a simple “agent” that you can run in your terminal to query or manipulate your local repository.

---

## 2. Main Components

| Package | Sub‑directory | Purpose |
|---------|---------------|---------|
| **cmd** | `cmd/rigel` | Entry‑point. Parses CLI flags, initializes the agent stack, and starts the TUI or CLI mode. |
| **internal/agent** | `internal/agent` | Core agent loop, orchestrates prompts, tool selection, and result aggregation. |
| **internal/config** | `internal/config` | Holds configuration structs (LLM settings, tool list, UI mode) and parses config files/flags. |
| **internal/llm** | `internal/llm` | LLM abstraction layer. Includes implementations for Anthropic (`anthropic.go`) and Ollama (`ollama.go`). `provider.go` defines the common interface. |
| **internal/tools** | `internal/tools` | Collection of small utilities that can be invoked by the agent. Each implements the `Tool` interface. |
| **internal/tui** | `internal/tui` | Text‑user‑interface modules. `analyzer.go` provides a code‑analysis view, while `simple.go` is a minimal UI mode. |
| **examples** | `examples/` | Example code snippets and usage demonstrations. |
| **bin** | `bin/` | Optional binary artifacts or scripts (currently empty). |

> **Why this structure?**
> The `internal/` package keeps the core logic private to the repository, enforcing encapsulation while still allowing unit tests to import sub‑packages. The `cmd/` folder remains the only public API, simplifying binary distribution.

---

## 3. Key Files

| File | Path | Why It Matters |
|------|------|----------------|
| **Agent orchestrator** | `internal/agent/agent.go` | Implements the main loop that feeds LLM prompts, selects tools, and composes the final answer. |
| **CLI entry point** | `cmd/rigel/main.go` | Parses command‑line flags, loads config, and boots the agent + TUI. |
| **LLM provider abstraction** | `internal/llm/provider.go` | Defines `LLMProvider` interface; enables adding new providers (OpenAI, Claude, etc.) without touching the agent logic. |
| **Anthropic integration** | `internal/llm/anthropic.go` | Handles authentication, request construction, and response parsing for Anthropic models. |
| **Ollama integration** | `internal/llm/ollama.go` | Supports local LLMs via the Ollama API – great for privacy or offline use. |
| **Config loader** | `internal/config/config.go` | Parses YAML/JSON config files, command‑line flags, and sets defaults. |
| **File tool** | `internal/tools/file_tool.go` | Allows the agent to read or write files safely – a critical tool for code analysis. |
| **Code analysis tool** | `internal/tools/code_tool.go` | Wraps simple code‑metrics (lines, complexity) for the LLM to query. |
| **Tool interface** | `internal/tools/tool.go` | Uniform contract for all tools. Facilitates easy addition of new tools. |
| **TUI analyzer view** | `internal/tui/analyzer.go` | Presents code analysis results interactively; uses termui for rich output. |
| **TUI simple mode** | `internal/tui/simple.go` | Fallback minimal interface (text‑only). Useful for CI or headless environments. |

---

## 4. Development Information

### 4.1 Prerequisites

- Go 1.22+ (tested with 1.23)
- Optional: an LLM provider account (Anthropic, OpenAI, etc.) or a local Ollama installation

### 4.2 Building

```bash
# Build the binary
go build -o bin/rigel ./cmd/rigel
```

### 4.3 Running

```bash
# Minimal example using local Ollama
./bin/rigel --llm-provider=ollama --ollama-model=llama3.1

# With Anthropic (set env var or use config file)
export ANTHROPIC_API_KEY="sk-..."
./bin/rigel --llm-provider=anthropic --anthropic-model=claude-3.5-sonnet
```

### 4.4 Testing

```bash
# Run all tests (unit + integration)
go test ./...

# Run tests for a specific package
go test ./internal/agent
```

> **Tip** – Most tests use `httptest` servers to mock LLM responses, so they run offline.

### 4.5 Contributing

1. Fork the repo and clone.
2. Create a feature branch `feat/<feature-name>`.
3. Write tests that cover your change.
4. Run `go test ./...` to ensure all tests pass.
5. Add a short comment in the PR description explaining the feature or bug fix.

All PRs must include unit tests and follow the existing coding style. The linter (`golangci-lint run`) will be automatically executed by GitHub Actions.

### 4.6 CI

- **Linting**: `golangci-lint` on push.
- **Tests**: Run on all Go versions 1.20–1.23.
- **Build**: Generate binaries for Linux, macOS, Windows.

---

## 5. Architecture Overview

```
+-----------------+      +----------------+      +------------------+
|  CLI / TUI      |<---->|  Agent Core    |<---->|   LLM Provider   |
| (cmd/rigel)     |      | (internal/agent)|      | (anthropic.go,   |
+-----------------+      +----------------+      |  ollama.go)      |
          |                     |               +------------------+
          |                     |                           ^
          v                     v                           |
+-----------------+      +----------------+                |
|  Config Loader  |<---->|  Tool Registry |<---------------+
| (internal/config)|      | (internal/tools)|
+-----------------+      +----------------+

```

* **CLI/TUI** – User-facing layer that collects input and displays results.
* **Agent Core** – Decision engine: chooses which tool to call based on LLM outputs, orchestrates conversation, and formats final answers.
* **LLM Provider** – Abstracts API calls; multiple back‑ends supported.
* **Tool Registry** – Holds all available tools; each implements a simple interface, enabling easy extension.
* **Config Loader** – Centralizes configuration from flags, env vars, and config files.

---

## 6. Additional Information for AI Agents

| Topic | Detail |
|-------|--------|
| **Tool Usage Pattern** | The agent emits a tool call in a JSON structure `<tool_name>:<payload>`. The agent core parses it, invokes the corresponding tool, then feeds the tool’s output back into the LLM as a system message. |
| **LLM Prompt Template** | The base prompt includes placeholders for context, the user's query, and tool usage guidelines. It is defined in `internal/agent/agent.go` and can be overridden via a config file. |
| **Extending LLM Providers** | Implement `LLMProvider` interface in `internal/llm/provider.go` and register it via `config` (e.g., `RegisterProvider("myprovider", NewMyProvider)`). |
| **Adding New Tools** | Create a struct that satisfies the `Tool` interface, register it in `internal/tools/tool.go`, and expose a name for the agent to reference. |
| **Testing Strategy** | Tests exercise both the high‑level agent flow and individual tools. Use `httptest` to mock LLM calls; no real API calls in CI. |
| **Error Handling** | All tool errors are surfaced to the LLM as system messages, allowing the LLM to respond appropriately. |
| **Security Notes** | The File Tool respects a whitelist of directories set via config; no arbitrary file access is permitted. |
| **Performance** | The code base is lightweight (< 3500 LOC). Tool invocations are synchronous; for async workloads consider spawning goroutines or caching results. |

---

### 🚀 Quick Start Summary

1. **Clone & Build**
   ```bash
   git clone https://github.com/your-org/rigel.git
   cd rigel
   go build -o bin/rigel ./cmd/rigel
   ```

2. **Run with a local LLM**
   ```bash
   ./bin/rigel --llm-provider=ollama --ollama-model=llama3.1
   ```

3. **Query the codebase**
   In the TUI, type:
   `Explain the purpose of internal/agent/agent.go.`

4. **Extend**
   Add a new tool or LLM provider following the patterns documented above.

---

Happy coding, and may your agents be ever helpful!
