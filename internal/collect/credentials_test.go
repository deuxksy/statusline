package collect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdminKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if key, err := AdminKey(path); err != nil || key != "" {
		t.Fatalf("missing file: key=%q err=%v", key, err)
	}
	if err := os.WriteFile(path, []byte(`{"openai_admin_key":"test-admin"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if key, err := AdminKey(path); err != nil || key != "test-admin" {
		t.Fatalf("file: key=%q err=%v", key, err)
	}
	t.Setenv("OPENAI_ADMIN_KEY", "environment-admin")
	if key, err := AdminKey(path); err != nil || key != "environment-admin" {
		t.Fatalf("environment: key=%q err=%v", key, err)
	}
	t.Setenv("OPENAI_ADMIN_KEY", "")
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := AdminKey(path); err == nil {
		t.Fatal("insecure permissions accepted")
	}
}

func TestZaiAuthToken(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	path := filepath.Join(t.TempDir(), "credentials.json")
	if tok, err := ZaiAuthToken(path); err != nil || tok != "" {
		t.Fatalf("missing file: token=%q err=%v", tok, err)
	}
	if err := os.WriteFile(path, []byte(`{"zai_auth_token":"file-token"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if tok, err := ZaiAuthToken(path); err != nil || tok != "file-token" {
		t.Fatalf("file fallback: token=%q err=%v", tok, err)
	}
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "env-token")
	if tok, err := ZaiAuthToken(path); err != nil || tok != "env-token" {
		t.Fatalf("environment wins: token=%q err=%v", tok, err)
	}
	t.Setenv("ZAI_AUTH_TOKEN", "zai-override-token")
	if tok, err := ZaiAuthToken(path); err != nil || tok != "zai-override-token" {
		t.Fatalf("provider-scoped env wins: token=%q err=%v", tok, err)
	}
	t.Setenv("ZAI_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ZaiAuthToken(path); err == nil {
		t.Fatal("insecure permissions accepted")
	}
}

func TestZaiBaseURL(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")
	path := filepath.Join(t.TempDir(), "credentials.json")
	if u, err := ZaiBaseURL(path); err != nil || u != "https://api.z.ai/api/anthropic" {
		t.Fatalf("default endpoint: url=%q err=%v", u, err)
	}
	if err := os.WriteFile(path, []byte(`{"zai_base_url":"https://open.bigmodel.cn/api/anthropic"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if u, err := ZaiBaseURL(path); err != nil || u != "https://open.bigmodel.cn/api/anthropic" {
		t.Fatalf("file fallback: url=%q err=%v", u, err)
	}
	t.Setenv("ANTHROPIC_BASE_URL", "https://api.z.ai/api/anthropic")
	if u, err := ZaiBaseURL(path); err != nil || u != "https://api.z.ai/api/anthropic" {
		t.Fatalf("environment wins: url=%q err=%v", u, err)
	}
	t.Setenv("ZAI_BASE_URL", "https://open.bigmodel.cn/api/anthropic")
	if u, err := ZaiBaseURL(path); err != nil || u != "https://open.bigmodel.cn/api/anthropic" {
		t.Fatalf("provider-scoped env wins: url=%q err=%v", u, err)
	}
}
