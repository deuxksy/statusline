package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"statusline/internal/adapter"
	"statusline/internal/collect"
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
	for _, a := range args {
		if a == "--provider" {
			return "", errProviderUsage
		}
		if strings.HasPrefix(a, "--provider=") {
			provider = strings.TrimPrefix(a, "--provider=")
		}
	}
	return provider, nil
}

// selectProviders — provider(계정) → 수집 목록 매핑.
// chatgpt는 codex+platform 출력 키 묶음. anthropic/gemini는 미지원 거부.
// unknown은 sentinel을 wrap하지 않는다 — 호출부가 사용법 안내로 분기할 수 있도록.
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

// zaiProviderFromBaseURL — provider 식별 시임(var 주입). SSRF 가드 주체는 adapter
// allowlist이며 프로덕션 기본값은 그대로 — 테스트가 httptest 로컬 URL을 통과시키는
// 주입점일 뿐이다.
var zaiProviderFromBaseURL = adapter.ProviderFromBaseURL

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
			provider := zaiProviderFromBaseURL(zc.baseURL)
			res := collect.RunZaiCollect(ctx, &http.Client{Timeout: 5 * time.Second}, zc.baseURL, zc.authToken, provider, zaiCachePath, collect.ZaiRefreshPath(zaiCachePath))
			mu.Lock()
			defer mu.Unlock()
			if res.Zai != nil {
				out.Zai = res.Zai
			}
			for _, v := range res.Errors {
				out.Errors["zai"] = v // RunZaiCollect는 zai 키 하나만 씀
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
