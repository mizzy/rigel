# Expect-based UI Tests

This directory contains interactive UI tests written with `expect` that drive the termflow REPL via a PTY. They are useful for verifying cursor behavior, welcome text spacing, and multiline input rendering.

How to run
- Build binaries first: `go build -o bin/rigel ./cmd/rigel` and ensure helper binaries (e.g., `rigel-clean`, `rigel-final`, etc.) are built/present in repo root.
- From the repository root, run a specific test:
  - `expect -f tests/expect/test_backspace.exp`
  - or `expect -d -f tests/expect/test_backspace_debug.exp` for debug output
- To run all: `bash tests/expect/run_all.sh`

Notes
- Tests set `RIGEL_TEST_MODE=1` and `PROVIDER=ollama` to avoid external dependencies and to unify terminal behavior.
- Paths are resolved relative to this directory, so tests can be run from anywhere.

