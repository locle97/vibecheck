package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// claudeAgent invokes the Claude Code CLI.
// Diff is piped via stdin; prompt is passed via -p flag.
// Command: echo "<diff>" | claude -p "<prompt>" --output-format json
type claudeAgent struct{}

type printEnvelope struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	IsError bool   `json:"is_error"`
	Result  string `json:"result"`
	Error   string `json:"error"`
}

func (c *claudeAgent) Complete(ctx context.Context, prompt, diff string) (string, error) {
	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--output-format", "json")
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
		return "", fmt.Errorf("claude agent error: %w\nstdout: %s\nstderr: %s", err, out, errOut)
	}

	return unwrapPrintJSON(stdout.String(), "claude")
}

func unwrapPrintJSON(raw, provider string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%s agent returned empty stdout", provider)
	}

	var envelope printEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err == nil {
		if envelope.IsError {
			msg := strings.TrimSpace(envelope.Result)
			if msg == "" {
				msg = strings.TrimSpace(envelope.Error)
			}
			if msg == "" {
				msg = "unknown error"
			}
			return "", fmt.Errorf("%s agent returned error: %s", provider, msg)
		}
		if strings.TrimSpace(envelope.Result) != "" {
			return envelope.Result, nil
		}
	}

	return raw, nil
}
