package quiz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/locle97/vibecheck/internal/git"
)

type fakeAgent struct {
	complete func(ctx context.Context, prompt, diff string) (string, error)
}

func (f fakeAgent) Complete(ctx context.Context, prompt, diff string) (string, error) {
	return f.complete(ctx, prompt, diff)
}

func TestGenerateQuestions_SendsPromptAndDiff(t *testing.T) {
	var gotPrompt, gotDiff string

	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		gotPrompt = prompt
		gotDiff = diff
		return `[{"id":1,"question":"What changed?","options":["A","B","C","D"],"answer":0}]`, nil
	}})

	files := []git.File{{
		Path: "main.go",
		Hunks: []git.Hunk{{
			Header: "@@ -1,3 +1,4 @@",
			Lines:  []git.Line{{Kind: git.LineAdded, Content: `+fmt.Println("hello")`}},
		}},
	}}

	questions, err := g.GenerateQuestions(context.Background(), files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(questions) != 1 || questions[0].Question != "What changed?" {
		t.Fatalf("unexpected questions: %+v", questions)
	}

	for _, want := range []string{"JSON", "3-5 overall questions", "one question per diff hunk", "Order questions strictly by file flow"} {
		if !strings.Contains(gotPrompt, want) {
			t.Fatalf("prompt should include %q instructions, got: %q", want, gotPrompt)
		}
	}

	for _, want := range []string{"main.go", "@@ -1,3 +1,4 @@", `fmt.Println`} {
		if !strings.Contains(gotDiff, want) {
			t.Fatalf("diff should contain %q, got: %q", want, gotDiff)
		}
	}
}

func TestGenerateQuestions_AgentError(t *testing.T) {
	wantErr := errors.New("boom")
	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		return "", wantErr
	}})

	_, err := g.GenerateQuestions(context.Background(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped agent error, got: %v", err)
	}
}

func TestParseQuestions_MarkdownFenced(t *testing.T) {
	raw := "```json\n[{\"id\":1,\"question\":\"Why?\",\"options\":[\"A\",\"B\",\"C\",\"D\"],\"answer\":1}]\n```"
	qs, err := parseQuestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(qs) != 1 || qs[0].Question != "Why?" || qs[0].Answer != 1 {
		t.Errorf("unexpected result: %+v", qs)
	}
}

func TestParseQuestions_NoArray(t *testing.T) {
	_, err := parseQuestions("no JSON array here")
	if err == nil {
		t.Fatal("expected error when no array found")
	}
}
