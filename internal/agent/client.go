package agent

import (
	"context"
	"fmt"
)

// Agent shells out to a coding agent CLI in print/non-interactive mode,
// passes the diff as context, and returns the raw JSON string response.
type Agent interface {
	// Complete sends prompt to the coding agent with diff as context.
	// Returns the raw JSON string from the agent's stdout.
	Complete(ctx context.Context, prompt, diff string) (string, error)
}

// New constructs the appropriate Agent from the binary name in config.
func New(binary string) (Agent, error) {
	switch binary {
	case "claude":
		return &claudeAgent{}, nil
	case "opencode":
		return &opencodeAgent{}, nil
	case "cursor-agent":
		return &cursorAgent{}, nil
	default:
		return nil, fmt.Errorf("unsupported agent binary %q: must be one of: claude, opencode, cursor-agent", binary)
	}
}
