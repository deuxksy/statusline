package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIMainAuto(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--cli=auto")
	cmd.Stdin = strings.NewReader(`{"omcLabel":"OMC","model":{"displayName":"Claude 3.7 Sonnet"}}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "CLAUDE") {
		t.Errorf("expected CLAUDE in output, got: %s", string(out))
	}
}

func TestCLIMainClaude(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--cli=claude")
	cmd.Stdin = strings.NewReader(`{"model":{"displayName":"Claude 3.7 Sonnet"}}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "CLAUDE") {
		t.Errorf("expected CLAUDE in output, got: %s", string(out))
	}
}

func TestCLIMainAntigravity(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--cli=antigravity")
	cmd.Stdin = strings.NewReader(`{"antigravity":"v1","model":"gemini-2.5-pro"}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "ANTIGRAVITY") {
		t.Errorf("expected ANTIGRAVITY in output, got: %s", string(out))
	}
}

func TestCLIMainInit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cmd := exec.Command("go", "run", ".", "--config", configPath, "init")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v, output: %s", err, string(out))
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}

	if !strings.Contains(string(data), "elements") {
		t.Errorf("expected config file to contain 'elements', got: %s", string(data))
	}
}
