package git_test

import (
	"strings"
	"testing"

	"github.com/locle97/vibecheck/internal/git"
)

// rawDiff is a minimal but realistic `git diff --cached` snippet used across tests.
const rawDiff = `diff --git a/foo.go b/foo.go
index 1234567..abcdefg 100644
--- a/foo.go
+++ b/foo.go
@@ -1,5 +1,6 @@
 package main

-func hello() {}
+func hello() {
+	fmt.Println("hello")
+}

 func main() {}
diff --git a/bar.go b/bar.go
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/bar.go
@@ -0,0 +1,3 @@
+package main
+
+// bar is new
`

func TestParseDiff_FileCount(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("want 2 files, got %d", len(files))
	}
}

func TestParseDiff_FilePaths(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files[0].Path != "foo.go" {
		t.Errorf("want foo.go, got %q", files[0].Path)
	}
	if files[1].Path != "bar.go" {
		t.Errorf("want bar.go, got %q", files[1].Path)
	}
}

func TestParseDiff_HunkCount(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files[0].Hunks) != 1 {
		t.Errorf("foo.go: want 1 hunk, got %d", len(files[0].Hunks))
	}
	if len(files[1].Hunks) != 1 {
		t.Errorf("bar.go: want 1 hunk, got %d", len(files[1].Hunks))
	}
}

func TestParseDiff_HunkHeader(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := files[0].Hunks[0]
	if !strings.HasPrefix(h.Header, "@@ -1,5 +1,6 @@") {
		t.Errorf("unexpected header: %q", h.Header)
	}
}

func TestParseDiff_HunkLines(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := files[0].Hunks[0]
	if len(h.Lines) == 0 {
		t.Fatal("expected hunk lines, got none")
	}

	// find the removed line
	var foundRemoved, foundAdded bool
	for _, l := range h.Lines {
		if l.Kind == git.LineRemoved && strings.Contains(l.Content, "func hello() {}") {
			foundRemoved = true
		}
		if l.Kind == git.LineAdded && strings.Contains(l.Content, `fmt.Println`) {
			foundAdded = true
		}
	}
	if !foundRemoved {
		t.Error("did not find removed line")
	}
	if !foundAdded {
		t.Error("did not find added line")
	}
}

func TestParseDiff_NewFile(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !files[1].IsNew {
		t.Error("bar.go should be marked as new file")
	}
	if files[0].IsNew {
		t.Error("foo.go should NOT be marked as new file")
	}
}

func TestParseDiff_EmptyInput(t *testing.T) {
	files, err := git.ParseDiff("")
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("want 0 files, got %d", len(files))
	}
}

func TestParseDiff_MultipleHunks(t *testing.T) {
	multiHunk := `diff --git a/x.go b/x.go
index aaaaaaa..bbbbbbb 100644
--- a/x.go
+++ b/x.go
@@ -1,3 +1,3 @@
 line1
-line2
+LINE2
 line3
@@ -10,3 +10,3 @@
 line10
-line11
+LINE11
 line12
`
	files, err := git.ParseDiff(multiHunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files[0].Hunks) != 2 {
		t.Errorf("want 2 hunks, got %d", len(files[0].Hunks))
	}
}

func TestParseDiff_LineKinds(t *testing.T) {
	files, err := git.ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, l := range files[0].Hunks[0].Lines {
		switch l.Content[0] {
		case '+':
			if l.Kind != git.LineAdded {
				t.Errorf("line starting with + should be LineAdded, got %v", l.Kind)
			}
		case '-':
			if l.Kind != git.LineRemoved {
				t.Errorf("line starting with - should be LineRemoved, got %v", l.Kind)
			}
		case ' ':
			if l.Kind != git.LineContext {
				t.Errorf("line starting with space should be LineContext, got %v", l.Kind)
			}
		}
	}
}

func TestParseDiff_DeletedFile(t *testing.T) {
	deletedDiff := `diff --git a/old.txt b/old.txt
deleted file mode 100644
index 1111111..0000000
--- a/old.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-hello
-world
`

	files, err := git.ParseDiff(deletedDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1 file, got %d", len(files))
	}
	if files[0].Path != "old.txt" {
		t.Fatalf("want path old.txt, got %q", files[0].Path)
	}
	if !files[0].IsDeleted {
		t.Fatal("old.txt should be marked as deleted")
	}
	if files[0].IsNew {
		t.Fatal("deleted file should not be marked as new")
	}
}
