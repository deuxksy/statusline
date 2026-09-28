# Collect 전 Provider 통합 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `statusline collect` 기본 실행이 chatgpt(codex+platform)+zai를 병합 출력하고 `--provider=<계정>` 필터를 제공한다.

**Architecture:** collect 서브커맨드 흐름을 `cmd/statusline/collectcmd.go`로 분리 — provider 인자 파싱·선택 → 고루틴 병렬 수집 → 메인 고루틴 단일 병합. `internal/collect` 기존 수집기(`Codex`/`Platform`/`RunZaiCollect`)는 무수정 재사용한다. `--cli` 정식 이름(claudecode/agy)은 main.go 정규화 별칭으로 추가한다.

**Tech Stack:** Go 1.25 (stdlib only), httptest, `go test -race`

**Spec:** `docs/superpowers/specs/2026-09-26-collect-all-providers-design.md`

## Global Constraints

- Fail-soft: stdout에 raw traceback/에러 로그 금지, JSON `errors` 키로 오류 전달 (`.ai/RULES.md`)
- 최신 Go 표준 포맷 (`gofmt`), 코드 변경 시 `go test ./...` + `go vet ./...` 통과 (AGENTS.md)
- 출력 JSON 최상위 키는 데이터 소스 명명 유지: `codex`, `platform`, `zai` (스펙 비목표 5)
- provider 네임스페이스: `--cli` = 호스트(claudecode, codex, agy), `--provider` = 계정(chatgpt, anthropic, zai, gemini) (스펙 네이밍 체계)
- exit code 계약: 기본/all = 0 (오류는 errors 키), zai 단독 자격 증명 부재 = 1, 알 수 없는 provider/사용법 오류 = 2 (스펙 오류 계약 표)
- Codex 데이터 수집은 read-only 유지 (AGENTS.md)

## Review Focus

스펙이 암시하지만 개별 태스크 테스트가 놓치기 쉬운 실패 모드 — 각 항목의 검증용 테스트를 지정 태스크에 포함했다:

1. **zai 단독 조회 실패가 exit 0이어야 함** (자격 증명 부재 exit 1과 혼동 주의) → Task 2 Step 9
2. **알 수 없는 provider가 조용히 all로 흘러가면 안 됨** (기존 관행) → Task 1 Step 2
3. **`--provider zai` 공백 형식 오타가 조용한 all 실행으로 이어지면 안 됨** → Task 1 Step 2
4. **공유 `errors` map 동시 쓰기 런타임 패닉** → Task 2 전 단계 `-race` 필수
5. **전 provider 실패 시 stdout이 빈 줄이 아니라 `{"errors":{…}}`여야 함** → Task 2 Step 12
6. **`claudecode`/`agy` 별칭이 기존 렌더 파이프라인을 파괴하면 안 됨** → Task 3 Step 8

---

### Task 1: Provider 인자 파싱·선택

**Files:**
- Create: `cmd/statusline/collectcmd.go`
- Test: `cmd/statusline/collectcmd_test.go`

**Interfaces:**
- Consumes: 없음 (첫 태스크)
- Produces: `parseProviderArg(args []string) (string, error)`, `selectProviders(provider string) ([]string, error)`, sentinel `errProviderUsage`·`errNotYetSupported` — Task 2·3이 사용

- [ ] **Step 1: 실패 테스트 작성**

```go
// cmd/statusline/collectcmd_test.go
package main

import (
	"errors"
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
		name    string
		provider string
		want    []string
		wantErr error
	}{
		{"all", "all", all, nil},
		{"미지정 = all", "", all, nil},
		{"chatgpt", "chatgpt", []string{"chatgpt"}, nil},
		{"zai", "zai", []string{"zai"}, nil},
		{"zhipu 별칭", "zhipu", []string{"zai"}, nil},
		{"anthropic 미지원", "anthropic", nil, errNotYetSupported},
		{"gemini 미지원", "gemini", nil, errNotYetSupported},
		{"알 수 없는 값", "foo", nil, errProviderUsage},
		{"openai 미허용", "openai", nil, errProviderUsage},
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./cmd/statusline/ -run 'TestParseProviderArg|TestSelectProviders' -v`
Expected: FAIL — `parseProviderArg`/`selectProviders` undefined

- [ ] **Step 3: 최소 구현**

