package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/locle97/vibecheck/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := config.Load("/tmp/vibecheck_nonexistent_123456/config.toml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Agent.Provider != config.ProviderClaude {
		t.Errorf("default Provider = %q, want %q", cfg.Agent.Provider, config.ProviderClaude)
	}
	if cfg.Agent.Model != "claude-opus-4-6" {
		t.Errorf("default Model = %q, want %q", cfg.Agent.Model, "claude-opus-4-6")
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[agent]
provider = "cursor-agent"
model    = "cursor-small"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Agent.Provider != config.ProviderCursor {
		t.Errorf("Provider = %q, want %q", cfg.Agent.Provider, config.ProviderCursor)
	}
	if cfg.Agent.Model != "cursor-small" {
		t.Errorf("Model = %q, want %q", cfg.Agent.Model, "cursor-small")
	}
}

func TestLoadConfig_OpenCode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[agent]
provider = "opencode"
model    = "gpt-4o"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Agent.Provider != config.ProviderOpenCode {
		t.Errorf("Provider = %q, want %q", cfg.Agent.Provider, config.ProviderOpenCode)
	}
}
