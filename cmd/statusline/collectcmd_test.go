package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseProviderArg(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{"없음", nil, "", nil},
		{"등호 형식", []string{"--provider=zai"}, "zai", nil},
		{"빈 값", []string{"--provider="}, "", nil},
		{"중복 지정 마지막 우선", []string{"--provider=zai", "--provider=chatgpt"}, "chatgpt", nil},
		{"단독 토큰", []string{"--provider", "zai"}, "", errProviderUsage},
		{"다른 인자 무시", []string{"extra", "--provider=zai"}, "zai", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseProviderArg(tc.args)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSelectProviders(t *testing.T) {
	all := []string{"chatgpt", "zai"}
	cases := []struct {
		name           string
		provider       string
		want           []string
		wantErr        error
		wantUnknownErr bool
	}{
		{"all", "all", all, nil, false},
		{"미지정 = all", "", all, nil, false},
		{"chatgpt", "chatgpt", []string{"chatgpt"}, nil, false},
		{"zai", "zai", []string{"zai"}, nil, false},
		{"zhipu 별칭", "zhipu", []string{"zai"}, nil, false},
		{"anthropic 미지원", "anthropic", nil, errNotYetSupported, false},
		{"gemini 미지원", "gemini", nil, errNotYetSupported, false},
		{"알 수 없는 값", "foo", nil, nil, true},
		{"openai 미허용", "openai", nil, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectProviders(tc.provider)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}
			if tc.wantUnknownErr {
				if err == nil {
					t.Fatal("expected unknown-provider error, got nil")
				}
				if errors.Is(err, errNotYetSupported) {
					t.Fatalf("expected unknown-provider error, got not-yet-supported: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", got, tc.want)
			}
		})
	}
}

// zaiMockBody — quota/limit 정규 파싱이 가능한 응답 형상.
// internal/collect/zai_test.go zaiBodyBoth와 동일 (플랜의 planUsage 형태는 파서가 소화하지 못함).
const zaiMockBody = `{"limits":[
	{"type":"TOKENS_LIMIT","percentage":1,"unit":1,"nextResetTime":1761368400000},
	{"type":"TOKENS_LIMIT","percentage":10,"unit":6,"nextResetTime":1762923600000},
	{"type":"TIME_LIMIT","percentage":34,"currentValue":341,"usage":1000,"nextResetTime":1764644400000}
]}`

// newZaiMockServer — Z.AI quota/limit API 응답 목 (internal/collect/zai_test.go 패턴).
func newZaiMockServer(t *testing.T) *httptest.Server {
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
		_, _ = w.Write([]byte(zaiMockBody))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// stubZaiProvider — httptest 로컬 URL은 adapter allowlist(SSRF 가드)에 없으므로
// provider 해석만 주입. 프로덕션 기본값은 adapter.ProviderFromBaseURL 그대로다.
func stubZaiProvider(t *testing.T) {
	t.Helper()
	prev := zaiProviderFromBaseURL
	zaiProviderFromBaseURL = func(string) string { return "zai" }
	t.Cleanup(func() { zaiProviderFromBaseURL = prev })
}

// writeCreds — 테스트용 0600 credentials.json 기록
func writeCreds(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCollectAllMergesZai(t *testing.T) {
	stubZaiProvider(t)
	srv := newZaiMockServer(t)
	t.Setenv("ZAI_BASE_URL", srv.URL)
	t.Setenv("ZAI_AUTH_TOKEN", "test-token")
	t.Setenv("OPENAI_ADMIN_KEY", "") // platform 실망 호출 차단 — 테스트 상동성
	tmp := t.TempDir()
	creds := writeCreds(t, tmp, `{}`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := runCollect(&out, &errOut, []string{"chatgpt", "zai"}, creds, filepath.Join(tmp, "cache"))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr: %s", code, errOut.String())
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v — %s", err, out.String())
	}
	if _, ok := got["zai"]; !ok {
		t.Errorf("expected zai key in output, got: %s", out.String())
	}
	// codex는 로컬 환경 의존 — 키 존재 또는 errors.codex 기록 둘 중 하나
	if _, ok := got["codex"]; !ok {
		if errs, eok := got["errors"]; !eok || !bytes.Contains(errs, []byte("codex")) {
			t.Errorf("expected codex key or errors.codex, got: %s", out.String())
		}
	}
}

func TestRunCollectAllOmitsZaiWithoutCreds(t *testing.T) {
	t.Setenv("ZAI_BASE_URL", "")
	t.Setenv("ZAI_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("OPENAI_ADMIN_KEY", "") // platform 실망 호출 차단 — 테스트 상동성
	tmp := t.TempDir()
	creds := writeCreds(t, tmp, `{}`) // 자격 증명 없음

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := runCollect(&out, &errOut, []string{"chatgpt", "zai"}, creds, filepath.Join(tmp, "cache"))
	if code != 0 {
		t.Fatalf("기본 모드 자격 증명 부재는 exit 0이어야 함, got %d", code)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := got["zai"]; ok {
		t.Errorf("zai 키는 생략되어야 함, got: %s", out.String())
	}
}

func TestRunCollectZaiStandaloneMissingCredsExit1(t *testing.T) {
	t.Setenv("ZAI_BASE_URL", "")
	t.Setenv("ZAI_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	tmp := t.TempDir()
	creds := writeCreds(t, tmp, `{}`)

	var out, errOut bytes.Buffer
	code := runCollect(&out, &errOut, []string{"zai"}, creds, filepath.Join(tmp, "cache"))
	if code != 1 {
		t.Fatalf("zai 단독 자격 증명 부재는 exit 1 (현행 계약), got %d", code)
	}
	if errOut.Len() == 0 {
		t.Error("stderr에 오류 메시지 필요")
	}
}

// 자격 증명은 있으나 HTTP 실패 — allowlist 거부가 아닌 실제 조회 실패 경로를
// 검증하기 위해 provider 주입 후 500 서버로 실행한다.
func TestRunCollectZaiStandaloneFetchFailExit0(t *testing.T) {
	stubZaiProvider(t)
	// 즉시 500을 반환하는 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("ZAI_BASE_URL", srv.URL)
	t.Setenv("ZAI_AUTH_TOKEN", "test-token")
	tmp := t.TempDir()
	creds := writeCreds(t, tmp, `{}`)

	var out, errOut bytes.Buffer
	code := runCollect(&out, &errOut, []string{"zai"}, creds, filepath.Join(tmp, "cache"))
	if code != 0 {
		t.Fatalf("zai 단독 조회 실패는 errors JSON + exit 0 (exit 1 아님), got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"errors"`)) || !bytes.Contains(out.Bytes(), []byte("zai")) {
		t.Errorf("expected errors.zai in stdout, got: %s", out.String())
	}
}
