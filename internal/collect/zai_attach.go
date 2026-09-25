package collect

import (
	"time"

	"statusline/internal/model"
)

// AttachZaiQuotaAt — 캐시에서 quota를 읽어 st.Quota에 주입하고, TTL 경과 시 비동기 갱신을 트리거.
// 렌더 경로 전용: HTTP 없이 파일 읽기만 한다 (fail-soft). spawn은 nil 허용(테스트).
func AttachZaiQuotaAt(st *model.UnifiedStatus, ttl time.Duration, authToken, cachePath, refreshPath string, spawn func() error) {
	if st == nil {
		return
	}
	if st.Provider != "zai" && st.Provider != "zhipu" {
		return
	}
	snap, mtime, ok := ReadZaiCacheAt(cachePath)
	if ok {
		name := snap.Provider
		if name == "" {
			name = st.Provider
		}
		st.Quota = []model.QuotaCategory{{
			Name:            name,
			FiveH:           snap.TokenRemaining,
			FiveHResetsAt:   snap.TokenResetsAt,
			Monthly:         snap.McpRemaining,
			MonthlyResetsAt: snap.McpResetsAt,
		}}
	}
	stale := !ok || time.Since(mtime) > ttl
	if stale && authToken != "" && TryMarkZaiRefreshAt(refreshPath, 60*time.Second) {
		if spawn != nil {
			_ = spawn() // 실패는 marker 스틸(60초)로 백프레셔
		}
	}
}

// AttachZaiQuota — 기본 경로·기본 spawn 래퍼 (main.go 진입점)
func AttachZaiQuota(st *model.UnifiedStatus, ttl time.Duration, authToken string) {
	cachePath, err := ZaiCachePath()
	if err != nil {
		return
	}
	AttachZaiQuotaAt(st, ttl, authToken, cachePath, ZaiRefreshPath(cachePath), func() error {
		return SpawnSelfCollect("zai")
	})
}
