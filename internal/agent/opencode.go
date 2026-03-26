package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// opencodeAgent invokes the OpenCode CLI.
// Diff is piped via stdin; prompt is passed as the final argument.
// Command: echo "<diff>" | opencode run --format json "<prompt>"
type opencodeAgent struct {
	model string
}

func (o *opencodeAgent) Complete(ctx context.Context, prompt, diff string) (string, error) {
	args := []string{"run", "--format", "json"}
	if strings.TrimSpace(o.model) != "" {
		args = append(args, "--model", o.model)
	}
	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Stdin = strings.NewReader(diff)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		out := strings.TrimSpace(stdout.String())
		errOut := strings.TrimSpace(stderr.String())
		if out == "" {
			out = "<empty>"
		}
		if errOut == "" {
			errOut = "<empty>"
		}
		return "", fmt.Errorf("opencode agent error: %w\nstdout: %s\nstderr: %s", err, out, errOut)
	}

	raw := strings.TrimSpace(stdout.String())
	if raw == "" {
		return "", fmt.Errorf("opencode agent returned empty stdout")
	}

	if content := extractNDJSONContent(raw); content != "" {
		return content, nil
	}

	return raw, nil
}

func extractNDJSONContent(raw string) string {
	fallback := ""

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}

		var event struct {
			Content string `json:"content"`
			Text    string `json:"text"`
			Part    struct {
				Content string `json:"content"`
				Text    string `json:"text"`
			} `json:"part"`
		}
		if json.Unmarshal([]byte(line), &event) != nil {
			continue
		}

		candidates := []string{
			strings.TrimSpace(event.Content),
			strings.TrimSpace(event.Text),
			strings.TrimSpace(event.Part.Content),
			strings.TrimSpace(event.Part.Text),
		}

		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}
			if strings.Contains(candidate, "[") {
				return candidate
			}
			if fallback == "" {
				fallback = candidate
			}
		}
	}

	return fallback
}
