package adapter

import (
	"net/url"
	"strings"
)

// ProviderFromBaseURL — ANTHROPIC_BASE_URL 도메인에서 provider 식별.
// 허용 도메인 allowlist가 곧 SSRF 가드다: 여기서 ""가 나오면 어디서도 API를 호출하지 않는다.
func ProviderFromBaseURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "api.z.ai":
		return "zai"
	case host == "bigmodel.cn" || strings.HasSuffix(host, ".bigmodel.cn"):
		return "zhipu"
	default:
		return ""
	}
}

// ProviderFromEnv — env 도메인 우선, 모델명 glm- 접두사 폴백
func ProviderFromEnv(env map[string]string, modelName string) string {
	if p := ProviderFromBaseURL(env["ANTHROPIC_BASE_URL"]); p != "" {
		return p
	}
	if strings.HasPrefix(strings.ToLower(modelName), "glm-") {
		return "zai"
	}
	return ""
}
