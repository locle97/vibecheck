package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// cursorAgent invokes the Cursor agent CLI.
// Diff is piped via stdin; prompt is passed via -p flag.
// Command: echo "<diff>" | cursor-agent -p "<prompt>" --output-format json
type cursorAgent struct{}

func (c *cursorAgent) Complete(ctx context.Context, prompt, diff string) (string, error) {
	cmd := exec.CommandContext(ctx, "cursor-agent", "-p", prompt, "--output-format", "json")
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
		return "", fmt.Errorf("cursor-agent error: %w\nstdout: %s\nstderr: %s", err, out, errOut)
	}

	return unwrapPrintJSON(stdout.String(), "cursor-agent")
}
