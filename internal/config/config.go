package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		Schema: "1.0",
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
		},
		Thresholds: ThresholdsConfig{
			ContextWarning:  70,
			ContextCritical: 85,
		},
		WrapMode: "truncate",
		Theme:    "sleek_dark",
		Layout: LayoutConfig{
			Line1:      []string{"hostname", "cwd", "gitRepo", "gitBranch", "gitStatus"},
			Main:       []string{"permission", "engineLabel", "model", "thinking", "contextBar", "tokens", "quota", "activeSkills", "lastTool"},
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
	cfg := *DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return &cfg
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
