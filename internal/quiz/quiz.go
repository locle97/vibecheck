package quiz

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/locle97/vibecheck/internal/agent"
	"github.com/locle97/vibecheck/internal/git"
)

//go:embed prompt.md
var promptTemplate string

// Question is a single multiple-choice quiz question generated from the diff.
type Question struct {
	ID            int      `json:"-"`
	IDLabel       string   `json:"-"`
	HunkID        string   `json:"-"` // hunk_id echoed back by the agent; used for precise hunk mapping
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
		HunkID      string          `json:"hunk_id,omitempty"`
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

	q.HunkID = wire.HunkID
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
	diff, hunkMap := renderDiff(files)
	raw, err := g.agent.Complete(ctx, buildPrompt(), diff)
	if err != nil {
		return nil, err
	}

	questions, err := parseQuestions(raw)
	if err != nil {
		return nil, fmt.Errorf("quiz: parse questions: %w", err)
	}

	annotateQuestions(questions, files, hunkMap)

	return questions, nil
}

func buildPrompt() string {
	return strings.TrimSpace(promptTemplate)
}

// renderDiff renders files as a human-readable diff with [hunk_id: <id>] tags before each hunk.
// It returns the rendered text and a map from hunk ID to 1-based flattened hunk index.
func renderDiff(files []git.File) (string, map[string]int) {
	var sb strings.Builder
	hunkMap := make(map[string]int)
	hunkCount := 0
	for fi, f := range files {
		fmt.Fprintf(&sb, "=== %s ===\n", f.Path)
		for hi, h := range f.Hunks {
			hunkCount++
			hunkID := fmt.Sprintf("hunk_f%d_h%d", fi, hi)
			hunkMap[hunkID] = hunkCount
			fmt.Fprintf(&sb, "[hunk_id: %s]\n", hunkID)
			sb.WriteString(h.Header)
			sb.WriteString("\n")
			for _, l := range h.Lines {
				sb.WriteString(l.Content)
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String(), hunkMap
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

func annotateQuestions(questions []Question, files []git.File, hunkMap map[string]int) {
	totalHunks := countHunks(files)
	if totalHunks == 0 {
		return
	}

	limit := len(questions)
	if limit > totalHunks {
		limit = totalHunks
	}

	for i := 0; i < limit; i++ {
		if idx, ok := hunkMap[questions[i].HunkID]; ok {
			questions[i].TargetHunkIdx = idx
		} else {
			// fallback: positional assignment when agent omits hunk_id
			questions[i].TargetHunkIdx = i + 1
		}
	}
}

func countHunks(files []git.File) int {
	total := 0
	for _, f := range files {
		total += len(f.Hunks)
	}
	return total
}
