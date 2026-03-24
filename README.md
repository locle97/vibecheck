# vibecheck

> A terminal-based pre-commit self-awareness tool for developers.

**vibecheck** is a CLI + TUI application that helps you validate your understanding of staged changes before committing. It quizzes you on your own diff — so you actually know what you're shipping.

---

## Why

Developers often:

- commit code without fully understanding all changes
- overlook edge cases or unintended side effects
- rely on code review to catch basic issues

**vibecheck** introduces a lightweight, local-first habit:

```bash
vibecheck
```

It launches an interactive TUI session that walks you through your staged changes and checks your comprehension with a quiz.

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Language | Go |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| TUI components | [Bubbles](https://github.com/charmbracelet/bubbles) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |

---

## Getting Started

### Prerequisites

- Go 1.21+
- Git

### Install

```bash
go install github.com/locle97/vibecheck@latest
```

### Build from source

```bash
git clone https://github.com/locle97/vibecheck
cd vibecheck
go build -o vibecheck .
```

### Run

```bash
# Inside a git repo with staged changes
git add <files>
vibecheck
```

---

## Usage

```bash
vibecheck              # launch TUI review session
vibecheck version      # show version info
```

### Behavior

- **Inside a git repo with staged changes** — launches the interactive TUI
- **No staged changes** — displays a message and exits gracefully
- **Outside a git repo** — returns an error message

---

## Project Structure

```
vibecheck/
├── cmd/               # Cobra commands (root, version)
├── internal/
│   ├── git/           # git diff parsing
│   └── tui/           # Bubble Tea models and views
├── main.go
└── go.mod
```

---

## Milestones

### Phase 0 — Skeleton Project _(current)_

Set up the foundational project structure with working CLI and TUI scaffolding.

- [ ] Initialize Go module
- [ ] Set up Cobra CLI with `root` and `version` commands
- [ ] Stub Bubble Tea app with a basic model/update/view loop
- [ ] Wire `vibecheck` command to launch the TUI stub
- [ ] Add `git diff --cached` integration (read staged changes)
- [ ] Graceful exit when no staged changes detected

**Exit criteria:** `vibecheck` launches, detects staged changes, renders a placeholder TUI screen, and exits cleanly.

---

### Phase 1 — Quiz Flow _(planned)_

Implement the core comprehension quiz driven by staged diff hunks.

- [ ] Parse diff into logical hunks
- [ ] Generate questions per hunk
- [ ] Build quiz TUI (question display, answer input, scoring)
- [ ] Show pass/fail summary at end of session

---

### Phase 2 — LLM-Powered Questions _(future)_

- [ ] Integrate LLM to generate adaptive, context-aware questions
- [ ] Support local model (Ollama) and remote API options

---

### Phase 3 — Git Hook Integration _(future)_

- [ ] Optional `vibecheck install-hook` to wire into `pre-commit`
- [ ] Block commit on quiz failure (opt-in)

---

## Non-Goals (v0.1)

- Git hook automation
- Commit blocking
- Team dashboards / analytics
- Cloud sync
- File filtering / partial review

---

## Key Characteristics

- **Manual-first** — no git hook integration yet
- **Local-first** — runs entirely on your machine
- **Editor-agnostic** — works in any terminal
- **Single binary** — no external runtime dependencies
