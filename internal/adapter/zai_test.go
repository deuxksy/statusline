package adapter_test

import (
	"testing"

	"statusline/internal/adapter"
)

func TestProviderFromBaseURL(t *testing.T) {
	cases := []struct{ name, rawURL, want string }{
		{"zai", "https://api.z.ai/api/anthropic", "zai"},
		{"zhipu-open", "https://open.bigmodel.cn/api/anthropic", "zhipu"},
		{"zhipu-dev", "https://dev.bigmodel.cn/api/anthropic", "zhipu"},
		{"anthropic-official", "https://api.anthropic.com", ""},
		{"empty", "", ""},
		{"no-host", "api.z.ai", ""}, // 스킴 없으면 Host 파싱 불가 → 미감지 (SSRF 보수 적용)
		{"garbage", "://bad url", ""},
		{"lookalike-zai", "https://api.z.ai.evil.com/api/anthropic", ""},
		{"lookalike-zhipu", "https://evil-bigmodel.cn.attacker.net", ""},
		{"port", "https://api.z.ai:8443/api/anthropic", "zai"},
	}
	for _, tc := range cases {
		if got := adapter.ProviderFromBaseURL(tc.rawURL); got != tc.want {
			t.Errorf("%s: ProviderFromBaseURL(%q) = %q, want %q", tc.name, tc.rawURL, got, tc.want)
		}
	}
}

func TestProviderFromEnv(t *testing.T) {
	env := map[string]string{"ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic"}
	if got := adapter.ProviderFromEnv(env, "whatever"); got != "zai" {
		t.Errorf("env 우선 감지 실패: got %q, want zai", got)
	}
	if got := adapter.ProviderFromEnv(map[string]string{}, "GLM-5.3"); got != "zai" {
		t.Errorf("glm- 폴백 실패: got %q, want zai", got)
	}
	if got := adapter.ProviderFromEnv(map[string]string{}, "claude-sonnet-5"); got != "" {
		t.Errorf("비-Z.AI 모델 감지 오류: got %q, want empty", got)
	}
}

func TestParseInputInjectsZaiProvider(t *testing.T) {
	payload := []byte(`{"omcLabel":"OMC","model":{"name":"glm-5.3"}}`)
	env := map[string]string{
		"ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
	}
	st, err := adapter.ParseInput("claude", payload, env)
	if err != nil {
		t.Fatal(err)
	}
	if st.Provider != "zai" {
		t.Errorf("Provider = %q, want zai", st.Provider)
	}
	if !st.Capabilities.HasQuota {
		t.Error("zai provider는 HasQuota = true여야 함")
	}
}

func TestParseInputNonZaiStaysEmpty(t *testing.T) {
	payload := []byte(`{"omcLabel":"OMC","model":{"name":"claude-sonnet-5"}}`)
	st, err := adapter.ParseInput("auto", payload, map[string]string{"CLAUDE_CONFIG_DIR": "/tmp/claude"})
	if err != nil {
		t.Fatal(err)
	}
	if st.Provider != "" {
		t.Errorf("Provider = %q, want empty", st.Provider)
	}
	if st.Capabilities.HasQuota {
		t.Error("비-Z.AI claude는 HasQuota = false여야 함")
	}
}
