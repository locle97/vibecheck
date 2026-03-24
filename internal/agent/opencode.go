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
type opencodeAgent struct{}

func (o *opencodeAgent) Complete(ctx context.Context, prompt, diff string) (string, error) {
	cmd := exec.CommandContext(ctx, "opencode", "run", "--format", "json", prompt)
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
		}
		if json.Unmarshal([]byte(line), &event) != nil {
			continue
		}

		if content := strings.TrimSpace(event.Content); content != "" {
			if strings.Contains(content, "[") {
				return content
			}
			if fallback == "" {
				fallback = content
			}
		}
		if text := strings.TrimSpace(event.Text); text != "" {
			if strings.Contains(text, "[") {
				return text
			}
			if fallback == "" {
				fallback = text
			}
		}
	}

	return fallback
}
