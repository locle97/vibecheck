package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/locle97/vibecheck/config"
	"github.com/locle97/vibecheck/internal/agent"
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
	var tuiCalled bool
	deps := rootDeps{
		loadConfig: func(path string) (config.Config, error) {
			return config.Config{Agent: config.AgentConfig{Provider: config.ProviderClaude, Model: "claude-opus-4-6"}}, nil
		},
		newAgent: func(binary, model string) (agent.Agent, error) {
			return fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
				return `[{"id":1,"question":"What changed?","options":["A","B","C","D"],"answer":2}]`, nil
			}}, nil
		},
		newGenerator: quiz.New,
		runTUI: func(gen *quiz.Generator, cfg config.Config) error {
			tuiCalled = true
			return nil
		},
	}

	root := newRootCmd(&buf, deps)
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !tuiCalled {
		t.Fatal("runTUI was not called")
	}
}
