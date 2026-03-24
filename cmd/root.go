package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/locle97/vibecheck/config"
	"github.com/locle97/vibecheck/internal/agent"
	"github.com/locle97/vibecheck/internal/git"
	"github.com/locle97/vibecheck/internal/quiz"
	"github.com/spf13/cobra"
)

type rootDeps struct {
	loadConfig      func(path string) (config.Config, error)
	parseStagedDiff func() ([]git.File, error)
	newAgent        func(binary, model string) (agent.Agent, error)
	newGenerator    func(a agent.Agent) *quiz.Generator
}

func defaultRootDeps() rootDeps {
	return rootDeps{
		loadConfig:      config.Load,
		parseStagedDiff: git.ParseStagedDiff,
		newAgent:        agent.New,
		newGenerator:    quiz.New,
	}
}

// NewRootCmd returns the root cobra command. out is used for command output,
// allowing tests to capture it.
func NewRootCmd(out io.Writer) *cobra.Command {
	return newRootCmd(out, defaultRootDeps())
}

func newRootCmd(out io.Writer, deps rootDeps) *cobra.Command {
	configPath := defaultConfigPath()

	root := &cobra.Command{
		Use:   "vibecheck",
		Short: "Quiz yourself on staged changes before committing",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := deps.loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			files, err := deps.parseStagedDiff()
			if err != nil {
				return fmt.Errorf("read staged diff: %w", err)
			}
			if len(files) == 0 {
				fmt.Fprintln(out, "No staged changes found. Stage files with git add and run vibecheck again.")
				return nil
			}

			a, err := deps.newAgent(string(cfg.Agent.Provider), cfg.Agent.Model)
			if err != nil {
				return fmt.Errorf("create agent: %w", err)
			}

			generator := deps.newGenerator(a)
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			questions, err := generator.GenerateQuestions(ctx, files)
			if err != nil {
				return fmt.Errorf("generate questions: %w", err)
			}

			fmt.Fprintf(out, "Provider: %s | Model: %s\n", cfg.Agent.Provider, cfg.Agent.Model)
			fmt.Fprintf(out, "Parsed %d changed file(s), generated %d question(s).\n\n", len(files), len(questions))

			if len(questions) == 0 {
				fmt.Fprintln(out, "No questions returned by agent.")
				return nil
			}

			for _, q := range questions {
				fmt.Fprintf(out, "%d) %s\n", q.ID, q.Question)
				for i, option := range q.Options {
					fmt.Fprintf(out, "  %c. %s\n", optionLabel(i), option)
				}
				if q.Answer >= 0 && q.Answer < len(q.Options) {
					fmt.Fprintf(out, "  Answer: %c. %s\n", optionLabel(q.Answer), q.Options[q.Answer])
				} else {
					fmt.Fprintln(out, "  Answer: (invalid answer index from agent)")
				}
				if q.Hint != "" {
					fmt.Fprintf(out, "  Hint: %s\n", q.Hint)
				}
				fmt.Fprintln(out)
			}

			return nil
		},
	}
	root.SetOut(out)
	root.Flags().StringVar(&configPath, "config", configPath, "path to config TOML file")

	return root
}

func defaultConfigPath() string {
	if _, err := os.Stat("config.toml"); err == nil {
		return "config.toml"
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "config.toml"
	}
	return filepath.Join(configDir, "vibecheck", "config.toml")
}

func optionLabel(idx int) byte {
	if idx < 0 || idx > 25 {
		return '?'
	}
	return byte('A' + idx)
}