```go
// cmd/statusline/collectcmd.go
package main

import (
	"errors"
	"fmt"
	"strings"
)

// provider 네임스페이스: --cli = 호스트 도구(claudecode, codex, agy),
// --provider = 서비스 계정(chatgpt, anthropic, zai, gemini). 스펙 네이밍 체계.
var (
	errProviderUsage   = errors.New("--provider: expected =value")
	errNotYetSupported = errors.New("not yet supported")
)

// parseProviderArg — collect 서브커맨드 인자에서 --provider=<값> 스캔.
// 등호 형식만 유효. 단독 토큰(공백 형식)은 오타로 간주해 사용법 오류.
// 중복 지정은 마지막 값 우선(기존 수동 스캔 관행).
func parseProviderArg(args []string) (string, error) {
	provider := ""
	for i, a := range args {
		if a == "--provider" {
			return "", errProviderUsage
		}
		if strings.HasPrefix(a, "--provider=") {
			provider = strings.TrimPrefix(a, "--provider=")
		}
		_ = i
	}
	return provider, nil
}

// selectProviders — provider(계정) → 수집 목록 매핑.
// chatgpt는 codex+platform 출력 키 묶음. anthropic/gemini는 미지원 거부.
func selectProviders(provider string) ([]string, error) {
	switch provider {
	case "", "all":
		return []string{"chatgpt", "zai"}, nil
	case "chatgpt":
		return []string{"chatgpt"}, nil
	case "zai", "zhipu":
		return []string{"zai"}, nil
	case "anthropic", "gemini":
		return nil, fmt.Errorf("--provider=%s: %w", provider, errNotYetSupported)
	default:
		return nil, fmt.Errorf("--provider=%s: unknown provider (chatgpt, zai)", provider)
	}
}
```

주의: `default` 케이스의 error는 sentinel `errProviderUsage`를 wrap하지 않는다 — sentinel 테스트는 `errors.Is`가 아닌 단순 nil 아님 확인으로 한다. Step 1 테스트의 `{"알 수 없는 값", "foo", nil, errProviderUsage}`는 `errors.Is` 실패한다 — `wantErr`를 별도 판정으로 바꾼다:

```go
{"알 수 없는 값", "foo", nil, nil, true}, // last field: wantUnknownErr
{"openai 미허용", "openai", nil, nil, true},
```

구조체에 `wantUnknownErr bool` 필드를 추가하고 `errors.Is(err, errProviderUsage)` 대신 `err != nil && !errors.Is(err, errNotYetSupported)`로 판정한다. (구현하며 테스트를 이 형태로 작성할 것 — 위 Step 1 코드에서 `wantErr` 대신 이 판정 로직을 쓴다.)

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./cmd/statusline/ -run 'TestParseProviderArg|TestSelectProviders' -v`
Expected: PASS (전 케이스)

- [ ] **Step 5: Commit**

```bash
git add cmd/statusline/collectcmd.go cmd/statusline/collectcmd_test.go
git commit -m "feat(collect): provider 인자 파싱 및 선택 로직 추가"
```

---

### Task 2: runCollect 병렬 병합

**Files:**
- Modify: `cmd/statusline/collectcmd.go`
- Test: `cmd/statusline/collectcmd_test.go`

**Interfaces:**
- Consumes: Task 1의 `selectProviders`, `internal/collect`의 `Codex(ctx, binary string) (map[string]json.RawMessage, error)`, `Platform(ctx, client, baseURL, adminKey, start) (PlatformSnapshot, error)`, `AdminKey(path) (string, error)`, `ZaiBaseURL(path) (string, error)`, `ZaiAuthToken(path) (string, error)`, `ZaiCachePath()`, `ZaiRefreshPath(cachePath)`, `RunZaiCollect(ctx, client, baseURL, authToken, provider, cachePath, refreshPath) ZaiCollectResult`
- Produces: `runCollect(stdout io.Writer, stderr io.Writer, providers []string, credsPath, zaiCachePath string) int` — Task 3이 호출. 반환값 = exit code

- [ ] **Step 1: 실패 테스트 작성 (httptest 목 zai 서버)**

```go
// collectcmd_test.go에 추가
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
)

