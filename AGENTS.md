# AGENTS.md

---

## 1. Repository Overview

**Rigel – AI Coding Agent**
Rigel is a lightweight, open‑source AI agent written in Go that can *write, read, edit and execute* code on a local machine.
It is designed to be run from the command line, but it also ships a fully‑functional terminal UI that allows you to:

- Ask the agent to implement or fix a piece of code
- Inspect the agent’s plan and the current state of the workspace
- Step through the agent’s actions and see the results in real time

The project is intentionally modular so you can drop in your own LLM provider or tools, or replace the UI with a web or desktop frontend.

---

## 2. Main Components

| Package | Purpose |
|---------|---------|
| `cmd/rigel` | The CLI entry point (`main.go`) – loads configuration, creates an `Agent` instance and launches the interactive loop or the TUI. |
| `internal/agent` | Core business logic. An `Agent` owns a *plan executor*, a *tool executor* and a *LLM provider*. |
| `internal/config` | Declarative JSON/YAML config parser. Describes which LLM provider to use, API keys, tool set, and agent defaults. |
| `internal/llm` | LLM abstraction layer. `Provider` interface + concrete implementations (`anthropic.go`, `ollama.go`). |
| `internal/tools` | Collection of *tools* that the agent can call. Each tool implements `Tool` interface – currently a *code editor* and a *file reader*. |
| `internal/tui` | Terminal UI (TUI). Uses `tview` to render the agent’s state, chat log, tool outputs and a syntax‑highlighted editor. |
| `examples` | Sample projects that demonstrate how to let Rigel auto‑generate/modify code. |

---

## 3. Key Files

| File | Why it matters |
|------|----------------|
| `cmd/rigel/main.go` | Starts the application: loads config, builds an `Agent`, selects UI mode, and begins the event loop. |
| `internal/agent/agent.go` | Exposes `NewAgent(cfg *config.Config)` and orchestrates the *LLM → Plan → Tool* pipeline. |
| `internal/config/config.go` | Parses `config.yaml` (or `$RIGEL_CONFIG_PATH`). Validates provider settings and tool configuration. |
| `internal/llm/provider.go` | Declares the `Provider` interface – `Chat(messages []Message) (string, error)`. |
| `internal/llm/anthropic.go` | Implements the Anthropic Claude API client. |
| `internal/llm/ollama.go` | Implements the local Ollama client (`POST /api/chat`). |
| `internal/tools/tool.go` | Declares the `Tool` interface – `Name()`, `Description()`, `Run(args string) (string, error)`. |
| `internal/tools/code_tool.go` | Reads, edits and writes code files; also can compile / run tests. |
| `internal/tools/file_tool.go` | Provides read-only file content. |
| `internal/tui/model.go` | Holds the UI state – messages, tool outputs, editor content. |
| `internal/tui/components.go` | UI widgets (chat pane, tool pane, editor). |
| `internal/tui/theme.go` | Light / dark color schemes for `tview`. |

---

## 4. Architecture

```
CLI (cmd/rigel) ──┐
                 │
          ┌───────┴───────────────┐
          │  Agent (internal/agent)│
          │  ┌───────────────┐     │
          │  │  LLM Provider │     │
          │  │ (anthropic.go │     │
          │  │  ollama.go)   │     │
          │  └───────┬───────┘     │
          │          │             │
          │  ┌───────┴───────┐     │
          │  │  Tool Set     │     │
          │  │  (code_tool, │     │
          │  │   file_tool) │     │
          │  └───────┬───────┘     │
          │          │             │
          └───────┬──┴─────────────┘
                  │
      ┌───────────┴─────────────┐
      │  Terminal UI (tview)    │
      └─────────────────────────┘
```

**Workflow**

1. **Prompt** – User types a request into the chat pane.
2. **LLM** – The request is sent to the chosen LLM provider; the model outputs a *plan* and optional *tool calls*.
3. **Planner** – Agent parses the plan, decides which tool(s) to invoke.
4. **Tools** – Each tool receives arguments and returns a string (or error).
5. **Re‑invoke LLM** – The tool outputs are fed back to the LLM, which may generate updated code or additional actions.
6. **Apply** – If the plan calls for code changes, the `code_tool` writes to disk and optionally runs tests.
7. **Display** – All messages, tool outputs and the latest file contents are rendered in the TUI.

