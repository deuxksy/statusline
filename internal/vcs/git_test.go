package vcs_test

import (
	"os"
	"testing"
	"time"

	"statusline/internal/model"
	"statusline/internal/vcs"
)

func TestEnrichGitInGitRepo(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, cwd, 500) // generous timeout for testing current repo

	if st.GitRepo == "" {
		t.Errorf("expected non-empty GitRepo in current git repository")
	}
	if st.GitBranch == "" {
		t.Errorf("expected non-empty GitBranch in current git repository")
	}
}

func TestEnrichGitNonGitRepo(t *testing.T) {
	tempDir := t.TempDir()

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, tempDir, 500)

	if st.GitRepo != "" {
		t.Errorf("expected empty GitRepo for non-git directory, got: %s", st.GitRepo)
	}
	if st.GitBranch != "" {
		t.Errorf("expected empty GitBranch for non-git directory, got: %s", st.GitBranch)
	}
}

func TestEnrichGitTimeoutBound(t *testing.T) {
	st := model.NewUnifiedStatus("generic")
	start := time.Now()
	// Pass a very short timeout (1ms) to verify deadline is respected
	vcs.EnrichGit(st, ".", 1)
	duration := time.Since(start)

	if duration > 50*time.Millisecond {
		t.Errorf("git enrichment exceeded safe threshold: %v", duration)
	}
}
