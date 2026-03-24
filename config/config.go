package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

// Provider identifies the coding agent backend.
type Provider string

const (
	ProviderClaude   Provider = "claude"
	ProviderCursor   Provider = "cursor-agent"
	ProviderOpenCode Provider = "opencode"
)

// ReviewPhase constants used in RequiredPhases.
const (
	ReviewPhaseQuiz = "quiz"
)

// ReviewConfig controls quiz behaviour.
type ReviewConfig struct {
	PassThreshold float64 `toml:"pass_threshold"`
}

// RequiresPhase reports whether phase is in the required phases list.
// Currently only "quiz" is supported.
func (r ReviewConfig) RequiresPhase(phase string) bool {
	return phase == ReviewPhaseQuiz
}

// Config is the top-level configuration structure.
type Config struct {
	Agent  AgentConfig  `toml:"agent"`
	Review ReviewConfig `toml:"review"`
}

// AgentConfig holds coding agent settings.
type AgentConfig struct {
	Provider Provider `toml:"provider"`
	Model    string   `toml:"model"`
}

func defaults() Config {
	return Config{
		Agent: AgentConfig{
			Provider: ProviderClaude,
			Model:    "claude-opus-4-6",
		},
		Review: ReviewConfig{
			PassThreshold: 0.70,
		},
	}
}

// Load reads a TOML config file at path. If the file does not exist, defaults
// are returned without error.
func Load(path string) (Config, error) {
	cfg := defaults()
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		if !os.IsNotExist(err) {
			return cfg, err
		}
	}
	return cfg, nil
}
