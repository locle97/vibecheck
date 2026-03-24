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

It launches an interactive TUI session that walks you through your staged changes and checks your comprehension with agent-generated questions.

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Language | Go |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| TUI components | [Bubbles](https://github.com/charmbracelet/bubbles) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Agent | Claude (Anthropic) |

---

## Getting Started

### Prerequisites

- Go 1.21+
- Git
- Anthropic API key (set `ANTHROPIC_API_KEY`)

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
vibecheck    # launch TUI review session
```

### Behavior

- **Inside a git repo with staged changes** — launches the interactive TUI
- **No staged changes** — displays a message and exits gracefully
- **Outside a git repo** — returns an error message

---

## Project Structure

```
vibecheck/
├── cmd/               # Cobra root command
├── internal/
│   ├── git/           # git diff parsing
│   └── agent/         # coding agent interface
├── tui/               # Bubble Tea models and views
├── config/            # config loader
├── main.go
└── go.mod
```

---

## Configuration

Config file: `~/.config/vibecheck/config.toml`

```toml
[agent]
provider = "claude"       # claude | cursor | opencode
model    = "claude-opus-4-6"
```

### Providers & models

| Provider | Example models |
|----------|---------------|
| `claude` | `claude-opus-4-6`, `claude-sonnet-4-6`, `claude-haiku-4-5` |
| `cursor` | `cursor-small` |
| `opencode` | any model supported by opencode |

---

## Milestones

### Phase 0 — Skeleton Project _(current)_

Set up the foundational project structure with working CLI and TUI scaffolding.

- [ ] Initialize Go module
- [ ] Set up Cobra CLI with single root command
- [ ] Stub Bubble Tea app with a basic model/update/view loop
- [ ] Wire `vibecheck` command to launch the TUI stub
- [ ] Add `git diff --cached` integration (read staged changes)
- [ ] Graceful exit when no staged changes detected

**Exit criteria:** `vibecheck` launches, detects staged changes, renders a placeholder TUI screen, and exits cleanly.

---

### Phase 1 — Quiz Flow _(planned)_

Implement the core comprehension quiz driven by staged diff hunks.

- [ ] Parse diff into logical hunks
- [ ] Integrate coding agent to generate questions per hunk
- [ ] Build quiz TUI (question display, answer input, scoring)
- [ ] Show pass/fail summary at end of session

---

### Phase 2 — Git Hook Integration _(future)_

- [ ] Optional `vibecheck install-hook` to wire into `pre-commit`
- [ ] Block commit on quiz failure (opt-in)

---

## Non-Goals (v0.1)

- Commit blocking
- Team dashboards / analytics
- Cloud sync
- File filtering / partial review

---

## Key Characteristics

- **Local-first** — runs entirely on your machine
- **Editor-agnostic** — works in any terminal
- **Single binary** — no external runtime dependencies beyond API key
