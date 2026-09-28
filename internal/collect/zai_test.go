package collect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newZaiServer — quota/limit 목업. Authorization 헤더가 "test-token"이어야 200을 반환한다.
func newZaiServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/monitor/usage/quota/limit" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

const zaiBodyBoth = `{"limits":[
	{"type":"TOKENS_LIMIT","percentage":1,"unit":1,"nextResetTime":1761368400000},
	{"type":"TOKENS_LIMIT","percentage":10,"unit":6,"nextResetTime":1762923600000},
	{"type":"TIME_LIMIT","percentage":34,"currentValue":341,"usage":1000,"nextResetTime":1764644400000}
]}`

func TestZaiQuotaDirection(t *testing.T) {
	srv := newZaiServer(t, http.StatusOK, zaiBodyBoth)
	snap, err := Zai(context.Background(), srv.Client(), srv.URL, "test-token", "zai")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Provider != "zai" {
		t.Errorf("Provider = %q, want zai", snap.Provider)
	}
	if diff := snap.TokenRemaining - 0.99; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("TokenRemaining = %v, want 0.99 (1%% 사용 → 99%% 잔여)", snap.TokenRemaining)
	}
	if diff := snap.McpRemaining - 0.66; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("McpRemaining = %v, want 0.66 (34%% 사용 → 66%% 잔여)", snap.McpRemaining)
	}
	if snap.TokenResetsAt != 1761368400000 {
		t.Errorf("TokenResetsAt = %d, want 1761368400000", snap.TokenResetsAt)
	}
	if snap.McpResetsAt != 1764644400000 {
		t.Errorf("McpResetsAt = %d, want 1764644400000", snap.McpResetsAt)
	}
	if diff := snap.WeeklyRemaining - 0.90; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("WeeklyRemaining(unit=6) = %v, want 0.90", snap.WeeklyRemaining)
	}
	if snap.FetchedAt.IsZero() {
		t.Error("FetchedAt 미설정")
	}
}

func TestZaiEnvelopeData(t *testing.T) {
	srv := newZaiServer(t, http.StatusOK, `{"data":`+zaiBodyBoth+`}`)
	snap, err := Zai(context.Background(), srv.Client(), srv.URL, "test-token", "zai")
	if err != nil {
		t.Fatal(err)
	}
	if diff := snap.McpRemaining - 0.66; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("envelope {data:{limits}} 파싱 실패: McpRemaining = %v", snap.McpRemaining)
	}
}

func TestZaiLegacyNoUnit(t *testing.T) {
	// unit 필드 없음 + TOKENS_LIMIT 2개: nextResetTime 정렬로 5h/weekly 분류 (OMC legacy fallback)
	body := `{"limits":[
		{"type":"TOKENS_LIMIT","percentage":3,"nextResetTime":1761368400000},
		{"type":"TOKENS_LIMIT","percentage":10,"nextResetTime":1762923600000},
		{"type":"TIME_LIMIT","percentage":34,"nextResetTime":1764644400000}
	]}`
	srv := newZaiServer(t, http.StatusOK, body)
	snap, err := Zai(context.Background(), srv.Client(), srv.URL, "test-token", "zai")
	if err != nil {
		t.Fatal(err)
	}
	if diff := snap.TokenRemaining - 0.97; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("legacy 5h 분류 실패: TokenRemaining = %v, want 0.97", snap.TokenRemaining)
	}
}

func TestZaiMissingLimit(t *testing.T) {
	body := `{"limits":[{"type":"TOKENS_LIMIT","percentage":1,"nextResetTime":1761368400000}]}`
	srv := newZaiServer(t, http.StatusOK, body)
	_, err := Zai(context.Background(), srv.Client(), srv.URL, "test-token", "zai")
	if err == nil {
		t.Fatal("TIME_LIMIT 누락 시 오류여야 함")
	}
}

func TestZaiHTTPError(t *testing.T) {
	srv := newZaiServer(t, http.StatusForbidden, `{}`)
	_, err := Zai(context.Background(), srv.Client(), srv.URL, "test-token", "zai")
	if err == nil {
		t.Fatal("HTTP 403에서 오류여야 함")
	}
}

func TestZaiInputGuards(t *testing.T) {
	srv := newZaiServer(t, http.StatusOK, zaiBodyBoth)
	if _, err := Zai(context.Background(), srv.Client(), srv.URL, "", "zai"); err == nil {
		t.Error("토큰 부재 시 오류여야 함")
	}
	if _, err := Zai(context.Background(), srv.Client(), "", "test-token", "zai"); err == nil {
		t.Error("baseURL 부재 시 오류여야 함")
	}
	if _, err := Zai(context.Background(), srv.Client(), srv.URL, "test-token", ""); err == nil {
		t.Error("provider 부재(SSRF 가드) 시 오류여야 함")
	}
}
