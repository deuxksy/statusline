package collect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ZaiCachePath — XDG ~/.cache/statusline/zai.json
func ZaiCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "statusline", "zai.json"), nil
}

// ZaiRefreshPath — 캐시 옆의 refresh marker (스폰 스톰 가드)
func ZaiRefreshPath(cachePath string) string {
	return cachePath + ".refresh"
}

// WriteZaiCacheAt — atomic write: tmp 기록 후 rename. 시크릿 미포함이라 0644.
func WriteZaiCacheAt(path string, snap ZaiSnapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ReadZaiCacheAt — 캐시 읽기. 모든 실패는 false로 흡수 (fail-soft).
// TTL 판정은 호출자가 mtime으로 수행한다.
func ReadZaiCacheAt(path string) (ZaiSnapshot, time.Time, bool) {
	var snap ZaiSnapshot
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return snap, time.Time{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return snap, time.Time{}, false
	}
	if err := json.Unmarshal(data, &snap); err != nil || snap.FetchedAt.IsZero() {
		return ZaiSnapshot{}, time.Time{}, false
	}
	return snap, fi.ModTime(), true
}

// TryMarkZaiRefreshAt — marker를 O_EXCL로 생성. maxAge보다 오래된 marker는 스틸한다.
// true 반환 시 이 호출자가 갱신 책임을 가진다.
func TryMarkZaiRefreshAt(refreshPath string, maxAge time.Duration) bool {
	if fi, err := os.Stat(refreshPath); err == nil {
		if time.Since(fi.ModTime()) < maxAge {
			return false
		}
		_ = os.Remove(refreshPath)
	}
	f, err := os.OpenFile(refreshPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// ClearZaiRefreshAt — 갱신 완료 후 marker 제거
func ClearZaiRefreshAt(refreshPath string) {
	_ = os.Remove(refreshPath)
}
