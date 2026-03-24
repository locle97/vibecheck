package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

// Provider identifies the coding agent backend.
type Provider string

const (
	ProviderClaude    Provider = "claude"
	ProviderCursor    Provider = "cursor-agent"
	ProviderOpenCode  Provider = "opencode"
)

// Config is the top-level configuration structure.
type Config struct {
	Agent AgentConfig `toml:"agent"`
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
