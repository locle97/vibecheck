# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**vibecheck** is a Go CLI + TUI tool that quizzes developers on their staged git diff before committing. It parses `git diff --cached`, walks through hunks interactively via a Bubble Tea TUI, and checks comprehension through a coding agent (Claude) that generates contextual questions about the diff.

## Commands

```bash
# Build
go build -o vibecheck .

# Run
./vibecheck          # launch TUI review session (requires staged changes)

# Test (all)
go test ./...

# Test (single package)
go test ./internal/git/...

# Test (single test function)
go test ./internal/git/... -run TestParseDiff

# Lint
go vet ./...
```

## Architecture

```
vibecheck/
├── main.go                  # entry point — calls cmd.Execute()
├── cmd/
│   └── root.go              # root command — single entry point, no subcommands
├── internal/
│   ├── git/
│   │   └── diff.go          # parse staged hunks from `git diff --cached`
│   └── agent/
│       └── client.go        # coding agent interface (generates questions from diff)
├── tui/
│   ├── app.go               # Bubble Tea root model, phase router
│   ├── diff_view.go         # syntax-highlighted diff panel
│   └── quiz.go              # quiz view (question display, answer input, scoring)
└── config/
    └── config.go            # ~/.config/vibecheck/config.toml loader
```

## Key Conventions

- **TDD**: write tests first, then implement. All `internal/` packages need `_test.go` covering the public API before the implementation is written.
- **Single root command**: `vibecheck` runs directly with no subcommands. The root `RunE` owns the full review flow.
- **Bubble Tea**: strict Elm architecture — `Model` is immutable, return new copies from `Update`, side effects via `tea.Cmd`. `tui/app.go` owns the phase router and dispatches to sub-models.
- **Coding agent**: `internal/agent/client.go` defines the provider-agnostic interface for question generation. Supported providers: `claude`, `cursor`, `opencode`. Callers never depend on a specific provider or model directly.
- **`internal/git`** must not import TUI or agent packages. `tui/` may import `internal/git` for diff types.
- **Config**: loaded once at startup from `~/.config/vibecheck/config.toml`; passed down via dependency injection, not global state.
