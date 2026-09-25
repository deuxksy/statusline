package collect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

// ZaiSnapshot — quota/limit 정규화 스냅샷. 비율은 전부 잔여(0.0~1.0), 리셋은 Unix 밀리초.
type ZaiSnapshot struct {
	FetchedAt       time.Time `json:"fetched_at"`
	Provider        string    `json:"provider"`
	TokenRemaining  float64   `json:"token_remaining_5h"`
	TokenResetsAt   int64     `json:"token_resets_at_ms"`
	McpRemaining    float64   `json:"mcp_remaining_month"`
	McpResetsAt     int64     `json:"mcp_resets_at_ms"`
	WeeklyRemaining float64   `json:"weekly_remaining,omitempty"` // unit=6 주간 버킷 — 기록 전용, 표시 미사용
	WeeklyResetsAt  int64     `json:"weekly_resets_at_ms,omitempty"`
}

// zaiUnitWeekly — 주간 TOKENS_LIMIT 버킷의 unit 코드 (OMC HUD 주석: observed, undocumented)
const zaiUnitWeekly = 6

type zaiLimit struct {
	Type          string  `json:"type"`
	Percentage    float64 `json:"percentage"`
	Unit          int     `json:"unit"`
	NextResetTime int64   `json:"nextResetTime"`
}

