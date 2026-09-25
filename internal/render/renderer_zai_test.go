package render

import (
	"strings"
	"testing"
	"time"

	"statusline/internal/config"
	"statusline/internal/model"
)

// 기준 시각 고정 — 카운트다운 golden value 계산용
var zaiNow = time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

func TestFormatResetCountdown(t *testing.T) {
	cases := []struct {
		name   string
		offset time.Duration
		want   string
	}{
		{"hours-minutes", 4*time.Hour + 46*time.Minute, "4h46m"},
		{"days-hours", 18*24*time.Hour + 19*time.Hour, "18d19h"},
		{"exactly-24h", 24 * time.Hour, "1d0h"},
		{"under-1h", 59 * time.Minute, "0h59m"},
		{"past", -time.Minute, ""},
		{"zero", 0, ""},
	}
	for _, tc := range cases {
		unixMs := zaiNow.Add(tc.offset).UnixMilli()
		if got := formatResetCountdown(unixMs, zaiNow); got != tc.want {
			t.Errorf("%s: formatResetCountdown = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestFormatZaiQuota(t *testing.T) {
	cats := []model.QuotaCategory{{
		Name:            "zai",
		FiveH:           0.97,
		FiveHResetsAt:   zaiNow.Add(4*time.Hour + 46*time.Minute).UnixMilli(),
		Monthly:         0.66,
		MonthlyResetsAt: zaiNow.Add(18*24*time.Hour + 19*time.Hour).UnixMilli(),
	}}
	want := "zai 5h:97%(4h46m) mo:66%(18d19h)"
	if got := formatZaiQuota(cats, zaiNow); got != want {
		t.Errorf("formatZaiQuota = %q, want %q", got, want)
	}

	// 리셋 정보 없으면 카운트다운 생략
	plain := []model.QuotaCategory{{Name: "zai", FiveH: 0.97, Monthly: 0.66}}
	if got := formatZaiQuota(plain, zaiNow); got != "zai 5h:97% mo:66%" {
		t.Errorf("no-reset 형태 오류: %q", got)
	}

	// 0% 잔여도 유효값 — 생략 아님
	zero := []model.QuotaCategory{{Name: "zai", FiveH: 0, Monthly: 0}}
	if got := formatZaiQuota(zero, zaiNow); got != "zai 5h:0% mo:0%" {
		t.Errorf("zero-remaining 형태 오류: %q", got)
	}

	// 반올림: truncation이면 57이 나오는 값
	rounding := []model.QuotaCategory{{Name: "zhipu", FiveH: 0.5799, Monthly: 0.335}}
	if got := formatZaiQuota(rounding, zaiNow); got != "zhipu 5h:58% mo:34%" {
		t.Errorf("rounding 오류: %q", got)
	}

	if got := formatZaiQuota(nil, zaiNow); got != "" {
		t.Errorf("빈 입력: %q", got)
	}
}

func TestRenderZaiQuotaSegment(t *testing.T) {
	st := model.NewUnifiedStatus("claude")
	st.Provider = "zai"
	st.Capabilities.HasQuota = true
	st.Quota = []model.QuotaCategory{{
		Name:            "zai",
		FiveH:           0.97,
		FiveHResetsAt:   time.Now().Add(4 * time.Hour).UnixMilli(),
		Monthly:         0.66,
		MonthlyResetsAt: time.Now().Add(400 * time.Hour).UnixMilli(),
	}}
	cfg := &config.Config{
		Theme:    "sleek_dark",
		Elements: config.ElementsConfig{Quota: true},
		Layout:   config.LayoutConfig{Main: []string{"quota"}},
	}
	out := Render(st, cfg)
	// zai 경로: 5h/mo 라벨 존재, antigravity 전용 wk 라벨 부재
	for _, want := range []string{"zai", "5h:97%", "mo:66%"} {
		if !strings.Contains(out, want) {
			t.Errorf("출력에 %q 없음: %q", want, out)
		}
	}
	if strings.Contains(out, "wk:") {
		t.Errorf("zai 경로에 wk 라벨이 있음: %q", out)
	}
}

func TestRenderQuotaAntigravityRegression(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Quota = []model.QuotaCategory{{Name: "gemini", FiveH: 0.93, Weekly: 0.52}}
	cfg := &config.Config{
		Theme:    "sleek_dark",
		Elements: config.ElementsConfig{Quota: true},
		Layout:   config.LayoutConfig{Main: []string{"quota"}},
	}
	out := Render(st, cfg)
	if !strings.Contains(out, "wk:52%") {
		t.Errorf("antigravity weekly 렌더 회귀: %q", out)
	}
	if strings.Contains(out, "mo:") {
		t.Errorf("antigravity 경로에 mo 라벨이 있음: %q", out)
	}
}
