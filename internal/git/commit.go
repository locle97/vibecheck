package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// ParseStagedDiff runs `git diff --cached` and returns the parsed result.
func Commit(commitMessage string) (error) {
	out, err := exec.Command("git", "commit", "-m", commitMessage).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
