// Package tray — 트레이 표시 계층의 순수 함수들. systray 의존 없이 단위 테스트된다.
package tray

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"statusline/internal/collect"
)

// Title — 메뉴바 타이틀. primary 무효값(빈 값 포함)은 zai 폴백 (스펙: 경고 없음).
// 캐시 부재·zai 부재는 "…".
func Title(primary string, live collect.LiveSession) string {
	if primary == "chatgpt" {
		return ChatGPTTitle(live.Codex)
	}
	if live.Zai == nil {
		return "…"
	}
	title := "z" + pct(live.Zai.TokenRemaining)
	if live.Zai.WeeklyRemaining > 0 { // omitempty — 0%와 부재를 구분하지 않고 미표시로 축소
		title += " w" + pct(live.Zai.WeeklyRemaining)
	}
	return title
}

type rateWindow struct {
	UsedPercent float64 `json:"used_percent"`
}

// ChatGPTTitle — codex 스냅샷 rate_limits에서 잔여 최솟값(=사용률 최댓값 창)을 표시.
// 스키마는 실측 전 Unverified(스펙) — 미제공·파손·창 부재는 전부 "c-"로 흡수한다.
func ChatGPTTitle(codex map[string]json.RawMessage) string {
	raw, ok := codex["rate_limits"]
	if !ok {
		return "c-"
	}
	var rl struct {
		Primary   *rateWindow `json:"primary"`
		Secondary *rateWindow `json:"secondary"`
	}
	if json.Unmarshal(raw, &rl) != nil {
		return "c-"
	}
	used := -1.0
	for _, w := range []*rateWindow{rl.Primary, rl.Secondary} {
		if w != nil && w.UsedPercent > used {
			used = w.UsedPercent
		}
	}
	if used < 0 {
		return "c-"
	}
	remain := 100 - used
	if remain < 0 {
		remain = 0
	}
	return "c" + strconv.Itoa(int(math.Round(remain)))
}

// pct — 잔여율(0.0~1.0) → 퍼센트 반올림 문자열 (clamp 0~100)
func pct(v float64) string {
	p := int(math.Round(v * 100))
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	return strconv.Itoa(p)
}

// resetAt — Unix 밀리초 리셋 시각. 오늘이면 HH:MM, 다른 날은 MM-DD. 0은 빈 문자열.
func resetAt(ms int64, now time.Time) string {
	if ms <= 0 {
		return ""
	}
	at := time.UnixMilli(ms).Local()
	if y, m, d := at.Date(); y == now.Year() && m == now.Month() && d == now.Day() {
		return at.Format("15:04")
	}
	return at.Format("01-02")
}

// ZaiQuotaLine — zai 잔여 라인 본문("87% · reset 17:00"). fiveH=true → 5h 창, false → MCP 월간.
// snap이 nil이면 "데이터 없음" (traycmd 고정 슬롯 플레이스홀더).
func ZaiQuotaLine(snap *collect.ZaiSnapshot, now time.Time, fiveH bool) string {
	if snap == nil {
		return "데이터 없음"
	}
	remain, ms := snap.McpRemaining, snap.McpResetsAt
	if fiveH {
		remain, ms = snap.TokenRemaining, snap.TokenResetsAt
	}
	line := pct(remain) + "%"
	if r := resetAt(ms, now); r != "" {
		line += " · reset " + r
	}
	return line
}

// MenuLines — 드롭다운 라인. 부재 provider는 라인 생략(스펙: 실패·미수집 구분 없음).
func MenuLines(live collect.LiveSession, primary string, now time.Time) []string {
	var lines []string
	if live.Zai != nil {
		lines = append(lines,
			"zai 5h "+ZaiQuotaLine(live.Zai, now, true),
			"zai MCP "+ZaiQuotaLine(live.Zai, now, false),
		)
	}
	if len(live.Codex) > 0 {
		lines = append(lines, "chatgpt rate "+strings.TrimPrefix(ChatGPTTitle(live.Codex), "c"))
	}

	zaiMark, chatMark := "○", "○"
	if primary != "chatgpt" { // 무효값 포함 — zai가 기본
		zaiMark = "✓"
	} else {
		chatMark = "✓"
	}
	lines = append(lines,
		zaiMark+" zai (main)",
		chatMark+" chatgpt (main)",
		"지금 갱신",
		"종료",
	)
	return lines
}
