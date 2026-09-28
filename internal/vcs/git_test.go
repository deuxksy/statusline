package vcs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// initRepo creates a git repository fixture with the given branch name.
func initRepo(t *testing.T, dir, branch string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("-c", "init.defaultBranch="+branch, "init", ".")
	run("commit", "--allow-empty", "-m", "init")
}

// TestEnrichGitBranchWithoutSubprocessBudget guards the flicker regression:
// branch/repo must resolve via direct .git metadata reads, not a git
// subprocess, because spawn latency under CPU load regularly exceeds the
// shared deadline and made the branch segment disappear on some frames.
func TestEnrichGitBranchWithoutSubprocessBudget(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo, "main")

	st := model.NewUnifiedStatus("generic")
	// 1ms: no git subprocess can complete this fast on a loaded machine.
	vcs.EnrichGit(st, repo, 1)

	if st.GitBranch != "main" {
		t.Errorf("expected branch main with negligible timeout, got: %q", st.GitBranch)
	}
	if st.GitRepo != filepath.Base(repo) {
		t.Errorf("expected repo %q, got: %q", filepath.Base(repo), st.GitRepo)
	}
}

func TestEnrichGitFromSubdirectory(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo, "feature-x")
	sub := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, sub, 1)

	if st.GitBranch != "feature-x" {
		t.Errorf("expected branch feature-x from subdirectory, got: %q", st.GitBranch)
	}
	if st.GitRepo != filepath.Base(repo) {
		t.Errorf("expected repo %q from subdirectory, got: %q", filepath.Base(repo), st.GitRepo)
	}
}

func TestEnrichGitDetachedHead(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo, "main")

	detach := exec.Command("git", "checkout", "--detach")
	detach.Dir = repo
	if out, err := detach.CombinedOutput(); err != nil {
		t.Fatalf("git checkout --detach: %v\n%s", err, out)
	}

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, repo, 1)

	if st.GitBranch == "" || st.GitBranch == "HEAD" || len(st.GitBranch) != 7 {
		t.Errorf("expected abbreviated commit id for detached HEAD, got: %q", st.GitBranch)
	}
}

func TestEnrichGitWorktree(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo, "main")

	wt := filepath.Join(t.TempDir(), "wt")
	add := exec.Command("git", "worktree", "add", "-b", "wt-branch", wt)
	add.Dir = repo
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, wt, 1)

	if st.GitBranch != "wt-branch" {
		t.Errorf("expected branch wt-branch in worktree, got: %q", st.GitBranch)
	}
	if st.GitRepo != filepath.Base(wt) {
		t.Errorf("expected repo %q (worktree root) in worktree, got: %q", filepath.Base(wt), st.GitRepo)
	}
}

func TestEnrichGitDirtyStatus(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo, "main")
	// -uno: dirty marker only reflects tracked file changes, not untracked files
	tracked := filepath.Join(repo, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("a"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("add", "tracked.txt")
	run("commit", "-m", "add tracked")
	if err := os.WriteFile(tracked, []byte("b"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, repo, 500)

	if st.GitStatus != "*" {
		t.Errorf("expected dirty marker *, got: %q", st.GitStatus)
	}
}

func TestEnrichGitTrimsHeadContent(t *testing.T) {
	repo := t.TempDir()
	gitDir := filepath.Join(repo, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	head := "ref: refs/heads/release/1.2\n"
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(head), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	st := model.NewUnifiedStatus("generic")
	vcs.EnrichGit(st, repo, 1)

	if !strings.HasPrefix(st.GitBranch, "release/1.2") || strings.Contains(st.GitBranch, "\n") {
		t.Errorf("expected trimmed slash branch release/1.2, got: %q", st.GitBranch)
	}
}