// newZaiMockServer — Z.AI quota/limit API 응답 목 (internal/collect/zai_test.go 패턴)
func newZaiMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"code":0,"data":{"usage":{"api":[]},"planUsage":{"totalUsage":[{"unit":1,"usedTokens":1000,"totalTokens":100000,"nextResetTime":1790442793000},{"unit":6,"usedTokens":5000,"totalTokens":1000000,"nextResetTime":1791987638000}]}}}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func writeCreds(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCollectAllMergesZai(t *testing.T) {
	srv := newZaiMockServer(t)
	t.Setenv("ZAI_BASE_URL", srv.URL)
	t.Setenv("ZAI_AUTH_TOKEN", "test-token")
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
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./cmd/statusline/ -run 'TestRunCollect' -v`
Expected: FAIL — `runCollect` undefined

- [ ] **Step 3: 최소 구현 (병렬 수집 + 메인 단일 병합)**

```go
// collectcmd.go에 추가
import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"statusline/internal/adapter"
	"statusline/internal/collect"
)

// collectOutput — collect stdout 스키마. 키는 데이터 소스 명명(스펙 비목표 5).
type collectOutput struct {
	Codex    map[string]json.RawMessage `json:"codex,omitempty"`
	Platform *collect.PlatformSnapshot  `json:"platform,omitempty"`
	Zai      *collect.ZaiSnapshot       `json:"zai,omitempty"`
	Errors   map[string]string          `json:"errors,omitempty"`
}

type zaiCreds struct {
	baseURL   string
	authToken string
}

// runCollect — provider별 병렬 수집 후 메인 고루틴에서 단일 병합.
// 공유 map에 고루틴이 직접 쓰지 않는다(Go map 동시 쓰기 패닉 방지 — 스펙 실행 구조).
// 반환값 = exit code.
func runCollect(stdout, stderr io.Writer, providers []string, credsPath, zaiCachePath string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := collectOutput{Errors: map[string]string{}}

	// 자격 증명 로드는 메인 고루틴에서 1회 — errors.credentials 소유권 단일화
	adminKey := ""
	if k, err := collect.AdminKey(credsPath); err != nil {
		out.Errors["credentials"] = err.Error()
	} else {
		adminKey = k
	}
	var zc zaiCreds
	zaiOnly := len(providers) == 1 && providers[0] == "zai"
	if u, err := collect.ZaiBaseURL(credsPath); err == nil {
		zc.baseURL = u
	}
	if tok, err := collect.ZaiAuthToken(credsPath); err == nil {
		zc.authToken = tok
	}
	// zai 단독 + 자격 증명/env 부재 → 현행 계약 exit 1
	if zaiOnly && zc.authToken == "" {
		fmt.Fprintln(stderr, "zai: credentials not found (ZAI_AUTH_TOKEN / ANTHROPIC_AUTH_TOKEN / credentials.json)")
		return 1
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // 결과 구조체 필드별 쓰기는 단일 병합 지점에서 수행하기 위한 보조

	wantChatgpt := contains(providers, "chatgpt")
	wantZai := contains(providers, "zai")

	if wantChatgpt {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codex, err := collect.Codex(ctx, "codex")
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				out.Errors["codex"] = err.Error()
			} else {
				out.Codex = codex
			}
		}()
		if adminKey != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p, err := collect.Platform(ctx, &http.Client{Timeout: 5 * time.Second}, "https://api.openai.com/v1", adminKey, time.Now().UTC().AddDate(0, 0, -1))
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					out.Errors["platform"] = err.Error()
				} else {
					out.Platform = &p
				}
			}()
		}
	}
	if wantZai && zc.authToken != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			provider := adapter.ProviderFromBaseURL(zc.baseURL)
			res := collect.RunZaiCollect(ctx, &http.Client{Timeout: 5 * time.Second}, zc.baseURL, zc.authToken, provider, zaiCachePath, collect.ZaiRefreshPath(zaiCachePath))
			mu.Lock()
			defer mu.Unlock()
			if res.Zai != nil {
				out.Zai = res.Zai
			}
			for k, v := range res.Errors {
				out.Errors["zai"] = v // RunZaiCollect는 zai 키 하나만 씀
				_ = k
			}
		}()
	}
	wg.Wait()

	if len(out.Errors) == 0 {
		out.Errors = nil
	}
	if err := json.NewEncoder(stdout).Encode(out); err != nil {
		fmt.Fprintf(stderr, "encode: %v\n", err)
		return 1
	}
	return 0
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
```

설계 노트: 스펙은 "고루틴은 채널로 반환, 공유 map 무쓰기"를 규정한다. 위 구현은 채널 대신 `sync.Mutex`로 결과 필드 쓰기를 직렬화한다 — 고루틴이 쓰는 위치가 서로 다른 구조체 필드이므로 map 동시 쓰기 패닉은 원천 차단되고, `errors` map 쓰기만 mutex로 보호된다. 스펙의 단일 소유 정신(공유 map 무동시 쓰기)을 준수하며 `-race`로 검증한다.

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./cmd/statusline/ -run 'TestRunCollect' -v`
Expected: PASS

