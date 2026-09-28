package collect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteReadLiveCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "live_session.json")
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	in := LiveSession{
		FetchedAt: now,
		Codex:     map[string]json.RawMessage{"thread": json.RawMessage(`{"a":1}`)},
		Zai:       &ZaiSnapshot{Provider: "zai", TokenRemaining: 0.87},
	}
	if err := WriteLiveCacheAt(path, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, ok := ReadLiveCacheAt(path)
	if !ok {
		t.Fatal("read: ok=false, want true")
	}
	if !got.FetchedAt.Equal(now) {
		t.Errorf("FetchedAt = %v, want %v", got.FetchedAt, now)
	}
	if got.Codex["thread"] == nil || string(got.Codex["thread"]) != `{"a":1}` {
		t.Errorf("Codex not preserved: %v", got.Codex)
	}
	if got.Zai == nil || got.Zai.TokenRemaining != 0.87 {
		t.Errorf("Zai not preserved: %+v", got.Zai)
	}
}

func TestWriteLiveCacheAtCreatesDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "deep", "live_session.json")
	if err := WriteLiveCacheAt(path, LiveSession{FetchedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("write should create parent dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file missing after write: %v", err)
	}
}

func TestReadLiveCacheAtCorrupt(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name    string
		content string
		absent  bool
	}{
		{"garbage", `{`, false},
		{"empty", "", false},
		{"zero fetchedAt", `{"codex":{}}`, false},
		{"absent", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name+".json")
			if !tc.absent {
				if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if _, ok := ReadLiveCacheAt(path); ok {
				t.Errorf("ok=true, want false")
			}
		})
	}
}

func TestMergeLiveSessionPreservesOthers(t *testing.T) {
	prev := LiveSession{
		FetchedAt: time.Now().Add(-time.Minute),
		Codex:     map[string]json.RawMessage{"t": json.RawMessage(`{}`)},
		Zai:       &ZaiSnapshot{Provider: "zai", TokenRemaining: 0.5},
	}
	now := time.Now().UTC()
	merged := MergeLiveSession(prev, nil, nil, &ZaiSnapshot{Provider: "zai", TokenRemaining: 0.9}, now)
	if merged.Codex == nil {
		t.Error("Codex must be preserved when not collected")
	}
	if merged.Zai == nil || merged.Zai.TokenRemaining != 0.9 {
		t.Errorf("Zai must be replaced: %+v", merged.Zai)
	}
	if !merged.FetchedAt.Equal(now) {
		t.Errorf("FetchedAt = %v, want %v", merged.FetchedAt, now)
	}
}

func TestMergeLiveSessionNilKeeps(t *testing.T) {
	prev := LiveSession{Zai: &ZaiSnapshot{Provider: "zai", TokenRemaining: 0.5}}
	merged := MergeLiveSession(prev, nil, nil, nil, time.Now().UTC())
	if merged.Zai == nil || merged.Zai.TokenRemaining != 0.5 {
		t.Errorf("nil zai must keep prev: %+v", merged.Zai)
	}
}
