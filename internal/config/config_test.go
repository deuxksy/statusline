package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"statusline/internal/config"
)

func TestLoadConfigEmbeddedDefault(t *testing.T) {
	cfg := config.LoadConfig("/non/existent/path/config.json")
	if cfg.Theme != "sleek_dark" {
		t.Errorf("expected sleek_dark theme, got %s", cfg.Theme)
	}
	if cfg.Thresholds.ContextWarning != 70 {
		t.Errorf("expected 70 contextWarning, got %d", cfg.Thresholds.ContextWarning)
	}
	if cfg.Layout.Line1Width != 28 {
		t.Errorf("expected 28 line1Width, got %d", cfg.Layout.Line1Width)
	}
}

func TestSaveDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subDir", "config.json")

	err := config.SaveDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected config file to exist at %s", configPath)
	}

	loadedCfg := config.LoadConfig(configPath)
	if loadedCfg.Theme != "sleek_dark" {
		t.Errorf("expected loaded config theme to be sleek_dark, got %s", loadedCfg.Theme)
	}
	if !loadedCfg.Elements.Permission {
		t.Errorf("expected loaded config permission to be true by default")
	}
	if loadedCfg.Layout.Line1Width != 28 {
		t.Errorf("expected loaded config line1Width to be 28, got %d", loadedCfg.Layout.Line1Width)
	}
}
