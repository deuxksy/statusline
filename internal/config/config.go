package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

type ElementsConfig struct {
	EngineLabel  bool   `json:"engineLabel"`
	Model        bool   `json:"model"`
	ModelFormat  string `json:"modelFormat"`
	GitRepo      bool   `json:"gitRepo"`
	GitBranch    bool   `json:"gitBranch"`
	GitStatus    bool   `json:"gitStatus"`
	Cwd          bool   `json:"cwd"`
	CwdFormat    string `json:"cwdFormat"`
	Hostname     bool   `json:"hostname"`
	ContextBar   bool   `json:"contextBar"`
	ShowTokens   bool   `json:"showTokens"`
	ActiveSkills bool   `json:"activeSkills"`
	LastTool     bool   `json:"lastTool"`
	Thinking     bool   `json:"thinking"`
	Quota        bool   `json:"quota"`
	Permission   bool   `json:"permission"`
	CycleMode    bool   `json:"cycleMode"`
	Sandbox      bool   `json:"sandbox"`
}

type ThresholdsConfig struct {
	ContextWarning  int `json:"contextWarning"`
	ContextCritical int `json:"contextCritical"`
}

type LayoutConfig struct {
	Line1      []string `json:"line1"`
	Main       []string `json:"main"`
	Line1Width int      `json:"line1Width,omitempty"`
}

type TrayConfig struct {
	Primary string `json:"primary"` // 메뉴바 타이틀에 노출할 메인 provider (zai | chatgpt)
}

type Config struct {
	Schema     string           `json:"$schema,omitempty"`
	Elements   ElementsConfig   `json:"elements"`
	Thresholds ThresholdsConfig `json:"thresholds"`
	WrapMode   string           `json:"wrapMode"`
	Theme      string           `json:"theme"`
	Layout     LayoutConfig     `json:"layout"`
	Tray       TrayConfig       `json:"tray"`
}

func DefaultConfig() *Config {
	return &Config{
		Schema: "1.1",
		Elements: ElementsConfig{
			EngineLabel:  true,
			Model:        true,
			ModelFormat:  "full",
			GitRepo:      true,
			GitBranch:    true,
			GitStatus:    true,
			Cwd:          true,
			CwdFormat:    "folder",
			Hostname:     false,
			ContextBar:   true,
			ShowTokens:   true,
			ActiveSkills: true,
			LastTool:     true,
			Thinking:     true,
			Quota:        true,
			Permission:   true,
			CycleMode:    true,
			Sandbox:      true,
		},
		Thresholds: ThresholdsConfig{
			ContextWarning:  70,
			ContextCritical: 85,
		},
		WrapMode: "truncate",
		Theme:    "sleek_dark",
		Layout: LayoutConfig{
			Line1:      []string{"hostname", "cwd", "gitRepo", "gitBranch", "gitStatus"},
			Main:       []string{"permission", "engineLabel", "model", "cycleMode", "sandbox", "thinking", "contextBar", "tokens", "quota", "activeSkills", "lastTool"},
			Line1Width: 28,
		},
		Tray: TrayConfig{Primary: "zai"},
	}
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}
	var raw struct {
		Schema string `json:"$schema"`
	}
	_ = json.Unmarshal(data, &raw)

	cfg := *DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	if raw.Schema == "" || raw.Schema == "1.0" {
		cfg.Layout.Main = migrateLayoutMain(cfg.Layout.Main, raw.Schema)
		cfg.Schema = "1.1"
	}
	return &cfg
}

// migrateLayoutMain upgrades pre-1.1 layout.main slices by ensuring cycleMode
// and sandbox are present after model (or appended if model is absent).
func migrateLayoutMain(main []string, schema string) []string {
	if schema != "" && schema != "1.0" {
		return main
	}
	hasCycle := slices.Contains(main, "cycleMode")
	hasSandbox := slices.Contains(main, "sandbox")
	if hasCycle && hasSandbox {
		return main
	}

	insertAt := len(main)
	if idx := slices.Index(main, "model"); idx >= 0 {
		insertAt = idx + 1
		if hasCycle && !hasSandbox && insertAt < len(main) && main[insertAt] == "cycleMode" {
			insertAt++
		}
	}

	toInsert := make([]string, 0, 2)
	if !hasCycle {
		toInsert = append(toInsert, "cycleMode")
	}
	if !hasSandbox {
		toInsert = append(toInsert, "sandbox")
	}

	out := make([]string, 0, len(main)+len(toInsert))
	out = append(out, main[:insertAt]...)
	out = append(out, toInsert...)
	out = append(out, main[insertAt:]...)
	return out
}

func SaveDefaultConfig(path string) error {
	return saveConfigAt(path, DefaultConfig())
}

// SaveConfig — 트레이 메인 provider 전환 등 런타임 변경 사항을 기존 파일 규약(0600)으로 저장.
func SaveConfig(path string, cfg *Config) error {
	return saveConfigAt(path, cfg)
}

func saveConfigAt(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
