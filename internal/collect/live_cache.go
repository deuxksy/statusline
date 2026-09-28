package collect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LiveSession — 트레이(v0.8.x)용 통합 수집 캐시. 성공 provider만 담는다 (errors 미기록).
// TTL은 mtime이 아닌 FetchedAt 기준으로 호출자가 판정한다.
type LiveSession struct {
	FetchedAt time.Time                  `json:"fetchedAt"`
	Codex     map[string]json.RawMessage `json:"codex,omitempty"`
	Platform  *PlatformSnapshot          `json:"platform,omitempty"`
	Zai       *ZaiSnapshot               `json:"zai,omitempty"`
}

// LiveCachePath — os.UserCacheDir 기반 통합 캐시 경로 (zai.json 대칭)
func LiveCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "statusline", "live_session.json"), nil
}

// LiveRefreshPath — 통합 캐시 옆 refresh marker (스폰 스톰 가드)
func LiveRefreshPath(cachePath string) string {
	return cachePath + ".refresh"
}

// ReadLiveCacheAt — 부재·파손·FetchedAt zero는 false로 흡수 (fail-soft)
func ReadLiveCacheAt(path string) (LiveSession, bool) {
	var live LiveSession
	data, err := os.ReadFile(path)
	if err != nil {
		return live, false
	}
	if json.Unmarshal(data, &live) != nil || live.FetchedAt.IsZero() {
		return LiveSession{}, false
	}
	return live, true
}

// WriteLiveCacheAt — atomic write: tmp 기록 후 rename. 시크릿 미포함이라 0644.
func WriteLiveCacheAt(path string, live LiveSession) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(live)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// MergeLiveSession — non-nil provider만 교체하고 나머지는 유지. FetchedAt = now.
// 호출자는 성공 provider가 1개 이상일 때만 쓴다 (전 실패 시 기존 캐시 불변).
func MergeLiveSession(prev LiveSession, codex map[string]json.RawMessage, platform *PlatformSnapshot, zai *ZaiSnapshot, now time.Time) LiveSession {
	merged := prev
	if codex != nil {
		merged.Codex = codex
	}
	if platform != nil {
		merged.Platform = platform
	}
	if zai != nil {
		merged.Zai = zai
	}
	merged.FetchedAt = now
	return merged
}