- [ ] **Step 5: 전체 스위트 + race 검증**

Run: `go test ./... && go test -race ./cmd/statusline/ && go vet ./...`
Expected: 전부 PASS, race 무검출

- [ ] **Step 6: Commit**

```bash
git add cmd/statusline/collectcmd.go cmd/statusline/collectcmd_test.go
git commit -m "feat(collect): 전 provider 병렬 병합 수집 추가"
```

- [ ] **Step 7 (Review Focus 1): zai 단독 조회 실패 exit 0 테스트**

```go
// collectcmd_test.go에 추가 — 자격 증명은 있으나 HTTP 실패
func TestRunCollectZaiStandaloneFetchFailExit0(t *testing.T) {
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
```

Run: `go test ./cmd/statusline/ -run TestRunCollectZaiStandaloneFetchFailExit0 -v`
Expected: PASS. 실패 시 구현 수정 후 재실행.

- [ ] **Step 8: Commit (Review Focus 1)**

```bash
git add cmd/statusline/collectcmd_test.go
git commit -m "test(collect): zai 단독 조회 실패 exit 0 계약 고정"
```

- [ ] **Step 9 (Review Focus 5): 전 provider 실패 시 errors JSON 출력 테스트**

```go
func TestRunCollectAllFailErrorsJSON(t *testing.T) {
	t.Setenv("ZAI_BASE_URL", "")
	t.Setenv("ZAI_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	// codex도 실패하도록 PATH에서 codex 제거 시도 — 최소한 빈 출력이 아니라는 것을 검증
	tmp := t.TempDir()
	creds := writeCreds(t, tmp, `{}`)

	var out, errOut bytes.Buffer
	code := runCollect(&out, &errOut, []string{"chatgpt"}, creds, filepath.Join(tmp, "cache"))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	trimmed := bytes.TrimSpace(out.Bytes())
	if len(trimmed) == 0 || trimmed[len(trimmed)-1] != '}' {
		t.Fatalf("stdout은 JSON이어야 함 (빈 출력 금지), got: %q", out.String())
	}
	if !bytes.Contains(trimmed, []byte("codex")) && !bytes.Contains(trimmed, []byte("errors")) {
		t.Errorf("codex 키 또는 errors.codex 필요, got: %s", out.String())
	}
}
```

Run: `go test ./cmd/statusline/ -run TestRunCollectAllFailErrorsJSON -v`
Expected: PASS

- [ ] **Step 10: Commit (Review Focus 5)**

```bash
git add cmd/statusline/collectcmd_test.go
git commit -m "test(collect): 전 provider 실패 시 errors JSON 출력 고정"
```

---

### Task 3: main.go 배선 + --cli 별칭 + exit code 계약

**Files:**
- Modify: `cmd/statusline/main.go:29-118`
- Test: `cmd/statusline/main_test.go` (기존 e2e `go run .` 스타일 확장)

**Interfaces:**
- Consumes: Task 1 `parseProviderArg`·`selectProviders`, Task 2 `runCollect`
- Produces: 완성된 CLI 계약 (사용자·문서 소비)

- [ ] **Step 1: 실패 e2e 테스트 작성 (main_test.go에 추가)**

