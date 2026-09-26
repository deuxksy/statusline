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
