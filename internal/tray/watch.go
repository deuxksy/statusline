package tray

import (
	"time"

	"statusline/internal/collect"
)

const (
	// PollInterval — 캐시 폴링 주기 (fsnotify 미사용 — 의존성 최소)
	PollInterval = 5 * time.Second
	// CacheTTL — FetchedAt 기준 캐시 유효 기간 (mtime 아님)
	CacheTTL = 5 * time.Minute
	// MarkerTTL — refresh marker 백프레셔 (zai_attach 관행 대칭)
	MarkerTTL = 60 * time.Second
)

// NeedsRefresh — 갱신 조건 3종 판정: 캐시 부재/파손(ok=false) 또는 FetchedAt TTL 만료.
// 최초 실행(파일 없음)도 ok=false로 즉시 갱신 대상이 된다.
func NeedsRefresh(live collect.LiveSession, ok bool, now time.Time) bool {
	if !ok {
		return true
	}
	return live.FetchedAt.Add(CacheTTL).Before(now)
}

// TryRefresh — marker 획득에 성공한 경우에만 스폰한다. 진행 중(marker 존재)이면 false로 병합.
// 스폰 실패에도 marker를 제거하지 않는다 — 60초 백프레셔로 매 폴링 재스폰을 막는다
// (zai_attach "marker 스틸로 백프레셔" 관행 대칭). marker의 완료 제거는 collect 측이 담당.
func TryRefresh(refreshPath string, spawn func() error) bool {
	if !collect.TryMarkZaiRefreshAt(refreshPath, MarkerTTL) {
		return false
	}
	_ = spawn()
	return true
}
