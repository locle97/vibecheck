# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**vibecheck** is a Go CLI + TUI tool that quizzes developers on their staged git diff before committing. It parses `git diff --cached`, walks through hunks interactively via a Bubble Tea TUI, and checks comprehension through annotation, Socratic Q&A, and quiz phases powered by an LLM backend.

## Commands

```bash
# Build
go build -o vibecheck .

# Run
./vibecheck review          # manual TUI review session (requires staged changes)
./vibecheck commit          # wraps git commit with vibecheck review gate
./vibecheck install         # sets up git pre-commit hook
./vibecheck version         # show version info

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
├── cmd/                     # Cobra commands
│   ├── root.go              # root command, global flags
│   ├── commit.go            # vibecheck commit — wraps git commit
│   ├── review.go            # vibecheck review — manual trigger
│   └── install.go           # vibecheck install — sets up git hook
├── internal/
│   ├── git/
│   │   ├── diff.go          # parse staged hunks from `git diff --cached`
│   │   └── hook.go          # hook installer / uninstaller
│   ├── session/
│   │   ├── session.go       # session struct and state machine
│   │   ├── local.go         # .vibecheck/<hash>.json (per-repo persistence)
│   │   └── global.go        # ~/.vibecheck/history.jsonl (cross-repo log)
│   ├── llm/
│   │   ├── client.go        # backend interface (Client, GenerateQuestion, etc.)
│   │   ├── ollama.go        # Ollama backend
│   │   ├── openai.go        # OpenAI backend
│   │   └── anthropic.go     # Anthropic backend
│   └── phases/
│       ├── annotation.go    # annotation phase logic
│       ├── socratic.go      # Socratic Q&A phase logic
│       └── quiz.go          # quiz phase logic (MCQ + short answer)
├── tui/
│   ├── app.go               # Bubble Tea root model, phase router
│   ├── diff_view.go         # syntax-highlighted diff panel
│   ├── annotation.go        # annotation input view
│   ├── socratic.go          # Q&A conversation view
│   ├── quiz.go              # quiz view
│   └── summary.go           # post-review result summary
└── config/
    └── config.go            # ~/.config/vibecheck/config.toml loader
```

## Key Conventions

- **TDD**: write tests first, then implement. All `internal/` packages need `_test.go` covering the public API before the implementation is written.
- **Cobra**: one command per file in `cmd/`, registered via `init()`.
- **Bubble Tea**: strict Elm architecture — `Model` is immutable, return new copies from `Update`, side effects via `tea.Cmd`. `tui/app.go` owns the phase router and dispatches to sub-models.
- **LLM backends**: all backends implement the `client.go` interface; callers never import a specific backend directly.
- **`internal/git`** must not import TUI or LLM packages. `tui/` and `phases/` may import `internal/git` for diff types.
- **Config**: loaded once at startup from `~/.config/vibecheck/config.toml`; passed down via dependency injection, not global state.
