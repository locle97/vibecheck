package git

import (
	"bufio"
	"os/exec"
	"strings"
)

// LineKind classifies a single diff line.
type LineKind int

const (
	LineContext LineKind = iota
	LineAdded
	LineRemoved
)

// Line is one line inside a hunk.
type Line struct {
	Kind    LineKind
	Content string // includes the leading +/- or space character
}

// Hunk represents a contiguous block of changes within a file.
type Hunk struct {
	Header string // e.g. "@@ -1,5 +1,6 @@ func foo()"
	Lines  []Line
}

// File is a changed file and all its hunks.
type File struct {
	Path      string
	IsNew     bool
	IsDeleted bool
	Hunks     []Hunk
}

// ParseStagedDiff runs `git diff --cached` and returns the parsed result.
func ParseStagedDiff() ([]File, error) {
	out, err := exec.Command("git", "diff", "--cached").Output()
	if err != nil {
		return nil, err
	}
	return ParseDiff(string(out))
}

// ParseDiff parses the raw unified diff text produced by `git diff --cached`.
func ParseDiff(raw string) ([]File, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var files []File
	var curFile *File
	var curHunk *Hunk

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "diff --git "):
			// flush previous hunk/file
			if curHunk != nil && curFile != nil {
				curFile.Hunks = append(curFile.Hunks, *curHunk)
				curHunk = nil
			}
			if curFile != nil {
				files = append(files, *curFile)
			}
			curFile = &File{}
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				curFile.Path = strings.TrimPrefix(parts[3], "b/")
			}

		case strings.HasPrefix(line, "new file mode"):
			if curFile != nil {
				curFile.IsNew = true
			}

		case strings.HasPrefix(line, "deleted file mode"):
			if curFile != nil {
				curFile.IsDeleted = true
			}

		case strings.HasPrefix(line, "+++ "):
			if curFile != nil {
				path := strings.TrimPrefix(line, "+++ ")
				if path == "/dev/null" {
					curFile.IsDeleted = true
					continue
				}
				path = strings.TrimPrefix(path, "b/")
				curFile.Path = path
			}

		case strings.HasPrefix(line, "--- "):
			if curFile != nil && curFile.Path == "" {
				path := strings.TrimPrefix(line, "--- ")
				if path != "/dev/null" {
					curFile.Path = strings.TrimPrefix(path, "a/")
				}
			}

		case strings.HasPrefix(line, "@@ "):
			if curFile == nil {
				continue
			}
			if curHunk != nil {
				curFile.Hunks = append(curFile.Hunks, *curHunk)
			}
			curHunk = &Hunk{Header: line}

		case curHunk != nil:
			if len(line) == 0 {
				curHunk.Lines = append(curHunk.Lines, Line{Kind: LineContext, Content: " "})
				continue
			}
			var kind LineKind
			switch line[0] {
			case '+':
				kind = LineAdded
			case '-':
				kind = LineRemoved
			default:
				kind = LineContext
			}
			curHunk.Lines = append(curHunk.Lines, Line{Kind: kind, Content: line})
		}
	}

	// flush last hunk/file
	if curHunk != nil && curFile != nil {
		curFile.Hunks = append(curFile.Hunks, *curHunk)
	}
	if curFile != nil {
		files = append(files, *curFile)
	}

	return files, scanner.Err()
}
