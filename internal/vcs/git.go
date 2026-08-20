package vcs

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"statusline/internal/model"
)

// EnrichGit inspects the repository at cwd and updates GitBranch, GitRepo, and
// GitStatus in status within the given timeout limit in milliseconds.
//
// Branch and repository name are resolved by reading .git metadata directly:
// git subprocess spawns stall well past the deadline under CPU load (agent
// builds/tests saturating cores), which made the branch segment flicker
// between statusline refreshes. Only the dirty check still shells out to git,
// now with the full deadline available to it.
func EnrichGit(status *model.UnifiedStatus, cwd string, timeoutMs int) {
	if cwd == "" {
		return
	}

	branch, repo, ok := readGitInfo(cwd)
	if !ok {
		return
	}
	status.GitBranch = branch
	status.GitRepo = repo

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	cmdStatus := exec.CommandContext(ctx, "git", "status", "--porcelain", "-uno")
	cmdStatus.Dir = cwd
	outStatus, err := cmdStatus.Output()
	if err == nil && len(strings.TrimSpace(string(outStatus))) > 0 {
		status.GitStatus = "*"
	}
}

// readGitInfo walks up from cwd to the repository root and resolves the
// current branch and repository name from .git metadata (plain file reads).
func readGitInfo(cwd string) (branch, repo string, ok bool) {
	dir := cwd
	for i := 0; i < 32; i++ {
		entry := filepath.Join(dir, ".git")
		info, err := os.Stat(entry)
		if err == nil {
			if info.IsDir() {
				b, okB := branchFromHead(filepath.Join(entry, "HEAD"))
				return b, filepath.Base(dir), okB
			}
			// worktree / submodule: .git is a "gitdir: <path>" pointer file
			gitDir, err := gitDirFromPointer(entry, dir)
			if err != nil {
				return "", "", false
			}
			b, okB := branchFromHead(filepath.Join(gitDir, "HEAD"))
			return b, filepath.Base(dir), okB
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", false
}

// branchFromHead parses a HEAD file: "ref: refs/heads/<branch>" yields the
// branch name (nested paths kept), anything else is a detached object id
// abbreviated like git's default (7 chars).
func branchFromHead(headPath string) (string, bool) {
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", false
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", false
	}
	if ref, found := strings.CutPrefix(content, "ref: "); found {
		if b, ok := strings.CutPrefix(ref, "refs/heads/"); ok {
			return b, true
		}
		return filepath.Base(ref), true
	}
	if len(content) > 7 {
		return content[:7], true
	}
	return content, true
}

func gitDirFromPointer(pointerPath, baseDir string) (string, error) {
	data, err := os.ReadFile(pointerPath)
	if err != nil {
		return "", err
	}
	gitDir := strings.TrimSpace(strings.TrimPrefix(string(data), "gitdir:"))
	if gitDir == "" {
		return "", os.ErrInvalid
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(baseDir, gitDir)
	}
	return filepath.Clean(gitDir), nil
}
