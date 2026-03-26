package quiz

import (
	"context"
	"strings"

	"github.com/locle97/vibecheck/internal/git"
)

// GenerateCommitMessage calls the agent to produce a conventional-commits
// formatted commit message for the staged diff.
func (g *Generator) GenerateCommitMessage(ctx context.Context, files []git.File) (string, error) {
	diff, _ := renderDiff(files)
	raw, err := g.agent.Complete(ctx, buildCommitPrompt(), diff)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

func buildCommitPrompt() string {
	return `You are an expert developer writing git commit messages.
Analyze the following staged diff and write a concise, informative commit message following conventional commits format.

Rules:
- First line: type(scope): short description (max 72 chars)
- type: feat, fix, refactor, style, docs, test, chore, perf
- Optional blank line + body with more details if needed
- Be specific about WHAT changed and WHY (not just "update files")
- No period at end of subject line

Output ONLY the commit message text, no explanation or markdown code fences.`
}
