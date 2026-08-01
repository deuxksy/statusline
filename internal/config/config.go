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
}

type ThresholdsConfig struct {
	ContextWarning  int `json:"contextWarning"`
	ContextCritical int `json:"contextCritical"`
}

type LayoutConfig struct {
	Line1 []string `json:"line1"`
	Main  []string `json:"main"`
}

type Config struct {
	Schema     string           `json:"$schema,omitempty"`
	Elements   ElementsConfig   `json:"elements"`
	Thresholds ThresholdsConfig `json:"thresholds"`
	WrapMode   string           `json:"wrapMode"`
	Theme      string           `json:"theme"`
	Layout     LayoutConfig     `json:"layout"`
}

func DefaultConfig() *Config {
	return &Config{
		Schema: "1.0",
		Elements: ElementsConfig{
			EngineLabel:  true,
			Model:        true,
			ModelFormat:  "short",
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
		},
		Thresholds: ThresholdsConfig{
			ContextWarning:  70,
			ContextCritical: 85,
		},
		WrapMode: "truncate",
		Theme:    "sleek_dark",
		Layout: LayoutConfig{
			Line1: []string{"hostname", "cwd", "gitRepo", "gitBranch", "gitStatus"},
			Main:  []string{"engineLabel", "model", "thinking", "contextBar", "tokens", "activeSkills", "lastTool"},
		},
	}
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return &cfg
}

func SaveDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(DefaultConfig(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
