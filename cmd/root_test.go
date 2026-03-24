package cmd_test

import (
	"bytes"
	"testing"

	"github.com/locle97/vibecheck/cmd"
)

func TestRootCommand_Executes(t *testing.T) {
	var buf bytes.Buffer
	root := cmd.NewRootCmd(&buf)
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := buf.String()
	if got == "" {
		t.Error("expected output, got empty string")
	}
}
