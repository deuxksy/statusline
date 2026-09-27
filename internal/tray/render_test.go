package tray_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"statusline/internal/collect"
	"statusline/internal/tray"
)

func TestTitleZai(t *testing.T) {
	live := collect.LiveSession{Zai: &collect.ZaiSnapshot{TokenRemaining: 0.87, WeeklyRemaining: 0.45}}
	if got := tray.Title("zai", live); got != "z87 w45" {
		t.Errorf("got %q, want z87 w45", got)
	}
}

func TestTitleZaiNoWeekly(t *testing.T) {
	live := collect.LiveSession{Zai: &collect.ZaiSnapshot{TokenRemaining: 0.87}}
	if got := tray.Title("zai", live); got != "z87" {
		t.Errorf("got %q, want z87", got)
	}
}

func TestTitleInvalidPrimaryFallback(t *testing.T) {
	live := collect.LiveSession{Zai: &collect.ZaiSnapshot{TokenRemaining: 0.87}}
	if got := tray.Title("gpt", live); got != "z87" {
		t.Errorf("got %q, want z87 (zai fallback)", got)
	}
}

func TestTitleMissing(t *testing.T) {
	if got := tray.Title("zai", collect.LiveSession{}); got != "…" {
		t.Errorf("got %q, want …", got)
	}
}

func TestChatGPTTitleMalformed(t *testing.T) {
	cases := map[string]map[string]json.RawMessage{
		"nil":        nil,
		"empty":      {},
		"absent key": {"account": json.RawMessage(`{}`)},
		"garbage":    {"rate_limits": json.RawMessage(`{`)},
		"no windows": {"rate_limits": json.RawMessage(`{}`)},
	}
	for name, codex := range cases {
		if got := tray.ChatGPTTitle(codex); got != "c-" {
			t.Errorf("%s: got %q, want c-", name, got)
		}
	}
}

func TestChatGPTTitleWithRateLimits(t *testing.T) {
	// 잔여 최솟값 = 사용률 최댓값 창: primary 20.4% 사용 → 잔여 79.6 → 80
	codex := map[string]json.RawMessage{
		"rate_limits": json.RawMessage(`{"primary":{"used_percent":20.4},"secondary":{"used_percent":12.9}}`),
	}
	if got := tray.ChatGPTTitle(codex); got != "c80" {
		t.Errorf("got %q, want c80", got)
	}
}

func TestMenuLines(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.Local)
	live := collect.LiveSession{Zai: &collect.ZaiSnapshot{
		TokenRemaining: 0.87,
		TokenResetsAt:  now.Add(5 * time.Hour).UnixMilli(),
		McpRemaining:   0.45,
	}}
	lines := tray.MenuLines(live, "zai", now)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"zai 5h 87% · reset 17:00", "MCP 45%", "✓ zai", "○ chatgpt", "지금 갱신", "종료"} {
		if !strings.Contains(joined, want) {
			t.Errorf("menu missing %q, got:\n%s", want, joined)
		}
	}

	// 다른 날 리셋은 MM-DD 포맷
	live.Zai.TokenResetsAt = now.Add(48 * time.Hour).UnixMilli()
	if got := strings.Join(tray.MenuLines(live, "zai", now), "\n"); !strings.Contains(got, "09-29") {
		t.Errorf("cross-day reset should use MM-DD, got:\n%s", got)
	}

	// zai 부재 시 zai 라인 생략
	for _, l := range tray.MenuLines(collect.LiveSession{}, "zai", now) {
		if strings.Contains(l, "5h") || strings.Contains(l, "MCP") {
			t.Errorf("zai lines must be omitted when absent, got %q", l)
		}
	}
}
