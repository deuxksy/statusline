package collect

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"statusline/internal/model"
)

func newZaiStatus() *model.UnifiedStatus {
	st := model.NewUnifiedStatus("claude")
	st.Provider = "zai"
	return st
}

func writeCacheWithMtime(t *testing.T, path string, snap ZaiSnapshot, mtime time.Time) {
	t.Helper()
	if err := WriteZaiCacheAt(path, snap); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestAttachZaiQuotaFreshCacheNoSpawn(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "zai.json")
	refreshPath := ZaiRefreshPath(cachePath)
	writeCacheWithMtime(t, cachePath, ZaiSnapshot{
		FetchedAt: time.Now().UTC(), Provider: "zai",
		TokenRemaining: 0.97, TokenResetsAt: time.Now().Add(4*time.Hour + 46*time.Minute).UnixMilli(),
		McpRemaining: 0.66, McpResetsAt: time.Now().Add(18*24*time.Hour + 19*time.Hour).UnixMilli(),
	}, time.Now())

	st := newZaiStatus()
	calls := 0
	AttachZaiQuotaAt(st, 5*time.Minute, "test-token", cachePath, refreshPath, func() error {
		calls++
		return nil
	})
	if calls != 0 {
		t.Errorf("TTL 내 캐시로 spawn 금지: calls = %d", calls)
	}
	if len(st.Quota) != 1 {
		t.Fatalf("Quota 주입 실패: len = %d", len(st.Quota))
	}
	cat := st.Quota[0]
	if cat.Name != "zai" || cat.FiveH != 0.97 || cat.Monthly != 0.66 {
		t.Errorf("Quota 매핑 오류: %+v", cat)
	}
	if cat.FiveHResetsAt == 0 || cat.MonthlyResetsAt == 0 {
		t.Errorf("reset 미매핑: %+v", cat)
	}
}

func TestAttachZaiQuotaStaleCacheSpawnsOnce(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "zai.json")
	refreshPath := ZaiRefreshPath(cachePath)
	past := time.Now().Add(-10 * time.Minute)
	writeCacheWithMtime(t, cachePath, ZaiSnapshot{FetchedAt: past, Provider: "zai", TokenRemaining: 0.9, McpRemaining: 0.6}, past)

	st := newZaiStatus()
	calls := 0
	spawn := func() error { calls++; return nil }

	AttachZaiQuotaAt(st, 5*time.Minute, "test-token", cachePath, refreshPath, spawn)
	if calls != 1 {
		t.Fatalf("stale 시 spawn 1회: calls = %d", calls)
	}
	if len(st.Quota) != 1 {
		t.Error("stale 캐시도 quota는 표시되어야 함 (stale-read)")
	}

	AttachZaiQuotaAt(st, 5*time.Minute, "test-token", cachePath, refreshPath, spawn)
	if calls != 1 {
		t.Errorf("marker가 두 번째 spawn을 억제해야 함: calls = %d", calls)
	}
}

func TestAttachZaiQuotaMissingCacheSpawns(t *testing.T) {
	dir := t.TempDir()
	st := newZaiStatus()
	calls := 0
	AttachZaiQuotaAt(st, 5*time.Minute, "test-token", filepath.Join(dir, "zai.json"), filepath.Join(dir, "zai.json.refresh"), func() error {
		calls++
		return nil
	})
	if calls != 1 {
		t.Errorf("캐시 부재 시 spawn 1회: calls = %d", calls)
	}
	if len(st.Quota) != 0 {
		t.Error("캐시 없으면 quota 미표시")
	}
}

func TestAttachZaiQuotaNoTokenNoSpawn(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "zai.json")
	writeCacheWithMtime(t, cachePath, ZaiSnapshot{FetchedAt: time.Now(), Provider: "zai"}, time.Now().Add(-10*time.Minute))
	st := newZaiStatus()
	calls := 0
	AttachZaiQuotaAt(st, 5*time.Minute, "", cachePath, ZaiRefreshPath(cachePath), func() error {
		calls++
		return nil
	})
	if calls != 0 {
		t.Error("토큰 부재 시 spawn 금지 (무의미한 프로세스 방지)")
	}
}

func TestAttachZaiQuotaWrongProviderNoop(t *testing.T) {
	st := model.NewUnifiedStatus("claude") // Provider ""
	calls := 0
	AttachZaiQuotaAt(st, 5*time.Minute, "test-token", "/nonexistent", "/nonexistent", func() error {
		calls++
		return nil
	})
	if calls != 0 || len(st.Quota) != 0 {
		t.Error("비-Z.AI provider는 완전 no-op여야 함")
	}
}

func TestAttachZaiQuotaNilSafe(t *testing.T) {
	AttachZaiQuotaAt(nil, 5*time.Minute, "t", "/n", "/n", nil) // panic 없음
}
