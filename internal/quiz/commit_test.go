package quiz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/locle97/vibecheck/internal/git"
)

func TestGenerateCommitMessage_SendsPromptAndDiff(t *testing.T) {
	var gotPrompt, gotDiff string

	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		gotPrompt = prompt
		gotDiff = diff
		return "feat(auth): add login endpoint", nil
	}})

	files := []git.File{{
		Path: "auth/login.go",
		Hunks: []git.Hunk{{
			Header: "@@ -0,0 +1,5 @@",
			Lines:  []git.Line{{Kind: git.LineAdded, Content: `+func Login() {}`}},
		}},
	}}

	msg, err := g.GenerateCommitMessage(context.Background(), files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg != "feat(auth): add login endpoint" {
		t.Fatalf("unexpected commit message: %q", msg)
	}

	for _, want := range []string{"conventional commits", "type(scope)", "feat", "fix", "refactor"} {
		if !strings.Contains(gotPrompt, want) {
			t.Errorf("prompt should contain %q, got: %q", want, gotPrompt)
		}
	}

	for _, want := range []string{"auth/login.go", "@@ -0,0 +1,5 @@", "Login"} {
		if !strings.Contains(gotDiff, want) {
			t.Errorf("diff should contain %q, got: %q", want, gotDiff)
		}
	}
}

func TestGenerateCommitMessage_TrimsWhitespace(t *testing.T) {
	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		return "  fix(parser): handle empty input\n\nBody line.\n  ", nil
	}})

	msg, err := g.GenerateCommitMessage(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg != "fix(parser): handle empty input\n\nBody line." {
		t.Fatalf("expected trimmed message, got: %q", msg)
	}
}

func TestGenerateCommitMessage_AgentError(t *testing.T) {
	wantErr := errors.New("agent boom")
	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		return "", wantErr
	}})

	_, err := g.GenerateCommitMessage(context.Background(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped agent error, got: %v", err)
	}
}

func TestBuildCommitPrompt_ContainsConventionalCommitsRules(t *testing.T) {
	prompt := buildCommitPrompt()

	for _, want := range []string{
		"conventional commits",
		"type(scope)",
		"72 chars",
		"feat",
		"fix",
		"refactor",
		"No period at end",
		"ONLY the commit message",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("buildCommitPrompt should contain %q", want)
		}
	}
}
