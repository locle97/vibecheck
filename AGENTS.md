# AGENTS.md

Guide for coding agents operating in `vibecheck`.

## Scope

- Project: Go CLI for reviewing staged git diffs and generating comprehension quiz questions.
- Module: `github.com/locle97/vibecheck`.
- Primary packages: `cmd`, `config`, `internal/git`, `internal/agent`, `internal/quiz`.
- Current UX is CLI output (no checked-in `tui/` package yet).
- `internal/quiz` now annotates questions as general vs hunk-targeted (`QuestionKind`, `TargetHunkIdx`) for future split-pane TUI work.
- Keep edits minimal, targeted, and consistent with existing package boundaries.

## Build / Run / Lint / Test

Run all commands from repo root: `/home/locle97/coding/github/vibecheck`.

### Build

```bash
go build -o vibecheck .
```

### Run

```bash
./vibecheck
```

### Format

```bash
gofmt -w .
```

### Static checks

```bash
go vet ./...
```

### All tests

```bash
go test ./...
```

### Single-package tests (important)

```bash
go test ./cmd
go test ./config
go test ./internal/git
go test ./internal/agent
go test ./internal/quiz
```

### Single-test execution (important)

```bash
go test ./cmd -run TestRootCommand_Executes
go test ./config -run TestLoadConfig_FromFile
go test ./internal/git -run TestParseDiff_FileCount
go test ./internal/agent -run TestNew_KnownBinaries
go test ./internal/quiz -run TestGenerateQuestions_SendsPromptAndDiff
```

### Debugging test runs

```bash
go test -v ./internal/agent -run TestUnwrapPrintJSON_EnvelopeResult -count=1
go test -v ./internal/quiz -run TestParseQuestions_MarkdownFenced -count=1
```

## Architecture and Dependency Boundaries

- Keep `main.go` thin; command wiring belongs in `cmd/`.
- `config/` owns config parsing/defaults; avoid spreading config logic elsewhere.
- `internal/git/` should remain focused on diff parsing and avoid UI/agent concerns.
- `internal/agent/` should encapsulate external CLI calls and output normalization.
- `internal/quiz/` should build prompts, render diff context, parse returned JSON, and classify questions for hunk mapping.
- Avoid circular dependencies; prefer small, explicit interfaces.

## Go Code Style

### Imports

- Use standard library imports first, then a blank line, then module imports.
- Keep import order/go formatting handled by `gofmt`.
- Do not keep unused imports.

### Formatting and Structure

- Always run `gofmt` on changed files.
- Prefer short functions with clear responsibilities.
- Prefer early returns to reduce nested conditionals.
- Keep control flow explicit and readable over cleverness.

### Types and APIs

- Use named types for domain concepts (`Provider`, `LineKind`).
- Keep exported APIs stable unless change is intentional and covered by tests.
- Preserve struct tags and wire formats (`json`, `toml`) when modifying structs.
- Keep public type/function comments when behavior is not obvious.

### Naming

- Exported identifiers: clear `CamelCase` nouns/verbs (`ParseDiff`, `GenerateQuestions`).
- Unexported helpers: concise lowerCamel (`buildPrompt`, `renderDiff`, `defaults`).
- Test naming pattern: `Test<Function>_<Scenario>`.
- Use common Go abbreviations only when idiomatic (`ctx`, `cfg`, `err`).

### Error Handling

- Return errors; do not panic in normal execution paths.
- Wrap propagated errors with context using `%w`.
- Keep user/debug messages actionable and specific.
- For subprocess failures, include both stdout and stderr context (existing pattern).
- Preserve graceful defaults where established (for example, missing config file).

### Context and Subprocesses

- Pass `context.Context` into operations that can block or call external processes.
- Use `exec.CommandContext` for external CLI execution.
- Capture stdout/stderr with buffers when error context is needed.
- Keep provider command flags/arguments consistent across changes.

### Parsing and Serialization

- Trim/sanitize raw model output before decoding.
- Keep defensive parsing behavior for envelope/fallback formats.
- Preserve compatibility for question `id` values returned as either number (`1`) or string (`"G1"`, `"H2"`).
- Prefer prompting providers to emit explicit IDs (`G*` for general, `H*` for hunk-specific); keep fallback order-based mapping intact.
- Avoid breaking existing JSON response parsing without corresponding test updates.

## Testing Guidelines

- Any behavior change should include test coverage updates.
- Prefer deterministic unit tests with no network dependencies.
- Use `t.TempDir()` for filesystem-related tests.
- Use `errors.Is` when asserting wrapped errors.
- Write assertions with clear expected vs actual messages.

## Change Hygiene for Agents

- Do not revert unrelated local edits.
- Do not run destructive git commands.
- Modify only files necessary for the requested task.
- Keep docs/comments aligned with behavior after code changes.
- Before handoff, run focused tests for touched packages; run `go test ./...` when feasible.

## Cursor and Copilot Rules

Repository check results:

- `.cursor/rules/` exists, but contains no rule files.
- `.cursorrules` is not present.
- `.github/copilot-instructions.md` is not present.

Practical implication:

- Use this file and repository conventions as the primary guidance.
- If Cursor/Copilot rule files are added later, treat them as higher-priority local rules and update `AGENTS.md`.
