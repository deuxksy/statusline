package collect

import (
	"encoding/json"
	"fmt"
	"os"
)

// ZaiDefaultBaseURL — credential 폴백 사용 시 URL 미지정의 기본값 (Z.AI 공식 anthropic 호환 엔드포인트)
const ZaiDefaultBaseURL = "https://api.z.ai/api/anthropic"

type credentialsFile struct {
	OpenAIAdminKey string `json:"openai_admin_key"`
	ZaiAuthToken   string `json:"zai_auth_token,omitempty"`
	ZaiBaseURL     string `json:"zai_base_url,omitempty"`
}

// readCredentialsFile — 0600 퍼미션 검사 후 파싱. 파일이 없으면 ok=false, 나머지 실패는 error.
func readCredentialsFile(path string) (creds credentialsFile, ok bool, err error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return credentialsFile{}, false, nil
	}
	if err != nil {
		return credentialsFile{}, false, err
	}
	if info.Mode().Perm()&0077 != 0 {
		return credentialsFile{}, false, fmt.Errorf("credential file permissions must be 0600")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return credentialsFile{}, false, err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return credentialsFile{}, false, err
	}
	return creds, true, nil
}

// AdminKey reads the local credential only when no explicit environment value exists.
func AdminKey(path string) (string, error) {
	if key := os.Getenv("OPENAI_ADMIN_KEY"); key != "" {
		return key, nil
	}
	creds, ok, err := readCredentialsFile(path)
	if err != nil || !ok {
		return "", err
	}
	return creds.OpenAIAdminKey, nil
}

// ZaiAuthToken — ZAI_AUTH_TOKEN → ANTHROPIC_AUTH_TOKEN → 로컬 credential 폴백.
// ZAI_*는 provider 고유 오버라이드(공식 Claude Code 환경 등 ANTHROPIC_* 충돌 회피용).
func ZaiAuthToken(path string) (string, error) {
	if token := os.Getenv("ZAI_AUTH_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("ANTHROPIC_AUTH_TOKEN"); token != "" {
		return token, nil
	}
	creds, ok, err := readCredentialsFile(path)
	if err != nil || !ok {
		return "", err
	}
	return creds.ZaiAuthToken, nil
}

// ZaiBaseURL — ZAI_BASE_URL → ANTHROPIC_BASE_URL → zai_base_url → Z.AI 기본 엔드포인트.
func ZaiBaseURL(path string) (string, error) {
	if u := os.Getenv("ZAI_BASE_URL"); u != "" {
		return u, nil
	}
	if u := os.Getenv("ANTHROPIC_BASE_URL"); u != "" {
		return u, nil
	}
	creds, ok, err := readCredentialsFile(path)
	if err != nil {
		return "", err
	}
	if ok && creds.ZaiBaseURL != "" {
		return creds.ZaiBaseURL, nil
	}
	return ZaiDefaultBaseURL, nil
}
