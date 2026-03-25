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

	for _, want := range []string{"JSON", "exactly one question per diff hunk", "Order questions strictly by file and hunk order", "Use id format H1..Hm"} {
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

func TestParseQuestions_EscapedArray(t *testing.T) {
	raw := `[{\"id\":1,\"question\":\"Why?\",\"options\":[\"A\",\"B\",\"C\",\"D\"],\"answer\":1}]`
	qs, err := parseQuestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(qs) != 1 || qs[0].Question != "Why?" || qs[0].Answer != 1 {
		t.Errorf("unexpected result: %+v", qs)
	}
}

func TestGenerateQuestions_AnnotatesTargetHunkFromID(t *testing.T) {
	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		return `[
			{"id":"H1","question":"Hunk one","options":["A","B","C","D"],"answer":0},
			{"id":"H2","question":"Hunk two","options":["A","B","C","D"],"answer":0}
		]`, nil
	}})

	files := []git.File{{
		Path:  "main.go",
		Hunks: []git.Hunk{{Header: "@@ -1,1 +1,1 @@"}, {Header: "@@ -3,1 +3,1 @@"}},
	}}

	questions, err := g.GenerateQuestions(context.Background(), files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(questions) != 2 {
		t.Fatalf("want 2 questions, got %d", len(questions))
	}

	if questions[0].TargetHunkIdx != 1 {
		t.Fatalf("question 0 should target hunk 1, got %+v", questions[0])
	}

	if questions[1].TargetHunkIdx != 2 {
		t.Fatalf("question 1 should target hunk 2, got %+v", questions[1])
	}
}

func TestGenerateQuestions_FallbackAnnotatesByOrder(t *testing.T) {
	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		return `[
			{"id":1,"question":"Hunk one","options":["A","B","C","D"],"answer":0},
			{"id":2,"question":"Hunk two","options":["A","B","C","D"],"answer":0}
		]`, nil
	}})

	files := []git.File{{
		Path:  "main.go",
		Hunks: []git.Hunk{{Header: "@@ -1,1 +1,1 @@"}, {Header: "@@ -3,1 +3,1 @@"}},
	}}

	questions, err := g.GenerateQuestions(context.Background(), files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(questions) != 2 {
		t.Fatalf("want 2 questions, got %d", len(questions))
	}

	if questions[0].TargetHunkIdx != 1 {
		t.Fatalf("question 0 should target hunk 1, got %+v", questions[0])
	}

	if questions[1].TargetHunkIdx != 2 {
		t.Fatalf("question 1 should target hunk 2, got %+v", questions[1])
	}
}

func TestGenerateQuestions_AnnotatesByOrderEvenWhenIDsAreOutOfOrder(t *testing.T) {
	g := New(fakeAgent{complete: func(ctx context.Context, prompt, diff string) (string, error) {
		return `[
			{"id":"H2","question":"First question","options":["A","B","C","D"],"answer":0},
			{"id":"H1","question":"Second question","options":["A","B","C","D"],"answer":0}
		]`, nil
	}})

	files := []git.File{{
		Path:  "main.go",
		Hunks: []git.Hunk{{Header: "@@ -1,1 +1,1 @@"}, {Header: "@@ -3,1 +3,1 @@"}},
	}}

	questions, err := g.GenerateQuestions(context.Background(), files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(questions) != 2 {
		t.Fatalf("want 2 questions, got %d", len(questions))
	}

	if questions[0].TargetHunkIdx != 1 {
		t.Fatalf("question 0 should target hunk 1, got %+v", questions[0])
	}

	if questions[1].TargetHunkIdx != 2 {
		t.Fatalf("question 1 should target hunk 2, got %+v", questions[1])
	}
}
