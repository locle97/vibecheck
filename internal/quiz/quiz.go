package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/locle97/vibecheck/internal/agent"
	"github.com/locle97/vibecheck/internal/git"
)

// Question is a single multiple-choice quiz question generated from the diff.
type Question struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   int      `json:"answer"`
	Hint     string   `json:"hint,omitempty"`
}

// Generator builds quiz prompts from staged diffs and parses returned questions.
type Generator struct {
	agent agent.Agent
}

func New(a agent.Agent) *Generator {
	return &Generator{agent: a}
}

func (g *Generator) GenerateQuestions(ctx context.Context, files []git.File) ([]Question, error) {
	raw, err := g.agent.Complete(ctx, buildPrompt(), renderDiff(files))
	if err != nil {
		return nil, err
	}

	questions, err := parseQuestions(raw)
	if err != nil {
		return nil, fmt.Errorf("quiz: parse questions: %w", err)
	}

	return questions, nil
}

func buildPrompt() string {
	var sb strings.Builder
	sb.WriteString("You are reviewing a staged git diff before the developer commits.\n")
	sb.WriteString("Generate multiple-choice quiz questions that test the developer's understanding of exactly what changed and why.\n")
	sb.WriteString("Create 3-5 overall questions about the broader structure/workflow of the change.\n")
	sb.WriteString("Also create exactly one question per diff hunk for hunk-level understanding.\n")
	sb.WriteString("Order questions strictly by file flow: all overall questions first, then hunk-specific questions in the same file and hunk order shown in the diff.\n")
	sb.WriteString("Each hunk-specific question should clearly anchor to that specific hunk (file path and hunk context).\n")
	sb.WriteString("Each question must have exactly 4 options and one correct answer.\n\n")
	sb.WriteString("Return ONLY a JSON array - no markdown fences, no prose - using this exact shape:\n")
	sb.WriteString(`[{"id":1,"question":"...","options":["choice A","choice B","choice C","choice D"],"answer":0,"hint":"optional"}]`)
	sb.WriteString("\n\"answer\" is the 0-based index of the correct option.")
	return sb.String()
}

func renderDiff(files []git.File) string {
	var sb strings.Builder
	for _, f := range files {
		fmt.Fprintf(&sb, "=== %s ===\n", f.Path)
		for _, h := range f.Hunks {
			sb.WriteString(h.Header)
			sb.WriteString("\n")
			for _, l := range h.Lines {
				sb.WriteString(l.Content)
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

var jsonArrayRe = regexp.MustCompile(`(?s)\[.*\]`)

func parseQuestions(raw string) ([]Question, error) {
	raw = strings.TrimSpace(raw)
	fmt.Print("Raw agent output:\n", raw, "\n---\n")
	if strings.HasPrefix(raw, "```") {
		if idx := strings.Index(raw, "\n"); idx != -1 {
			raw = raw[idx+1:]
		}
		raw = strings.TrimSuffix(strings.TrimSpace(raw), "```")
	}

	match := jsonArrayRe.FindString(raw)
	if match == "" {
		return nil, fmt.Errorf("no JSON array in response")
	}

	var questions []Question
	if err := json.Unmarshal([]byte(match), &questions); err != nil {
		unescaped, unescapeErr := strconv.Unquote("\"" + match + "\"")
		if unescapeErr != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(unescaped), &questions); err != nil {
			return nil, err
		}
	}

	return questions, nil
}