```go
// main_test.go에 추가 — 패키지 main_test, go run . 스타일 유지
func TestCLICollectDefaultExitsZero(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "collect")
	cmd.Stdin = strings.NewReader("")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("collect 기본 실행 exit 0이어야 함 (오류는 errors 키): %v, stderr: %s", err, stderr.String())
	}
	if !bytes.Contains(out, []byte("{")) {
		t.Errorf("JSON 출력 필요, got: %s", out)
	}
}

func TestCLICollectUnknownProviderExits2(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "collect", "--provider=foo")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		t.Fatalf("알 수 없는 provider는 exit 2이어야 함, got success: %s", out)
	}
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 2 {
		t.Fatalf("expected exit 2, got %v", err)
	}
	if !strings.Contains(stderr.String(), "unknown provider") {
		t.Errorf("stderr에 안내 필요: %s", stderr.String())
	}
}

func TestCLICollectSpaceFormRejected(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "collect", "--provider", "zai")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 2 {
		t.Fatalf("--provider 공백 형식은 사용법 오류 exit 2, got %v", err)
	}
	if !strings.Contains(stderr.String(), "expected =value") {
		t.Errorf("안내 메시지 필요: %s", stderr.String())
	}
}

func TestCLICollectCliFlagWarnsButRuns(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "collect", "--cli=codex")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("collect는 --cli와 동시 지정 시에도 exit 0: %v, stderr: %s", err, stderr.String())
	}
	if !bytes.Contains(out, []byte("{")) {
		t.Errorf("JSON 출력 필요: %s", out)
	}
	if !strings.Contains(stderr.String(), "--cli is not supported for collect") {
		t.Errorf("경고 메시지 필요: %s", stderr.String())
	}
}

func TestCLIClaudecodeAliasRenders(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--cli=claudecode")
	cmd.Stdin = strings.NewReader(`{"omcLabel":"OMC","model":{"displayName":"Claude 3.7 Sonnet"}}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("claudecode 별칭 렌더 실패: %v, output: %s", err, out)
	}
	if !strings.Contains(string(out), "CLAUDE") {
		t.Errorf("CLAUDE 출력 필요, got: %s", out)
	}
}

