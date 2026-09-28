package tray_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"statusline/internal/collect"
	"statusline/internal/tray"
)

func TestNeedsRefresh(t *testing.T) {
	now := time.Now()
	if !tray.NeedsRefresh(collect.LiveSession{}, false, now) {
		t.Error("absent/corrupt cache (ok=false) must need refresh")
	}
	fresh := collect.LiveSession{FetchedAt: now.Add(-4 * time.Minute)}
	if tray.NeedsRefresh(fresh, true, now) {
		t.Error("cache within TTL must not need refresh")
	}
	stale := collect.LiveSession{FetchedAt: now.Add(-6 * time.Minute)}
	if !tray.NeedsRefresh(stale, true, now) {
		t.Error("cache past TTL must need refresh")
	}
}

func TestTryRefreshSpawns(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "live_session.json.refresh")
	calls := 0
	if !tray.TryRefresh(marker, func() error { calls++; return nil }) {
		t.Fatal("expected true with no marker")
	}
	if calls != 1 {
		t.Errorf("spawn calls = %d, want 1", calls)
	}
}

func TestTryRefreshBackpressure(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "m.refresh")
	old := time.Now().Add(-2 * time.Minute)
	calls := 0
	spawn := func() error { calls++; return nil }

	// 1차: marker 없음 → 스폰 + marker 생성
	if !tray.TryRefresh(marker, spawn) || calls != 1 {
		t.Fatal("first refresh should spawn")
	}
	// 2차: 신선 marker → 백프레셔 (스폰 스톰 방지)
	if tray.TryRefresh(marker, spawn) {
		t.Error("fresh marker must block refresh")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (backpressure)", calls)
	}
	// 3차: marker 낡음 → 스틸 후 재스폰
	if err := os.Chtimes(marker, old, old); err != nil {
		t.Fatal(err)
	}
	if !tray.TryRefresh(marker, spawn) || calls != 2 {
		t.Error("stale marker (60s+) should be stolen and spawn again")
	}
}

func TestTryRefreshSpawnFailKeepsMarker(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "m.refresh")
	if !tray.TryRefresh(marker, func() error { return errors.New("boom") }) {
		t.Fatal("expected true (spawn attempted)")
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("marker must be kept after spawn failure (backpressure): %v", err)
	}
}
