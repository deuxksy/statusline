package config_test

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestLoadConfigTrayDefault(t *testing.T) {
	if cfg := config.LoadConfig("/non/existent/path/config.json"); cfg.Tray.Primary != "zai" {
		t.Errorf("absent config: expected tray.primary zai, got %q", cfg.Tray.Primary)
	}
	p := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(p, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if cfg := config.LoadConfig(p); cfg.Tray.Primary != "zai" {
		t.Errorf("empty config: expected tray.primary zai, got %q", cfg.Tray.Primary)
	}
}

func TestLoadConfigTrayValue(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(p, []byte(`{"tray":{"primary":"chatgpt"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if cfg := config.LoadConfig(p); cfg.Tray.Primary != "chatgpt" {
		t.Errorf("expected tray.primary chatgpt, got %q", cfg.Tray.Primary)
	}
}

func TestSaveConfigRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "config.json")
	cfg := config.DefaultConfig()
	cfg.Tray.Primary = "chatgpt"
	if err := config.SaveConfig(p, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := config.LoadConfig(p)
	if got.Tray.Primary != "chatgpt" {
		t.Errorf("round trip: expected chatgpt, got %q", got.Tray.Primary)
	}
}

func TestLoadConfigMigratesLegacyLayout(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		wantMain   []string
		wantSchema string
	}{
		{
			name:       "legacy 1.0 inserts cycleMode and sandbox after model",
			json:       `{"$schema":"1.0","layout":{"main":["permission","engineLabel","model","thinking"]}}`,
			wantMain:   []string{"permission", "engineLabel", "model", "cycleMode", "sandbox", "thinking"},
			wantSchema: "1.1",
		},
		{
			name:       "missing schema without model appends cycleMode and sandbox",
			json:       `{"layout":{"main":["engineLabel","thinking"]}}`,
			wantMain:   []string{"engineLabel", "thinking", "cycleMode", "sandbox"},
			wantSchema: "1.1",
		},
		{
			name:       "legacy already containing cycleMode inserts only sandbox",
			json:       `{"$schema":"1.0","layout":{"main":["permission","engineLabel","model","cycleMode","thinking"]}}`,
			wantMain:   []string{"permission", "engineLabel", "model", "cycleMode", "sandbox", "thinking"},
			wantSchema: "1.1",
		},
		{
			name:       "current schema 1.1 preserves layout untouched",
			json:       `{"$schema":"1.1","layout":{"main":["engineLabel","model"]}}`,
			wantMain:   []string{"engineLabel", "model"},
			wantSchema: "1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(p, []byte(tt.json), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg := config.LoadConfig(p)
			if cfg.Schema != tt.wantSchema {
				t.Errorf("schema = %q, want %q", cfg.Schema, tt.wantSchema)
			}
			if !reflect.DeepEqual(cfg.Layout.Main, tt.wantMain) {
				t.Errorf("layout.main = %v, want %v", cfg.Layout.Main, tt.wantMain)
			}
		})
	}
}
