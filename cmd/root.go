package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/locle97/vibecheck/config"
	"github.com/locle97/vibecheck/internal/agent"
	"github.com/locle97/vibecheck/internal/quiz"
	"github.com/locle97/vibecheck/tui"
)

type rootDeps struct {
	loadConfig   func(path string) (config.Config, error)
	newAgent     func(binary, model string) (agent.Agent, error)
	newGenerator func(a agent.Agent) *quiz.Generator
	runTUI       func(gen *quiz.Generator, cfg config.Config) error
}

func defaultRootDeps() rootDeps {
	return rootDeps{
		loadConfig:   config.Load,
		newAgent:     agent.New,
		newGenerator: quiz.New,
		runTUI:       defaultRunTUI,
	}
}

func defaultRunTUI(gen *quiz.Generator, cfg config.Config) error {
	app := tui.NewApp(gen, cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	finalApp, ok := m.(tui.App)
	if !ok {
		return nil
	}
	if !finalApp.Passed() {
		return fmt.Errorf("vibecheck: review failed")
	}
	if finalApp.CommitConfirmed() {
		out, err := exec.Command("git", "commit", "-m", finalApp.CommitMessage()).CombinedOutput()
		if err != nil {
			return fmt.Errorf("git commit: %w\n%s", err, strings.TrimSpace(string(out)))
		}
		fmt.Print(strings.TrimSpace(string(out)))
	}
	return nil
}

// NewRootCmd returns the root cobra command. out is used for non-TUI output
// (errors), allowing tests to capture it.
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

			a, err := deps.newAgent(string(cfg.Agent.Provider), cfg.Agent.Model)
			if err != nil {
				return fmt.Errorf("create agent: %w", err)
			}

			gen := deps.newGenerator(a)
			return deps.runTUI(gen, cfg)
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

	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}
	return filepath.Join(home, ".config", "vibecheck", "config.toml")
}