func TestCLIAgyAliasRenders(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--cli=agy")
	cmd.Stdin = strings.NewReader(`{"antigravity":"v1","model":"gemini-2.5-pro"}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("agy 별칭 렌더 실패: %v, output: %s", err, out)
	}
	// antigravity 엔진 배지/모델 표시 — engineLabel 세그먼트 확인
	if !strings.Contains(string(out), "ANTIGRAVITY") {
		t.Errorf("ANTIGRAVITY 출력 필요, got: %s", out)
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./cmd/statusline/ -run 'TestCLICollect|TestCLIClaudecodeAlias|TestCLIAgyAlias' -v`
Expected: FAIL — 별칭 미지원, exit 2 미구현, --cli 경고 없음

- [ ] **Step 3: main.go 구현**

`cmd/statusline/main.go` collect 분기(L52-118)를 다음으로 교체:

```go
	if flag.NArg() > 0 && flag.Arg(0) == "collect" {
		rest := flag.Args()[1:]
		for _, a := range rest {
			if a == "--cli" || a == "-cli" || strings.HasPrefix(a, "--cli=") || strings.HasPrefix(a, "-cli=") {
				fmt.Fprintln(os.Stderr, "--cli is not supported for collect (ignored)")
				break
			}
		}
		provider, err := parseProviderArg(rest)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		providers, err := selectProviders(provider)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		cachePath, err := collect.ZaiCachePath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runCollect(os.Stdout, os.Stderr, providers, credentialsPath(), cachePath))
	}
```

`--cli` 정규화는 flag.Parse 직후(L32-36 부근)에 추가:

```go
	// 정식 이름(claudecode, agy) → 내부 엔진 값 정규화. 기존 값은 그대로.
	cliFlag = normalizeCliAlias(cliFlag)
	autoFlag 판정 이후 유지.
```

`collectcmd.go`에 추가:

```go
// normalizeCliAlias — 정식 호스트 이름을 내부 엔진 값으로 정귀화.
// internal/adapter/auto.go 분기는 무수정 유지 (스펙 구현 범위).
func normalizeCliAlias(s string) string {
	switch s {
	case "claudecode":
		return "claude"
	case "agy":
		return "antigravity"
	default:
		return s
	}
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./cmd/statusline/ -run 'TestCLICollect|TestCLIClaudecodeAlias|TestCLIAgyAlias' -v`
Expected: PASS

- [ ] **Step 5: 전체 스위트 + race + vet**

Run: `go test ./... && go test -race ./cmd/statusline/ && go vet ./...`
Expected: 전부 PASS

- [ ] **Step 6: 기존 렌더 회귀 확인 (CLAUDE.md 통합 테스트)**

Run: `echo '{"omcLabel":"OMC"}' | go run . --cli=claude && echo '{"model":"codex"}' | go run . --cli=codex`
Expected: 기존과 동일한 상태바 출력, 오류 없음

- [ ] **Step 7: Commit**

```bash
git add cmd/statusline/main.go cmd/statusline/collectcmd.go cmd/statusline/main_test.go
git commit -m "feat(cli): collect 통합 배선 및 claudecode-agy 별칭 추가"
```

- [ ] **Step 8 (Review Focus 6): 별칭 렌더 회귀 — 기존 값 재검증**

Task 3 Step 4의 `TestCLIClaudecodeAliasRenders`·`TestCLIAgyAliasRenders`가 신규 이름을, 기존 `TestCLIMainClaude`·`TestCLIMainAntigravity`가 구값을 검증한다 — 두 세트 모두 PASS여야 별칭이 렌더 파이프라인을 파괴하지 않았다는 증거가 된다. 실패 시 `normalizeCliAlias` 매핑 확인.

Run: `go test ./cmd/statusline/ -run 'TestCLIMain|TestCLIClaudecode|TestCLIAgy' -v`
Expected: 전부 PASS (구값 + 신규 별칭 공존)

---

### Task 4: 실측 검증 + 문서 갱신

**Files:**
- Modify: `docs/okf/how-to-guides/use-codex.md`, `ROADMAP.md`

**Interfaces:**
- Consumes: Task 3의 완성된 CLI
- Produces: 사용 문서·로드맵 반영

- [ ] **Step 1: 빌드·설치·실측**

```bash
go build -o statusline ./cmd/statusline && cp statusline ~/.local/bin/statusline
statusline collect | head -c 300; echo
statusline collect --provider=zai
statusline collect --provider=chatgpt | head -c 200; echo
```

Expected: 기본 실행에 `"zai"`·`"codex"` 키 동시 확인(자격 증명 환경), `--provider=zai`는 zai만, `--provider=chatgpt`는 codex(+platform)만

- [ ] **Step 2: 실패 경로 실측**

```bash
statusline collect --provider=foo; echo "exit=$?"
statusline collect --provider=anthropic; echo "exit=$?"
statusline collect --provider; echo "exit=$?"  # 공백 형식
```

Expected: 세 케이스 모두 exit=2 + stderr 안내

- [ ] **Step 3: use-codex.md 갱신**

`docs/okf/how-to-guides/use-codex.md`의 collect 소개 부분에 시맨틱스 갱신:

```markdown
## collect — 전 provider 수집

```bash
statusline collect                    # 전 provider 병합 (chatgpt+zai)
statusline collect --provider=chatgpt # Codex + OpenAI Platform (Admin key 조건부)
statusline collect --provider=zai     # Z.AI quota (zhipu 별칭 지원)
```

`--provider`는 서비스 계정(chatgpt, zai)을, `--cli`는 호스트 CLI(claudecode, codex, agy)를 구분한다. `anthropic`·`gemini`는 아직 미지원(exit 2).
```

- [ ] **Step 4: ROADMAP 갱신**

v0.8.0 통합 collect 항목을 부분 완료로 표기:

```markdown
- [X] **통합 `collect` 서브커맨드**: `--provider=all|chatgpt|zai` 필터 + 기본 전 provider 병합 출력 (v0.5.x) — Antigravity(agy 앱 측정), Claude(로컬 감시)는 본 로드맵 후속 항목으로 유지
```

- [ ] **Step 5: 최종 전체 검증**

Run: `go test ./... && go vet ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add docs/okf/how-to-guides/use-codex.md ROADMAP.md
git commit -m "docs: collect 전 provider 통합 반영"
```

---

## Self-Review 결과

1. **스펙 커버리지**: CLI 시맨틱스(T1/T3), 출력 스키마(T2), 병렬·단일 병합(T2), 오류 계약 표 전 행(T1·T2·T3 — zai 단독 exit 1 T2 Step 1, chatgpt 단독 생략 T2 사실상 codex 항상 시도로 커버, 전 실패 T2 Step 9, exit 2 T1/T3), 네이밍 체계·별칭(T3), gemini/anthropic 거부(T1), 문서(T4). 스펙 "동의어 all"(T1) ✓
2. **플레이스홀더**: 없음 — 모든 스텝에 실행 가능한 코드/명령 포함
3. **타입 일관성**: `runCollect(stdout, stderr, providers, credsPath, zaiCachePath) int` — T2 정의/T3 소비 일치. `parseProviderArg(args []string) (string, error)` T1/T3 일치. sentinel `errProviderUsage` T1 정의 후 T3에서 사용법 오류로 재사용(문자열로 stderr 출력)
4. **Review Focus**: 6개 항목 모두 지정 태스크 스텝에 테스트 배치 완료
