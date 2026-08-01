package adapter

import (
	"bytes"
	"statusline/internal/model"
)

func ParseInput(cliFlag string, rawInput []byte, env map[string]string) (*model.UnifiedStatus, error) {
	targetEngine := cliFlag
	if targetEngine == "auto" || targetEngine == "" {
		if bytes.Contains(rawInput, []byte("omcLabel")) || env["CLAUDE_CONFIG_DIR"] != "" {
			targetEngine = "claude"
		} else if bytes.Contains(rawInput, []byte("antigravity")) || env["ANTIGRAVITY_APP_DIR"] != "" {
			targetEngine = "antigravity"
		} else if env["CODEX_ENV"] != "" {
			targetEngine = "codex"
		} else {
			targetEngine = "generic"
		}
	}

	var a EngineAdapter
	switch targetEngine {
	case "claude":
		a = &ClaudeAdapter{}
	case "codex":
		a = &CodexAdapter{}
	case "antigravity":
		a = &AntigravityAdapter{}
	default:
		st := model.NewUnifiedStatus("generic")
		return st, nil
	}
	return a.Parse(rawInput, env)
}