// Zai — GET {base}/api/monitor/usage/quota/limit 1회 조회 후 잔여율로 정규화.
// provider는 호출자가 adapter.ProviderFromBaseURL로 결정해 전달한다(""이면 거부 — SSRF 가드).
func Zai(ctx context.Context, client *http.Client, baseURL, authToken, provider string) (ZaiSnapshot, error) {
	var snap ZaiSnapshot
	if authToken == "" {
		return snap, errors.New("zai: ANTHROPIC_AUTH_TOKEN is not set")
	}
	if baseURL == "" {
		return snap, errors.New("zai: ANTHROPIC_BASE_URL is not set")
	}
	if provider == "" {
		return snap, errors.New("zai: unrecognized provider")
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" {
		return snap, fmt.Errorf("zai: invalid ANTHROPIC_BASE_URL %q", baseURL)
	}
	if client == nil {
		client = http.DefaultClient
	}
	endpoint := u.Scheme + "://" + u.Host + "/api/monitor/usage/quota/limit"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return snap, err
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Accept-Language", "en-US,en")
	resp, err := client.Do(req)
	if err != nil {
		return snap, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return snap, fmt.Errorf("quota/limit: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return snap, err
	}
	limits, err := parseZaiLimits(body)
	if err != nil {
		return snap, err
	}
	return normalizeZai(limits, provider, time.Now().UTC())
}

// parseZaiLimits — {limits:[...]}와 {data:{limits:[...]}} envelope 모두 수용
func parseZaiLimits(body []byte) ([]zaiLimit, error) {
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(body, &outer); err != nil {
		return nil, fmt.Errorf("quota/limit: %w", err)
	}
	limitsRaw := outer["limits"]
	if data, ok := outer["data"]; ok && limitsRaw == nil {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(data, &inner); err == nil {
			limitsRaw = inner["limits"]
		}
	}
	if limitsRaw == nil {
		return nil, errors.New("quota/limit: limits not found")
	}
	var limits []zaiLimit
	if err := json.Unmarshal(limitsRaw, &limits); err != nil {
		return nil, fmt.Errorf("quota/limit: %w", err)
	}
	return limits, nil
}

// normalizeZai — 버킷 분류와 잔여 변환. OMC parseZaiResponse 규칙 동일 적용:
// unit=6이 주간, 나머지 TOKENS_LIMIT 중 nextResetTime이 가장 빠른 것이 5시간 버킷.
func normalizeZai(limits []zaiLimit, provider string, now time.Time) (ZaiSnapshot, error) {
	var snap ZaiSnapshot
	var tokens []zaiLimit
	var timeLimit *zaiLimit
	for i := range limits {
		switch limits[i].Type {
		case "TOKENS_LIMIT":
			tokens = append(tokens, limits[i])
		case "TIME_LIMIT":
			timeLimit = &limits[i]
		}
	}
	if len(tokens) == 0 || timeLimit == nil {
		return snap, errors.New("quota/limit: missing TOKENS_LIMIT or TIME_LIMIT")
	}
	remaining := func(used float64) float64 {
		used = math.Max(0, math.Min(100, used))
		return 1 - used/100
	}
	snap = ZaiSnapshot{
		FetchedAt:    now,
		Provider:     provider,
		McpRemaining: remaining(timeLimit.Percentage),
		McpResetsAt:  timeLimit.NextResetTime,
	}
	var weekly *zaiLimit
	for i := range tokens {
		if tokens[i].Unit == zaiUnitWeekly {
			weekly = &tokens[i]
			break
		}
	}
	byReset := func(a, b zaiLimit) bool {
		if a.NextResetTime != b.NextResetTime {
			return resetKey(a.NextResetTime) < resetKey(b.NextResetTime)
		}
		return a.Percentage < b.Percentage
	}
	if weekly != nil {
		var rest []zaiLimit
		for _, tk := range tokens {
			if tk.Unit != zaiUnitWeekly {
				rest = append(rest, tk)
			}
		}
		if len(rest) > 0 {
			snap.TokenRemaining = remaining(minBy(rest, byReset).Percentage)
			snap.TokenResetsAt = minBy(rest, byReset).NextResetTime
		}
		snap.WeeklyRemaining = remaining(weekly.Percentage)
		snap.WeeklyResetsAt = weekly.NextResetTime
	} else {
		// legacy: unit 필드 없음 → 정렬 후 첫 번째가 5h, 두 번째가 주간
		sorted := sortBy(tokens, byReset)
		snap.TokenRemaining = remaining(sorted[0].Percentage)
		snap.TokenResetsAt = sorted[0].NextResetTime
		if len(sorted) > 1 {
			snap.WeeklyRemaining = remaining(sorted[1].Percentage)
			snap.WeeklyResetsAt = sorted[1].NextResetTime
		}
	}
	return snap, nil
}

func resetKey(nextResetTime int64) int64 {
	if nextResetTime > 0 {
		return nextResetTime
	}
	return math.MaxInt64
}

func minBy(xs []zaiLimit, less func(a, b zaiLimit) bool) zaiLimit {
	best := xs[0]
	for _, x := range xs[1:] {
		if less(x, best) {
			best = x
		}
	}
	return best
}

func sortBy(xs []zaiLimit, less func(a, b zaiLimit) bool) []zaiLimit {
	out := append([]zaiLimit(nil), xs...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && less(out[j], out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// ZaiCollectResult — statusline collect --provider=zai 출력 형태
type ZaiCollectResult struct {
	Zai    *ZaiSnapshot      `json:"zai,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

// RunZaiCollect — 조회 → 캐시 기록 → marker 제거까지의 CLI 플로우.
// 실패 시 기존 캐시·marker를 유지한다(재시도 백프레셔).
func RunZaiCollect(ctx context.Context, client *http.Client, baseURL, authToken, provider, cachePath, refreshPath string) ZaiCollectResult {
	var result ZaiCollectResult
	fail := func(msg string) ZaiCollectResult {
		return ZaiCollectResult{Errors: map[string]string{"zai": msg}}
	}
	if baseURL == "" {
		return fail("zai: ANTHROPIC_BASE_URL is not set")
	}
	if provider == "" {
		return fail("zai: unrecognized ANTHROPIC_BASE_URL (api.z.ai / bigmodel.cn만 지원)")
	}
	if authToken == "" {
		return fail("zai: ANTHROPIC_AUTH_TOKEN is not set")
	}
	snap, err := Zai(ctx, client, baseURL, authToken, provider)
	if err != nil {
		return fail(err.Error())
	}
	if err := WriteZaiCacheAt(cachePath, snap); err != nil {
		return fail(err.Error())
	}
	ClearZaiRefreshAt(refreshPath)
	result.Zai = &snap
	return result
}
