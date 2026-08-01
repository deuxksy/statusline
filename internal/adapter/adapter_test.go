package adapter_test

import (
	"testing"

	"statusline/internal/adapter"
)

func TestClaudeAdapter(t *testing.T) {
	a := &adapter.ClaudeAdapter{}
	jsonPayload := []byte(`{
		"omcLabel": "OMC",
		"model": {"displayName": "Claude 3.7 Sonnet"},
		"contextBar": {"percentage": 45},
		"thinking": {"state": "thinking"}
	}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.EngineName != "claude" {
		t.Errorf("expected engine claude, got %s", status.EngineName)
	}
	if status.Model != "Claude 3.7 Sonnet" {
		t.Errorf("expected model Claude 3.7 Sonnet, got %s", status.Model)
	}
	if status.ContextTokens != 45 {
		t.Errorf("expected context tokens 45, got %d", status.ContextTokens)
	}
	if status.ThinkingState != "thinking" {
		t.Errorf("expected thinking state 'thinking', got %s", status.ThinkingState)
	}
}

func TestCodexAdapter(t *testing.T) {
	a := &adapter.CodexAdapter{}
	status, err := a.Parse([]byte{}, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.EngineName != "codex" {
		t.Errorf("expected engine codex, got %s", status.EngineName)
	}
	if status.Model != "codex" {
		t.Errorf("expected model codex, got %s", status.Model)
	}
}

func TestAntigravityAdapter(t *testing.T) {
	a := &adapter.AntigravityAdapter{}
	jsonPayload := []byte(`{
		"model": "gemini-2.5-pro",
		"antigravity": true
	}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.EngineName != "antigravity" {
		t.Errorf("expected engine antigravity, got %s", status.EngineName)
	}
	if status.Model != "gemini-2.5-pro" {
		t.Errorf("expected model gemini-2.5-pro, got %s", status.Model)
	}
}

func TestParseInputAutoDetection(t *testing.T) {
	tests := []struct {
		name           string
		cliFlag        string
		input          string
		env            map[string]string
		expectedEngine string
	}{
		{
			name:           "Auto detect Claude by payload discriminator",
			cliFlag:        "auto",
			input:          `{"omcLabel":"OMC","model":{"displayName":"Claude 3.7 Sonnet"}}`,
			env:            map[string]string{},
			expectedEngine: "claude",
		},
		{
			name:           "Auto detect Claude by env var",
			cliFlag:        "auto",
			input:          `{}`,
			env:            map[string]string{"CLAUDE_CONFIG_DIR": "/home/user/.claude"},
			expectedEngine: "claude",
		},
		{
			name:           "Auto detect Antigravity by payload discriminator",
			cliFlag:        "auto",
			input:          `{"antigravity": true, "model": "gemini-2.5-pro"}`,
			env:            map[string]string{},
			expectedEngine: "antigravity",
		},
		{
			name:           "Auto detect Antigravity by env var",
			cliFlag:        "auto",
			input:          `{}`,
			env:            map[string]string{"ANTIGRAVITY_APP_DIR": "/home/user/.antigravity"},
			expectedEngine: "antigravity",
		},
		{
			name:           "Auto detect Codex by env var",
			cliFlag:        "auto",
			input:          `{}`,
			env:            map[string]string{"CODEX_ENV": "1"},
			expectedEngine: "codex",
		},
		{
			name:           "Generic fallback",
			cliFlag:        "auto",
			input:          `{}`,
			env:            map[string]string{},
			expectedEngine: "generic",
		},
		{
			name:           "Explicit flag override claude",
			cliFlag:        "claude",
			input:          `{}`,
			env:            map[string]string{},
			expectedEngine: "claude",
		},
		{
			name:           "Explicit flag override codex",
			cliFlag:        "codex",
			input:          `{}`,
			env:            map[string]string{},
			expectedEngine: "codex",
		},
		{
			name:           "Explicit flag override antigravity",
			cliFlag:        "antigravity",
			input:          `{}`,
			env:            map[string]string{},
			expectedEngine: "antigravity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := adapter.ParseInput(tt.cliFlag, []byte(tt.input), tt.env)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status.EngineName != tt.expectedEngine {
				t.Errorf("expected engine %s, got %s", tt.expectedEngine, status.EngineName)
			}
		})
	}
}