---

## 5. Technologies Used

| Category | Library / Tool | Reason |
|----------|----------------|--------|
| **Language** | Go 1.20+ | Strong static typing, easy concurrency, great tooling. |
| **CLI** | `cobra` (indirect) | Structured command handling and flags. |
| **Configuration** | `encoding/json` & `gopkg.in/yaml.v3` | Human‑readable YAML config with optional JSON overrides. |
| **LLM Providers** | HTTP client (`net/http`) | Custom implementation for Anthropic & Ollama REST APIs. |
| **Terminal UI** | `github.com/rivo/tview` | Rich, cross‑platform TUI with syntax highlighting and focus management. |
| **Testing** | `testing`, `httptest` | Unit and integration tests for LLM providers, tools, and agent orchestration. |
| **Build** | `go build`, `go test` | Standard Go toolchain. |
| **Linting / Formatting** | `golangci-lint`, `go fmt` | Maintains code quality. |

---

## 6. Getting Started

### 1. Prerequisites

| Requirement | Install |
|-------------|---------|
| Go 1.20+ | `brew install go` or from <https://go.dev/dl/> |
| Ollama (optional) | `docker pull ollama/ollama && docker run -p 11434:11434 ollama/ollama` |
| Anthropic API key | Set `ANTHROPIC_API_KEY` env var |

> **Tip** – If you only want to try the CLI, the bundled `examples/simple` project will work with either Ollama or Anthropic.

### 2. Clone & Build

```bash
git clone https://github.com/yourorg/feat-interactive.git
cd feat-interactive
go build ./cmd/rigel
```

### 3. Create a Config

```yaml
# config.yaml
llm:
  provider: "anthropic"     # or "ollama"
  model: "claude-3-5-sonnet-20240620"
  temperature: 0.2
  anthropic_api_key: "${ANTHROPIC_API_KEY}"   # can be env var reference

tools:
  - name: "code_tool"
    description: "Edit code files and optionally run tests."
  - name: "file_tool"
    description: "Read file contents."

agent:
  max_iterations: 10
```

> Place the file at `config.yaml` or export `RIGEL_CONFIG_PATH=/path/to/config.yaml`.

### 4. Run Rigel

```bash
# Interactive chat + TUI
./rigel
```

You’ll see a split view: left is the chat log, right is a syntax‑highlighted editor.
Type a request such as:

```
Add a unit test for the `Sum` function in `math.go`.
```

The agent will plan, edit the file, run `go test`, and show you the diff and test results.

### 5. Run Tests

```bash
go test ./...
```

All packages are covered by unit tests. Integration tests use `httptest` servers to mock LLM providers.

---

## 7. Development Guidelines

| Guideline | What to follow |
|-----------|----------------|
| **Interface‑First** | Use the `Provider` and `Tool` interfaces to write plug‑in components. Add new tools by creating a struct that implements `Name()`, `Description()`, `Run(args string)`. |
| **JSON/YAML Config** | Keep configuration declarative. Add new fields to `config.Config` and update validation logic. Avoid hard‑coding values. |
| **Error Handling** | Wrap LLM errors with context (`fmt.Errorf("LLM request failed: %w", err)`). Never swallow errors – they must surface to the UI. |
| **Testing** | Prefer table‑driven tests. Use `httptest.NewServer` for LLM provider tests. Mock file I/O when possible (`io/fs`). |
| **TUI Updates** | All UI updates must happen on the main goroutine. Use `tview.Application.QueueUpdateDraw` when updating from goroutines. |
| **Command Flags** | Keep CLI flags minimal. Expose a `--config` flag for custom config path. |
| **Documentation** | Add comments to exported types/methods. `go doc` should be useful. |
| **Linting** | Run `golangci-lint run` before PRs. |
| **Security** | Never log raw API keys. Redact sensitive fields in logs. |

---

### Quick Reference

| Command | Description |
|---------|-------------|
| `./rigel` | Launch TUI. |
| `./rigel --config=/path/to/config.yaml` | Override config path. |
| `go run cmd/rigel/main.go` | Run without building. |
| `go test -run TestAgent` | Run agent unit tests. |

Feel free to open issues or PRs if you discover bugs or want to add new features – Rigel is intentionally minimal and extensible!

---

*This file was automatically generated by Rigel AI to help AI agents understand this codebase.*
