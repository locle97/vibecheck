package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// ParseUntrackedFiles returns new files that git is not yet tracking.
// Each file is returned as a File with IsNew=true and its full content as added lines.
func ParseUntrackedFiles() ([]File, error) {
	out, err := exec.Command("git", "ls-files", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil, err
	}

	var files []File
	for path := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if path == "" {
			continue
		}
		// git diff --no-index exits 1 when files differ (always true for non-empty files).
		diffOut, _ := exec.Command("git", "diff", "--no-index", "--", "/dev/null", path).Output()
		if len(diffOut) == 0 {
			// Empty file — add a minimal entry so the user can stage it.
			files = append(files, File{Path: path, IsNew: true})
			continue
		}
		parsed, parseErr := ParseDiff(string(diffOut))
		if parseErr != nil || len(parsed) == 0 {
			files = append(files, File{Path: path, IsNew: true})
			continue
		}
		f := parsed[0]
		f.Path = path
		f.IsNew = true
		files = append(files, f)
	}
	return files, nil
}

// ParseUnstagedDiff runs `git diff` (working tree vs index) and returns parsed files.
func ParseUnstagedDiff() ([]File, error) {
	out, err := exec.Command("git", "diff").Output()
	if err != nil {
		return nil, err
	}
	return ParseDiff(string(out))
}

// StageFile runs `git add -- <path>` to stage an entire file.
func StageFile(path string) error {
	return exec.Command("git", "add", "--", path).Run()
}

// UnstageFile runs `git restore --staged -- <path>` to unstage an entire file.
func UnstageFile(path string) error {
	return exec.Command("git", "restore", "--staged", "--", path).Run()
}

// StageHunk constructs a minimal patch for hunk and pipes it to `git apply --cached`.
func StageHunk(path string, hunk Hunk) error {
	return applyHunkPatch(path, hunk, false)
}

// UnstageHunk constructs a minimal patch for hunk and pipes it to `git apply --cached --reverse`.
func UnstageHunk(path string, hunk Hunk) error {
	return applyHunkPatch(path, hunk, true)
}

// applyHunkPatch builds a minimal unified diff patch for a single hunk and
// pipes it to git apply --cached. reverse=true reverses the patch (unstage).
func applyHunkPatch(path string, hunk Hunk, reverse bool) error {
	// Fetch the raw diff for this specific file to extract the file header.
	var diffArgs []string
	if reverse {
		diffArgs = []string{"diff", "--cached", "--", path}
	} else {
		diffArgs = []string{"diff", "--", path}
	}
	out, err := exec.Command("git", diffArgs...).Output()
	if err != nil {
		return fmt.Errorf("git diff for patch: %w", err)
	}

	// Extract header lines (everything before the first @@ line).
	var headerLines []string
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.HasPrefix(line, "@@") {
			break
		}
		headerLines = append(headerLines, line)
	}

	// Reconstruct the hunk body from the parsed Hunk struct.
	var hunkBody strings.Builder
	hunkBody.WriteString(hunk.Header)
	hunkBody.WriteString("\n")
	for _, l := range hunk.Lines {
		hunkBody.WriteString(l.Content)
		hunkBody.WriteString("\n")
	}

	patch := strings.Join(headerLines, "\n") + "\n" + hunkBody.String()

	var applyArgs []string
	if reverse {
		applyArgs = []string{"apply", "--cached", "--reverse"}
	} else {
		applyArgs = []string{"apply", "--cached"}
	}
	cmd := exec.Command("git", applyArgs...)
	cmd.Stdin = strings.NewReader(patch)
	if applyOut, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git apply: %w\n%s", err, strings.TrimSpace(string(applyOut)))
	}
	return nil
}
