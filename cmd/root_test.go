package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/locle97/vibecheck/config"
	"github.com/locle97/vibecheck/internal/agent"
	"github.com/locle97/vibecheck/internal/git"
	"github.com/locle97/vibecheck/internal/quiz"
)

type fakeAgent struct {
	complete func(ctx context.Context, prompt, diff string) (string, error)
}

func (f fakeAgent) Complete(ctx context.Context, prompt, diff string) (string, error) {
	return f.complete(ctx, prompt, diff)
}

func TestRootCommand_ExecutesFullFlow(t *testing.T) {
	var buf bytes.Buffer
	deps := rootDeps{
		loadConfig: func(path string) (config.Config, error) {
			return config.Config{Agent: config.AgentConfig{Provider: config.ProviderClaude, Model: "claude-opus-4-6"}}, nil
		},
		parseStagedDiff: func() ([]git.File, error) {
			return []git.File{{
				Path: "cmd/root.go",
				Hunks: []git.Hunk{{
					Header: "@@ -1,1 +1,1 @@",
					Lines:  []git.Line{{Kind: git.LineAdded, Content: "+hello"}},
				}},
			}}, nil
		},
		newAgent: func(binary, model string) (agent.Agent, error) {
			return fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
				return `[{"id":1,"question":"What changed?","options":["A","B","C","D"],"answer":2}]`, nil
			}}, nil
		},
		newGenerator: quiz.New,
	}

	root := newRootCmd(&buf, deps)
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := buf.String()
	for _, want := range []string{"Provider: claude", "What changed?", "Answer: C. C"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output should contain %q, got: %q", want, got)
		}
	}
}

func TestRootCommand_NoStagedChanges(t *testing.T) {
	var buf bytes.Buffer
	deps := rootDeps{
		loadConfig: func(path string) (config.Config, error) {
			return config.Config{}, nil
		},
		parseStagedDiff: func() ([]git.File, error) {
			return nil, nil
		},
		newAgent: func(binary, model string) (agent.Agent, error) {
			t.Fatal("newAgent should not be called when no staged changes")
			return nil, nil
		},
		newGenerator: quiz.New,
	}

	root := newRootCmd(&buf, deps)
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), "No staged changes found") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestRootCommand_ParseDiffError(t *testing.T) {
	var buf bytes.Buffer
	wantErr := errors.New("not a git repo")
	deps := rootDeps{
		loadConfig: func(path string) (config.Config, error) {
			return config.Config{}, nil
		},
		parseStagedDiff: func() ([]git.File, error) {
			return nil, wantErr
		},
		newAgent:     func(binary, model string) (agent.Agent, error) { return nil, nil },
		newGenerator: quiz.New,
	}

	root := newRootCmd(&buf, deps)
	err := root.Execute()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error wrapping parse diff error, got: %v", err)
	}
}
