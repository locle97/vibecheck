package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewRootCmd returns the root cobra command. out is used for command output,
// allowing tests to capture it.
func NewRootCmd(out io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:   "vibecheck",
		Short: "Quiz yourself on staged changes before committing",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(out, "vibecheck starting...")
			return nil
		},
	}
	root.SetOut(out)

	return root
}
