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
