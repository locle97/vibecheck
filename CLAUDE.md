# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**vibecheck** is currently a Go CLI tool that quizzes developers on their staged git diff before committing. It parses `git diff --cached`, sends rendered diff context to a coding agent, and prints multiple-choice comprehension questions in the terminal.

Current implementation status:

- CLI review flow is implemented in `cmd/root.go`.
- Question generation/parsing/classification is implemented in `internal/quiz`.
- No checked-in `tui/` package yet; hunk-targeting metadata exists to support future split-pane TUI work.

## Commands

```bash
# Build
go build -o vibecheck .

# Run
./vibecheck          # run CLI quiz flow (requires staged changes)

# Test (all)
go test ./...

# Test (single package)
go test ./internal/quiz

# Test (single test function)
go test ./internal/quiz -run TestGenerateQuestions_AnnotatesKindsFromID

# Lint
go vet ./...
```

## Architecture

```
vibecheck/
├── main.go                  # entry point — calls cmd.Execute()
├── cmd/
│   └── root.go              # root command — executes full CLI review flow
├── config/
│   └── config.go            # ~/.config/vibecheck/config.toml loader
├── internal/
│   ├── git/
│   │   └── diff.go          # parse staged hunks from `git diff --cached`
│   ├── quiz/
│   │   └── quiz.go          # prompt building, JSON parsing, question classification
│   └── agent/
│       └── client.go        # provider-agnostic agent interface and constructors
```

## Key Conventions

- **TDD**: write tests first, then implement. All `internal/` packages need `_test.go` covering the public API before the implementation is written.
- **Single root command**: `vibecheck` runs directly with no subcommands. The root `RunE` owns the full review flow.
- **Coding agent**: `internal/agent/client.go` defines the provider-agnostic interface for question generation. Supported providers: `claude`, `cursor-agent`, `opencode`. Callers never depend on a specific provider or model directly.
- **Question IDs and classification**: prefer `id` values like `G1`, `G2`, `H1`, `H2` from providers. `internal/quiz` must still handle numeric IDs and preserve fallback order-based mapping.
- **Dependency boundaries**: `internal/git` must not import UI or agent packages; `internal/quiz` may depend on `internal/git` and `internal/agent` but not vice versa.
- **Config**: loaded once at startup from `~/.config/vibecheck/config.toml`; passed down via dependency injection, not global state.
