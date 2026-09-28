package main_test

import (
	"bytes"
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
	requireExit2(t, err, stderr.String())
	if !strings.Contains(stderr.String(), "unknown provider") {
		t.Errorf("stderr에 안내 필요: %s", stderr.String())
	}
}

func TestCLICollectSpaceFormRejected(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "collect", "--provider", "zai")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	requireExit2(t, err, stderr.String())
	if !strings.Contains(stderr.String(), "expected =value") {
		t.Errorf("안내 메시지 필요: %s", stderr.String())
	}
}

// requireExit2 — go run 래퍼는 자식 exit code를 1로 평준화하고 stderr에
// "exit status N"을 남긴다 (go1.26 실측). 전파하는 버전에서는 ExitCode==2로,
// 평준화하는 버전에서는 마커 문자열로 프로그램 exit 2를 검증한다.
func requireExit2(t *testing.T, err error, stderr string) {
	t.Helper()
	ee, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if ee.ExitCode() != 2 && !strings.Contains(stderr, "exit status 2") {
		t.Fatalf("expected exit 2, got exit %d, stderr: %s", ee.ExitCode(), stderr)
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
