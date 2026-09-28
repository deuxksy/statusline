package collect

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestZaiCacheRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "statusline", "zai.json")
	snap := ZaiSnapshot{
		FetchedAt:      time.Now().UTC().Truncate(time.Second),
		Provider:       "zai",
		TokenRemaining: 0.99,
		TokenResetsAt:  1761368400000,
		McpRemaining:   0.66,
		McpResetsAt:    1764644400000,
	}
	if err := WriteZaiCacheAt(path, snap); err != nil {
		t.Fatal(err)
	}
	got, mtime, ok := ReadZaiCacheAt(path)
	if !ok {
		t.Fatal("읽기 실패 — ok = false")
	}
	if got.Provider != "zai" || got.TokenRemaining != snap.TokenRemaining || got.McpRemaining != snap.McpRemaining {
		t.Errorf("roundtrip 불일치: %+v", got)
	}
	if !got.FetchedAt.Equal(snap.FetchedAt) {
		t.Errorf("FetchedAt = %v, want %v", got.FetchedAt, snap.FetchedAt)
	}
	if time.Since(mtime) > 5*time.Second {
		t.Errorf("mtime이 너무 오래됨: %v", mtime)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp 파일이 남아있음 — atomic write 위반")
	}
}

func TestZaiCacheMissing(t *testing.T) {
	if _, _, ok := ReadZaiCacheAt(filepath.Join(t.TempDir(), "nope.json")); ok {
		t.Error("없는 파일은 ok = false여야 함")
	}
}

func TestZaiCacheCorrupt(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "zai.json")

	os.WriteFile(p, []byte("not json"), 0o644)
	if _, _, ok := ReadZaiCacheAt(p); ok {
		t.Error("파손 JSON은 ok = false여야 함")
	}

	os.WriteFile(p, []byte(`{"provider":"zai"}`), 0o644) // fetched_at 누락
	if _, _, ok := ReadZaiCacheAt(p); ok {
		t.Error("fetched_at 누락 스냅샷은 ok = false여야 함")
	}
}

func TestZaiRefreshMarkerLifecycle(t *testing.T) {
	p := filepath.Join(t.TempDir(), "zai.json.refresh")
	if !TryMarkZaiRefreshAt(p, 60*time.Second) {
		t.Fatal("첫 mark는 성공해야 함")
	}
	if TryMarkZaiRefreshAt(p, 60*time.Second) {
		t.Error("maxAge 내 두 번째 mark는 실패해야 함 (스폰 스톰 가드)")
	}
	ClearZaiRefreshAt(p)
	if !TryMarkZaiRefreshAt(p, 60*time.Second) {
		t.Error("clear 후 재mark는 성공해야 함")
	}
}

func TestZaiRefreshMarkerStaleSteal(t *testing.T) {
	p := filepath.Join(t.TempDir(), "zai.json.refresh")
	past := time.Now().Add(-5 * time.Minute)
	os.WriteFile(p, nil, 0o644)
	os.Chtimes(p, past, past)
	if !TryMarkZaiRefreshAt(p, 60*time.Second) {
		t.Error("maxAge 경과 marker는 스틸해야 함 (죽은 collect 복구)")
	}
}

func TestZaiCachePathShape(t *testing.T) {
	p, err := ZaiCachePath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "zai.json" {
		t.Errorf("경로 마지막 요소 = %q, want zai.json", filepath.Base(p))
	}
	if ZaiRefreshPath(p) != p+".refresh" {
		t.Errorf("ZaiRefreshPath = %q", ZaiRefreshPath(p))
	}
}
