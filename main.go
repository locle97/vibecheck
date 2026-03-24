package main

import (
	"os"

	"github.com/locle97/vibecheck/cmd"
)

func main() {
	root := cmd.NewRootCmd(os.Stdout)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
