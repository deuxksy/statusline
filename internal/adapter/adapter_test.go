package adapter_test

import (
	"os"
	"testing"
	"time"

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

func TestClaudeAdapterContextWindow(t *testing.T) {
	a := &adapter.ClaudeAdapter{}
	jsonPayload := []byte(`{
		"model": {"id": "glm-5.3", "display_name": "GLM-5.3"},
		"transcript_path": "/home/user/.claude/projects/-x/edd837a0.jsonl",
		"context_window": {
			"total_input_tokens": 57340,
			"total_output_tokens": 1574,
			"context_window_size": 1000000,
			"used_percentage": 5.7,
			"current_usage": {
				"input_tokens": 124,
				"output_tokens": 1574,
				"cache_creation_input_tokens": 0,
				"cache_read_input_tokens": 57216
			}
		}
	}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.ContextTokens != 57340 {
		t.Errorf("expected context tokens 57340, got %d", status.ContextTokens)
	}
	if status.ContextLimit != 1000000 {
		t.Errorf("expected context limit 1000000, got %d", status.ContextLimit)
	}
}

func TestClaudeAdapterContextWindowPrecedenceOverLegacyFields(t *testing.T) {
	a := &adapter.ClaudeAdapter{}
	jsonPayload := []byte(`{
		"contextBar": {"percentage": 45},
		"context_window": {
			"total_input_tokens": 20000,
			"context_window_size": 200000
		}
	}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.ContextTokens != 20000 || status.ContextLimit != 200000 {
		t.Errorf("expected context_window to take precedence (20000/200000), got %d/%d", status.ContextTokens, status.ContextLimit)
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

func TestCodexAdapterSessionSnapshot(t *testing.T) {
	input, err := os.ReadFile("../../docs/samples/codex.json")
	if err != nil {
		t.Fatal(err)
	}
	status, err := (&adapter.CodexAdapter{}).Parse(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status.Model != "gpt-6-astra" || status.Cwd != "/workspace/statusline" || status.ReasoningEffort != "low" {
		t.Fatalf("Codex context not parsed: %+v", status)
	}
	if status.TotalTokenUsage.TotalTokens != 354563 || status.LastTokenUsage.InputTokens != 42064 || status.ContextLimit != 237500 {
		t.Fatalf("Codex token usage not parsed: %+v", status)
	}
	if !status.Capabilities.HasTokens || status.ContextTokens != 18 {
		t.Fatalf("Codex context display not enabled: %+v", status)
	}
	if !status.Capabilities.HasQuota || len(status.Quota) != 1 || status.Quota[0].FiveH != 0.83 || status.Quota[0].Weekly != 0.39 {
		t.Fatalf("Codex quota not parsed: %+v", status.Quota)
	}
	if status.Quota[0].FiveHResetsAt != 1790366542 || status.Quota[0].WeeklyResetsAt != 1790573243 {
		t.Fatalf("Codex quota reset times not parsed: %+v", status.Quota)
	}
}

func TestCodexAdapterPartialSnapshot(t *testing.T) {
	status, err := (&adapter.CodexAdapter{}).Parse([]byte(`{"model":"gpt-6-sol","rate_limits":{"primary":{"used_percent":40}}}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if status.Model != "gpt-6-sol" || status.Capabilities.HasQuota || status.Capabilities.HasTokens {
		t.Fatalf("missing values should not render as zero usage: %+v", status)
	}
}

func TestAntigravityAdapter(t *testing.T) {
	a := &adapter.AntigravityAdapter{}
	jsonPayload := []byte(`{
		"model": "gemini-2.5-pro",
		"antigravity": true,
		"contextTokens": 55,
		"thinkingState": "thinking",
		"activeSkills": ["superpowers:systematic-debugging"],
		"lastTool": "replace_file_content"
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
	if status.ContextTokens != 55 {
		t.Errorf("expected context tokens 55, got %d", status.ContextTokens)
	}
	if status.ThinkingState != "thinking" {
		t.Errorf("expected thinking state thinking, got %s", status.ThinkingState)
	}
	if len(status.ActiveSkills) != 1 || status.ActiveSkills[0] != "superpowers:systematic-debugging" {
		t.Errorf("expected active skill 'superpowers:systematic-debugging', got %v", status.ActiveSkills)
	}
	if status.LastTool != "replace_file_content" {
		t.Errorf("expected last tool replace_file_content, got %s", status.LastTool)
	}
}

func TestAntigravityAdapterRealPayload(t *testing.T) {
	a := &adapter.AntigravityAdapter{}
	jsonPayload := []byte(`{
		"product": "antigravity",
		"model": {
			"id": "Gemini 3.6 Flash (Medium)",
			"display_name": "Gemini 3.6 Flash (Medium)",
			"effort": "medium"
		},
		"context_window": {
			"total_input_tokens": 50527,
			"used_percentage": 4.81863
		},
		"agent_state": "tool_use",
		"workspace": {
			"current_dir": "/home/crong/git/statusline"
		},
		"tool_confirmation_pending": true,
		"terminal_width": 209
	}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.EngineName != "antigravity" {
		t.Errorf("expected engine antigravity, got %s", status.EngineName)
	}
	if status.Model != "Gemini 3.6 Flash (Medium)" {
		t.Errorf("expected model Gemini 3.6 Flash (Medium), got %s", status.Model)
	}
	if status.ContextTokens != 5 {
		t.Errorf("expected context tokens 5, got %d", status.ContextTokens)
	}
	if status.ThinkingState != "tool_use" {
		t.Errorf("expected thinking state tool_use, got %s", status.ThinkingState)
	}
	if status.Cwd != "/home/crong/git/statusline" {
		t.Errorf("expected cwd /home/crong/git/statusline, got %s", status.Cwd)
	}
	if status.Permission != "pending" {
		t.Errorf("expected permission pending, got %s", status.Permission)
	}
	if status.TerminalWidth != 209 {
		t.Errorf("expected terminal width 209, got %d", status.TerminalWidth)
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
			name:           "Auto detect Codex by payload discriminator",
			cliFlag:        "auto",
			input:          `{"product":"codex","model":"gpt-6-sol"}`,
			env:            map[string]string{},
			expectedEngine: "codex",
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

func TestAntigravityAdapterQuota(t *testing.T) {
	a := &adapter.AntigravityAdapter{}
	jsonPayload := []byte(`{
		"model": "gemini-2.5-pro",
		"quota": {
			"3p-5h": {"remaining_fraction": 1.0, "reset_time": "2026-08-01T12:13:36Z"},
			"3p-weekly": {"remaining_fraction": 0.222908, "reset_time": "2026-08-03T07:00:41Z"},
			"gemini-5h": {"remaining_fraction": 0.933109, "reset_time": "2026-08-01T11:18:07Z"},
			"gemini-weekly": {"remaining_fraction": 0.5284749, "reset_time": "2026-08-05T04:59:12Z"}
		}
	}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Quota) != 2 {
		t.Fatalf("expected 2 quota categories, got %d", len(status.Quota))
	}

	// gemini가 먼저 와야 함
	if status.Quota[0].Name != "gemini" {
		t.Errorf("expected first category 'gemini', got %s", status.Quota[0].Name)
	}
	if status.Quota[1].Name != "3rd" {
		t.Errorf("expected second category '3rd', got %s", status.Quota[1].Name)
	}

	// gemini 값 검증
	if status.Quota[0].FiveH != 0.933109 {
		t.Errorf("expected gemini 5h 0.933109, got %f", status.Quota[0].FiveH)
	}
	if status.Quota[0].Weekly != 0.5284749 {
		t.Errorf("expected gemini weekly 0.5284749, got %f", status.Quota[0].Weekly)
	}
	expGemini5h, _ := time.Parse(time.RFC3339, "2026-08-01T11:18:07Z")
	if status.Quota[0].FiveHResetsAt != expGemini5h.UnixMilli() {
		t.Errorf("expected gemini 5h reset %d, got %d", expGemini5h.UnixMilli(), status.Quota[0].FiveHResetsAt)
	}
	expGeminiWk, _ := time.Parse(time.RFC3339, "2026-08-05T04:59:12Z")
	if status.Quota[0].WeeklyResetsAt != expGeminiWk.UnixMilli() {
		t.Errorf("expected gemini weekly reset %d, got %d", expGeminiWk.UnixMilli(), status.Quota[0].WeeklyResetsAt)
	}

	// 3rd 값 검증
	if status.Quota[1].FiveH != 1.0 {
		t.Errorf("expected 3rd 5h 1.0, got %f", status.Quota[1].FiveH)
	}
	if status.Quota[1].Weekly != 0.222908 {
		t.Errorf("expected 3rd weekly 0.222908, got %f", status.Quota[1].Weekly)
	}
	exp3p5h, _ := time.Parse(time.RFC3339, "2026-08-01T12:13:36Z")
	if status.Quota[1].FiveHResetsAt != exp3p5h.UnixMilli() {
		t.Errorf("expected 3rd 5h reset %d, got %d", exp3p5h.UnixMilli(), status.Quota[1].FiveHResetsAt)
	}
	exp3pWk, _ := time.Parse(time.RFC3339, "2026-08-03T07:00:41Z")
	if status.Quota[1].WeeklyResetsAt != exp3pWk.UnixMilli() {
		t.Errorf("expected 3rd weekly reset %d, got %d", exp3pWk.UnixMilli(), status.Quota[1].WeeklyResetsAt)
	}
}

func TestAntigravityAdapterNoQuota(t *testing.T) {
	a := &adapter.AntigravityAdapter{}
	jsonPayload := []byte(`{"model": "gemini-2.5-pro"}`)

	status, err := a.Parse(jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(status.Quota) != 0 {
		t.Errorf("expected 0 quota categories, got %d", len(status.Quota))
	}
}
