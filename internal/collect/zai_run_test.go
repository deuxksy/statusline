package collect

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunZaiCollectWritesCache(t *testing.T) {
	srv := newZaiServer(t, http.StatusOK, zaiBodyBoth)
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "zai.json")
	refreshPath := ZaiRefreshPath(cachePath)
	os.WriteFile(refreshPath, nil, 0o644) // 갱신 중 상태 가정

	res := RunZaiCollect(context.Background(), srv.Client(), srv.URL, "test-token", "zai", cachePath, refreshPath)
	if res.Errors != nil {
		t.Fatalf("예기치 못한 오류: %v", res.Errors)
	}
	if res.Zai == nil || res.Zai.Provider != "zai" {
		t.Fatal("성공 시 Zai 스냅샷이 채워져야 함")
	}
	if _, _, ok := ReadZaiCacheAt(cachePath); !ok {
		t.Error("캐시가 기록되지 않음")
	}
	if _, err := os.Stat(refreshPath); !os.IsNotExist(err) {
		t.Error("성공 시 refresh marker가 제거되어야 함")
	}
}

func TestRunZaiCollectFailureKeepsCache(t *testing.T) {
	srv := newZaiServer(t, http.StatusInternalServerError, `{}`)
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "zai.json")
	refreshPath := ZaiRefreshPath(cachePath)

	prev := ZaiSnapshot{FetchedAt: time.Now().UTC(), Provider: "zai", TokenRemaining: 0.5, McpRemaining: 0.5}
	if err := WriteZaiCacheAt(cachePath, prev); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(refreshPath, nil, 0o644)

	res := RunZaiCollect(context.Background(), srv.Client(), srv.URL, "test-token", "zai", cachePath, refreshPath)
	if res.Errors == nil {
		t.Fatal("실패 시 Errors가 설정되어야 함")
	}
	if res.Zai != nil {
		t.Error("실패 시 Zai는 nil이어야 함")
	}
	if _, _, ok := ReadZaiCacheAt(cachePath); !ok {
		t.Error("기존 캐시가 유지되어야 함")
	}
	if _, err := os.Stat(refreshPath); err != nil {
		t.Error("실패 시 marker가 유지되어야 함 (재시도 백프레셔)")
	}
}

func TestRunZaiCollectUnrecognized(t *testing.T) {
	res := RunZaiCollect(context.Background(), nil, "https://api.anthropic.com", "test-token", "", filepath.Join(t.TempDir(), "zai.json"), "/dev/null-marker")
	if res.Errors == nil {
		t.Fatal("allowlist 외 URL은 오류여야 함 (SSRF 가드)")
	}
	if res.Errors["zai"] == "" {
		t.Error("errors[zai]에 메시지가 있어야 함")
	}
}
