# vibecheck

![Go Version](https://img.shields.io/badge/Go-1.25.7-00ADD8?logo=go)
![CLI](https://img.shields.io/badge/interface-CLI-informational)

`vibecheck` is a Go CLI that quizzes you on your staged git diff before you commit. It parses `git diff --cached`, sends the rendered change context to a coding-agent CLI, and prints multiple-choice questions to help verify that you understand what changed.

## Project Name and Description

**vibecheck** is a local-first pre-commit comprehension tool for developers.

Primary purpose:

- Reduce low-context commits by forcing a quick understanding check.
- Turn staged diff review into a repeatable CLI workflow.
- Keep the flow editor-agnostic and lightweight.

## Technology Stack

Primary stack (from `go.mod` and repository sources):

| Layer | Technology | Version |
| --- | --- | --- |
| Language | Go | `1.25.7` |
| CLI framework | [Cobra](https://github.com/spf13/cobra) | `v1.10.2` |
| Config parsing | [BurntSushi/toml](https://github.com/BurntSushi/toml) | `v1.6.0` |
| VCS integration | Git CLI (`git diff --cached`) | system-installed |
| Agent providers | Claude CLI, Cursor Agent CLI, OpenCode CLI | system-installed |

## Getting Started

### Prerequisites

- Go `1.25.7+`
- Git
- At least one supported agent CLI available in `PATH`:
  - `claude`
  - `cursor-agent`
  - `opencode`

### Install

```bash
go install github.com/locle97/vibecheck@latest
```

### Build From Source

```bash
git clone https://github.com/locle97/vibecheck
cd vibecheck
go build -o vibecheck .
```

### Configure

`vibecheck` reads `config.toml` in this order:

1. Local `./config.toml` (if present)
2. `~/.config/vibecheck/config.toml`

Example config:

```toml
[agent]
provider = "claude"        # claude | cursor-agent | opencode
model    = "claude-opus-4-6"
```

### Run

```bash
git add <files>
./vibecheck
```

Behavior:

- Inside a git repo with staged changes: generates and prints quiz questions.
- No staged changes: prints guidance and exits.
- Outside a git repo: returns an error from git parsing.

## Project Structure

```text
vibecheck/
├── main.go
├── cmd/
│   ├── root.go
│   └── root_test.go
├── config/
│   ├── config.go
│   └── config_test.go
├── internal/
│   ├── agent/
│   │   ├── client.go
│   │   ├── claude.go
│   │   ├── cursor.go
│   │   ├── opencode.go
│   │   └── client_test.go
│   ├── git/
│   │   ├── diff.go
│   │   └── diff_test.go
│   └── quiz/
│       ├── quiz.go
│       └── quiz_test.go
├── AGENTS.md
└── CLAUDE.md
```

## Key Features

- Parses staged diffs into structured files, hunks, and line kinds.
- Generates multiple-choice comprehension questions via external coding-agent CLIs.
- Supports provider/model selection via TOML config.
- Handles mixed question ID formats (`1`, `"G1"`, `"H2"`) with fallback mapping.
- Annotates questions as global or hunk-targeted for future richer UI use.
- Includes defensive parsing for agent output wrappers/fences/NDJSON-like streams.

## Development Workflow

Typical local workflow:

```bash
gofmt -w .
go vet ./...
go test ./...
go build -o vibecheck .
```

Repository workflow conventions:

- Keep `main.go` thin; place orchestration in `cmd`.
- Preserve package boundaries across `config`, `internal/git`, `internal/agent`, and `internal/quiz`.
- Add or update tests for behavior changes.
- Prefer small, targeted changes and explicit error wrapping.

## Coding Standards

Key conventions used in this repository:

- Use `gofmt` and idiomatic Go import ordering.
- Prefer short functions and early returns.
- Return wrapped errors with context (`%w`) instead of panics.
- Use deterministic tests and clear assertion messages.
- Keep parsing resilient to provider output variations.

Detailed guidance: `AGENTS.md` and `CLAUDE.md`.

## Testing

Test strategy:

- Unit-test focused across all core packages.
- Fake dependencies for command and agent layers.
- Parser tests cover normal and edge cases (empty diff, multi-hunk, wrapped/escaped JSON).

Run tests:

```bash
go test ./...
```

Run package-specific tests:

```bash
go test ./cmd
go test ./config
go test ./internal/git
go test ./internal/agent
go test ./internal/quiz
```

## Contributing

Contributions are welcome. To keep changes aligned with project conventions:

1. Follow package boundaries and keep orchestration in `cmd`.
2. Add tests with any behavior change.
3. Run formatting, vet, and tests before opening a PR.
4. Use existing patterns in `internal/*_test.go` as exemplars.

Supporting docs:

- `AGENTS.md`
- `CLAUDE.md`

## License

No license file is currently present in this repository. Add a `LICENSE` file to define usage terms.
