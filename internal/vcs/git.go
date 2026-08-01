package vcs

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"statusline/internal/model"
)

// EnrichGit inspects the repository at cwd and updates GitBranch, GitRepo, and GitStatus
// in status within the given timeout limit in milliseconds.
func EnrichGit(status *model.UnifiedStatus, cwd string, timeoutMs int) {
	if cwd == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// Get current branch name
	cmdBranch := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmdBranch.Dir = cwd
	outBranch, err := cmdBranch.Output()
	if err != nil {
		return
	}
	status.GitBranch = strings.TrimSpace(string(outBranch))

	// Get repository top-level directory name
	cmdRepo := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmdRepo.Dir = cwd
	outRepo, err := cmdRepo.Output()
	if err == nil {
		topLevel := strings.TrimSpace(string(outRepo))
		if topLevel != "" {
			status.GitRepo = filepath.Base(topLevel)
		}
	}

	// Get dirty status if deadline not exceeded
	if ctx.Err() == nil {
		cmdStatus := exec.CommandContext(ctx, "git", "status", "--porcelain", "-uno")
		cmdStatus.Dir = cwd
		outStatus, err := cmdStatus.Output()
		if err == nil && len(strings.TrimSpace(string(outStatus))) > 0 {
			status.GitStatus = "*"
		}
	}
}
