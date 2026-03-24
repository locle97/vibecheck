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

// QuestionKind indicates whether a question is global or tied to a specific hunk.
type QuestionKind string

const (
	QuestionKindGeneral QuestionKind = "general"
	QuestionKindHunk    QuestionKind = "hunk"
)

// Question is a single multiple-choice quiz question generated from the diff.
type Question struct {
	ID            int          `json:"-"`
	IDLabel       string       `json:"-"`
	Question      string       `json:"question"`
	Options       []string     `json:"options"`
	Answer        int          `json:"answer"`
	Hint          string       `json:"hint,omitempty"`
	Kind          QuestionKind `json:"-"`
	TargetHunkIdx int          `json:"-"` // 1-based index in flattened diff hunk order
}

func (q *Question) UnmarshalJSON(data []byte) error {
	type wireQuestion struct {
		ID       json.RawMessage `json:"id"`
		Question string          `json:"question"`
		Options  []string        `json:"options"`
		Answer   int             `json:"answer"`
		Hint     string          `json:"hint,omitempty"`
	}

	var wire wireQuestion
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	q.Question = wire.Question
	q.Options = wire.Options
	q.Answer = wire.Answer
	q.Hint = wire.Hint

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
	sb.WriteString("Generate multiple-choice quiz questions that test the developer's understanding of exactly what changed and why.\n")
	sb.WriteString("Create 3-5 overall questions about the broader structure/workflow of the change.\n")
	sb.WriteString("Also create exactly one question per diff hunk for hunk-level understanding.\n")
	sb.WriteString("Order questions strictly by file flow: all overall questions first, then hunk-specific questions in the same file and hunk order shown in the diff.\n")
	sb.WriteString("Use id format G1..Gn for overall questions and H1..Hm for hunk-specific questions where Hk maps to the k-th hunk in the rendered diff order.\n")
	sb.WriteString("Each hunk-specific question should clearly anchor to that specific hunk (file path and hunk context).\n")
	sb.WriteString("Each question must have exactly 4 options and one correct answer.\n\n")
	sb.WriteString("Return ONLY a JSON array - no markdown fences, no prose - using this exact shape:\n")
	sb.WriteString(`[{"id":"G1","question":"...","options":["choice A","choice B","choice C","choice D"],"answer":0,"hint":"optional"}]`)
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
var (
	generalIDRe = regexp.MustCompile(`(?i)^g(?:eneral)?[-_ ]?(\d+)?$`)
	hunkIDRe    = regexp.MustCompile(`(?i)^h(?:unk)?[-_ ]?(\d+)$`)
)

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

func annotateQuestions(questions []Question, files []git.File) {
	totalHunks := countHunks(files)

	for i := range questions {
		label := strings.TrimSpace(questions[i].IDLabel)
		if label == "" {
			continue
		}

		if generalIDRe.MatchString(label) {
			questions[i].Kind = QuestionKindGeneral
			continue
		}

		matches := hunkIDRe.FindStringSubmatch(label)
		if len(matches) != 2 {
			continue
		}

		hunkIdx, err := strconv.Atoi(matches[1])
		if err != nil || hunkIdx < 1 || hunkIdx > totalHunks {
			continue
		}

		questions[i].Kind = QuestionKindHunk
		questions[i].TargetHunkIdx = hunkIdx
	}

	if totalHunks == 0 {
		for i := range questions {
			if questions[i].Kind == "" {
				questions[i].Kind = QuestionKindGeneral
			}
		}
		return
	}

	hunkQuestions := make([]int, 0, len(questions))
	for i := range questions {
		if questions[i].Kind == QuestionKindHunk {
			hunkQuestions = append(hunkQuestions, i)
		}
	}

	if len(hunkQuestions) == totalHunks {
		for i := range questions {
			if questions[i].Kind == "" {
				questions[i].Kind = QuestionKindGeneral
			}
		}
		return
	}

	generalCount := len(questions) - totalHunks
	if generalCount < 0 {
		generalCount = 0
	}

	for i := range questions {
		if i < generalCount {
			questions[i].Kind = QuestionKindGeneral
			questions[i].TargetHunkIdx = 0
			continue
		}
		questions[i].Kind = QuestionKindHunk
		questions[i].TargetHunkIdx = i - generalCount + 1
	}
}

func countHunks(files []git.File) int {
	total := 0
	for _, f := range files {
		total += len(f.Hunks)
	}
	return total
}
