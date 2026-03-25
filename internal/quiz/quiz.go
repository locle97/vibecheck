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
	ID            int      `json:"-"`
	IDLabel       string   `json:"-"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	Answer        int      `json:"answer"`
	Hint          string   `json:"hint,omitempty"`
	Explanation   string   `json:"explanation,omitempty"`
	TargetHunkIdx int      `json:"-"` // 1-based index in flattened diff hunk order
}

func (q *Question) UnmarshalJSON(data []byte) error {
	type wireQuestion struct {
		ID          json.RawMessage `json:"id"`
		Question    string          `json:"question"`
		Options     []string        `json:"options"`
		Answer      int             `json:"answer"`
		Hint        string          `json:"hint,omitempty"`
		Explanation string          `json:"explanation,omitempty"`
	}

	var wire wireQuestion
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	q.Question = wire.Question
	q.Options = wire.Options
	q.Answer = wire.Answer
	q.Hint = wire.Hint
	q.Explanation = wire.Explanation

	if len(wire.ID) == 0 {
		return nil
	}

	var asInt int
	if err := json.Unmarshal(wire.ID, &asInt); err == nil {
		q.ID = asInt
		q.IDLabel = strconv.Itoa(asInt)
		return nil
	}

	var asString string
	if err := json.Unmarshal(wire.ID, &asString); err != nil {
		return fmt.Errorf("question id must be number or string: %w", err)
	}

	asString = strings.TrimSpace(asString)
	q.IDLabel = asString
	if parsed, err := strconv.Atoi(asString); err == nil {
		q.ID = parsed
	}

	return nil
}

func (q Question) DisplayID() string {
	if q.IDLabel != "" {
		return q.IDLabel
	}
	if q.ID > 0 {
		return strconv.Itoa(q.ID)
	}
	return "?"
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

	annotateQuestions(questions, files)

	return questions, nil
}

func buildPrompt() string {
	var sb strings.Builder
	sb.WriteString("You are reviewing a staged git diff before the developer commits.\n")
	sb.WriteString("Generate simple, straightforward multiple-choice questions that verify the developer read and understood the changes.\n")
	sb.WriteString("Questions should be factual and directly answerable from the diff — avoid trick questions, ambiguous wording, or testing obscure edge cases.\n")
	sb.WriteString("Create exactly one question per diff hunk.\n")
	sb.WriteString("Order questions strictly by file and hunk order shown in the diff.\n")
	sb.WriteString("Use id format H1..Hm where Hk maps to the k-th hunk in the rendered diff order.\n")
	sb.WriteString("Each question should clearly anchor to its specific hunk (file path and hunk context).\n")
	sb.WriteString("Each question must have exactly 4 options and one correct answer.\n")
	sb.WriteString("For each question, include an \"explanation\" field: a brief, clear explanation of why the correct answer is right, shown to the developer when they answer incorrectly.\n\n")
	sb.WriteString("Return ONLY a JSON array - no markdown fences, no prose - using this exact shape:\n")
	sb.WriteString(`[{"id":"H1","question":"...","options":["choice A","choice B","choice C","choice D"],"answer":0,"hint":"optional","explanation":"why the correct answer is right"}]`)
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

func annotateQuestions(questions []Question, files []git.File) {
	totalHunks := countHunks(files)
	if totalHunks == 0 {
		return
	}

	limit := len(questions)
	if limit > totalHunks {
		limit = totalHunks
	}

	for i := 0; i < limit; i++ {
		questions[i].TargetHunkIdx = i + 1
	}
}

func countHunks(files []git.File) int {
	total := 0
	for _, f := range files {
		total += len(f.Hunks)
	}
	return total
}
