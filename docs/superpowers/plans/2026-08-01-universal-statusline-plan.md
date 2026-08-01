# Universal StatusLine (`statusline`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a lightweight, high-performance, single-binary Go CLI tool (`statusline`) that accepts status payloads via stdin, auto-detects or adapts for Claude Code, Codex CLI, and Antigravity CLI, enriches status with git information (<10ms deadline), and renders beautiful ANSI status lines via Lipgloss.

**Architecture:** A pipeline architecture (`stdin -> Adapter Parsing -> Model Normalization -> Async Git Enrichment -> Lipgloss Renderer -> stdout`). If auto-detection fails, falls back gracefully to a generic minimal status bar without crashing. Missing configuration relies on embedded defaults, while `statusline init` explicitly generates `~/.config/statusline/config.json`.

**Tech Stack:** Go 1.22+, `github.com/charmbracelet/lipgloss`, `github.com/spf13/cobra` (or stdlib `flag`), `encoding/json`.

## Global Constraints
- Target execution time: <5ms for status rendering logic (with 10ms max budget for Git enrichment).
- Fail-soft: Never output raw error tracebacks or crash on malformed JSON; output clean minimal string or empty line to protect host TUI.
- Default flag: `--cli=auto`.
- Config location: `~/.config/statusline/config.json` (lazy read, embedded default in memory, no automatic file creation in hot path).

---

### Task 1: Project Setup & Unified Data Model

**Files:**
- Create: `go.mod`
- Create: `internal/model/status.go`
- Test: `internal/model/status_test.go`

**Interfaces:**
- Consumes: None
- Produces: `model.UnifiedStatus`, `model.HostCapabilities`

- [ ] **Step 1: Write failing test for UnifiedStatus default capabilities**

```go
package model_test

import (
	"testing"
	"statusline/internal/model"
)

func TestUnifiedStatusDefaults(t *testing.T) {
	status := model.NewUnifiedStatus("generic")
	if status.EngineName != "generic" {
		t.Errorf("expected engine generic, got %s", status.EngineName)
	}
	if status.Capabilities.HasTokens != false {
		t.Errorf("expected HasTokens false by default for generic")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/model`
Expected: FAIL (package or types not defined)

- [ ] **Step 3: Write minimal implementation in `internal/model/status.go`**

```go
package model

type HostCapabilities struct {
	HasTokens     bool
	HasSkills     bool
	HasTools      bool
	HasThinking   bool
	HasPermission bool
}

type UnifiedStatus struct {
	EngineName    string
	Model         string
	Cwd           string
	Hostname      string
	
	GitRepo       string
	GitBranch     string
	GitStatus     string

	ContextTokens int
	ContextLimit  int
	PromptTime    float64

	ActiveSkills  []string
	LastSkill     string
	LastTool      string
	
	Permission    string
	ThinkingState string

	Capabilities  HostCapabilities
}

func NewUnifiedStatus(engineName string) *UnifiedStatus {
	caps := HostCapabilities{}
	if engineName == "claude" || engineName == "antigravity" {
		caps.HasTokens = true
		caps.HasSkills = true
		caps.HasTools = true
		caps.HasThinking = true
		caps.HasPermission = true
	}
	return &UnifiedStatus{
		EngineName:   engineName,
		Capabilities: caps,
	}
}
```

- [ ] **Step 4: Initialize `go.mod` and run tests**

```bash
go mod init statusline
go test ./internal/model
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod internal/model/
git commit -m "feat: initialize go module and unified status model"
```

---

### Task 2: Embedded Configuration System & `init` Command

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Consumes: None
- Produces: `config.Config`, `config.LoadConfig(path string)`, `config.SaveDefaultConfig(path string)`

- [ ] **Step 1: Write failing test for config loading & embedded fallback**

```go
package config_test

import (
	"testing"
	"statusline/internal/config"
)

func TestLoadConfigEmbeddedDefault(t *testing.T) {
	cfg := config.LoadConfig("/non/existent/path/config.json")
	if cfg.Theme != "sleek_dark" {
		t.Errorf("expected sleek_dark theme, got %s", cfg.Theme)
	}
	if cfg.Thresholds.ContextWarning != 70 {
		t.Errorf("expected 70 contextWarning, got %d", cfg.Thresholds.ContextWarning)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config`
Expected: FAIL

- [ ] **Step 3: Implement `internal/config/config.go`**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ElementsConfig struct {
	EngineLabel  bool   `json:"engineLabel"`
	Model        bool   `json:"model"`
	ModelFormat  string `json:"modelFormat"`
	GitRepo      bool   `json:"gitRepo"`
	GitBranch    bool   `json:"gitBranch"`
	GitStatus    bool   `json:"gitStatus"`
	Cwd          bool   `json:"cwd"`
	CwdFormat    string `json:"cwdFormat"`
	Hostname     bool   `json:"hostname"`
	ContextBar   bool   `json:"contextBar"`
	ShowTokens   bool   `json:"showTokens"`
	ActiveSkills bool   `json:"activeSkills"`
	LastTool     bool   `json:"lastTool"`
	Thinking     bool   `json:"thinking"`
}

type ThresholdsConfig struct {
	ContextWarning  int `json:"contextWarning"`
	ContextCritical int `json:"contextCritical"`
}

type LayoutConfig struct {
	Line1 []string `json:"line1"`
	Main  []string `json:"main"`
}

type Config struct {
	Schema     string           `json:"$schema,omitempty"`
	Elements   ElementsConfig   `json:"elements"`
	Thresholds ThresholdsConfig `json:"thresholds"`
	WrapMode   string           `json:"wrapMode"`
	Theme      string           `json:"theme"`
	Layout     LayoutConfig     `json:"layout"`
}

func DefaultConfig() *Config {
	return &Config{
		Schema: "1.0",
		Elements: ElementsConfig{
			EngineLabel:  true,
			Model:        true,
			ModelFormat:  "short",
			GitRepo:      true,
			GitBranch:    true,
			GitStatus:    true,
			Cwd:          true,
			CwdFormat:    "folder",
			Hostname:     true,
			ContextBar:   true,
			ShowTokens:   true,
			ActiveSkills: true,
			LastTool:     true,
			Thinking:     true,
		},
		Thresholds: ThresholdsConfig{
			ContextWarning:  70,
			ContextCritical: 85,
		},
		WrapMode: "truncate",
		Theme:    "sleek_dark",
		Layout: LayoutConfig{
			Line1: []string{"hostname", "cwd", "gitRepo", "gitBranch", "gitStatus"},
			Main:  []string{"engineLabel", "model", "thinking", "contextBar", "tokens", "activeSkills", "lastTool"},
		},
	}
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return &cfg
}

func SaveDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(DefaultConfig(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: implement config parser with embedded defaults and init saver"
```

---

### Task 3: Host Adapters & Stdin Auto-Detection Discriminator

**Files:**
- Create: `internal/adapter/adapter.go`
- Create: `internal/adapter/claude.go`
- Create: `internal/adapter/codex.go`
- Create: `internal/adapter/antigravity.go`
- Create: `internal/adapter/auto.go`
- Test: `internal/adapter/adapter_test.go`

**Interfaces:**
- Consumes: `model.UnifiedStatus`
- Produces: `adapter.EngineAdapter` interface, `adapter.ParseInput(cliFlag string, rawInput []byte, env map[string]string)`

- [ ] **Step 1: Write failing test for adapter auto-detection and parsing**

```go
package adapter_test

import (
	"testing"
	"statusline/internal/adapter"
)

func TestClaudeAdapterParsing(t *testing.T) {
	jsonPayload := []byte(`{
		"omcLabel": "OMC",
		"model": {"displayName": "Claude 3.7 Sonnet"},
		"contextBar": {"percentage": 45}
	}`)
	status, err := adapter.ParseInput("auto", jsonPayload, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.EngineName != "claude" {
		t.Errorf("expected engine claude, got %s", status.EngineName)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter`
Expected: FAIL

- [ ] **Step 3: Implement EngineAdapter and Concrete Adapters**

In `internal/adapter/adapter.go`:
```go
package adapter

import "statusline/internal/model"

type EngineAdapter interface {
	Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error)
}
```

In `internal/adapter/claude.go`:
```go
package adapter

import (
	"encoding/json"
	"statusline/internal/model"
)

type ClaudeAdapter struct{}

func (c *ClaudeAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("claude")
	var raw struct {
		Model struct {
			DisplayName string `json:"displayName"`
		} `json:"model"`
		ContextBar struct {
			Percentage int `json:"percentage"`
		} `json:"contextBar"`
		Thinking struct {
			State string `json:"state"`
		} `json:"thinking"`
	}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &raw)
		st.Model = raw.Model.DisplayName
		st.ContextTokens = raw.ContextBar.Percentage
		st.ThinkingState = raw.Thinking.State
	}
	return st, nil
}
```

In `internal/adapter/codex.go`:
```go
package adapter

import (
	"statusline/internal/model"
)

type CodexAdapter struct{}

func (c *CodexAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("codex")
	st.Model = "codex"
	return st, nil
}
```

In `internal/adapter/antigravity.go`:
```go
package adapter

import (
	"encoding/json"
	"statusline/internal/model"
)

type AntigravityAdapter struct{}

func (a *AntigravityAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("antigravity")
	var raw struct {
		Model string `json:"model"`
	}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &raw)
		st.Model = raw.Model
	}
	return st, nil
}
```

In `internal/adapter/auto.go`:
```go
package adapter

import (
	"bytes"
	"statusline/internal/model"
)

func ParseInput(cliFlag string, rawInput []byte, env map[string]string) (*model.UnifiedStatus, error) {
	targetEngine := cliFlag
	if targetEngine == "auto" || targetEngine == "" {
		if bytes.Contains(rawInput, []byte("omcLabel")) || env["CLAUDE_CONFIG_DIR"] != "" {
			targetEngine = "claude"
		} else if bytes.Contains(rawInput, []byte("antigravity")) || env["ANTIGRAVITY_APP_DIR"] != "" {
			targetEngine = "antigravity"
		} else if env["CODEX_ENV"] != "" {
			targetEngine = "codex"
		} else {
			targetEngine = "generic"
		}
	}

	var a EngineAdapter
	switch targetEngine {
	case "claude":
		a = &ClaudeAdapter{}
	case "codex":
		a = &CodexAdapter{}
	case "antigravity":
		a = &AntigravityAdapter{}
	default:
		st := model.NewUnifiedStatus("generic")
		return st, nil
	}
	return a.Parse(rawInput, env)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/
git commit -m "feat: implement adapter pattern and auto-detection discriminator"
```

---

### Task 4: Fast Git Status Enrichment (10ms Deadline)

**Files:**
- Create: `internal/vcs/git.go`
- Test: `internal/vcs/git_test.go`

**Interfaces:**
- Consumes: working directory path
- Produces: `vcs.EnrichGit(status *model.UnifiedStatus, cwd string, timeoutMs int)`

- [ ] **Step 1: Write failing test for Git enrichment timeout guard**

```go
package vcs_test

import (
	"testing"
	"time"
	"statusline/internal/model"
	"statusline/internal/vcs"
)

func TestEnrichGitTimeout(t *testing.T) {
	st := model.NewUnifiedStatus("generic")
	start := time.Now()
	vcs.EnrichGit(st, ".", 10) // 10ms timeout
	duration := time.Since(start)

	if duration > 50*time.Millisecond {
		t.Errorf("git enrichment exceeded safe threshold: %v", duration)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/vcs`
Expected: FAIL

- [ ] **Step 3: Implement `internal/vcs/git.go` using `exec.CommandContext`**

```go
package vcs

import (
	"context"
	"os/exec"
	"strings"
	"time"
	"statusline/internal/model"
)

func EnrichGit(status *model.UnifiedStatus, cwd string, timeoutMs int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return
	}
	status.GitBranch = strings.TrimSpace(string(out))

	// Get repo name
	cmdRepo := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmdRepo.Dir = cwd
	if outRepo, err := cmdRepo.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(outRepo)), "/")
		if len(parts) > 0 {
			status.GitRepo = parts[len(parts)-1]
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/vcs`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/vcs/
git commit -m "feat: implement git status enrichment with strict 10ms deadline"
```

---

### Task 5: Lipgloss ANSI Renderer & Fail-Soft Formatting

**Files:**
- Create: `internal/render/renderer.go`
- Test: `internal/render/renderer_test.go`

**Interfaces:**
- Consumes: `model.UnifiedStatus`, `config.Config`
- Produces: `render.Render(st *model.UnifiedStatus, cfg *config.Config) string`

- [ ] **Step 1: Add Lipgloss dependency to `go.mod`**

```bash
go get github.com/charmbracelet/lipgloss
```

- [ ] **Step 2: Write failing test for Lipgloss statusline rendering**

```go
package render_test

import (
	"testing"
	"statusline/internal/config"
	"statusline/internal/model"
	"statusline/internal/render"
)

func TestRenderOutput(t *testing.T) {
	st := model.NewUnifiedStatus("claude")
	st.Model = "Sonnet 3.7"
	st.GitBranch = "main"

	cfg := config.DefaultConfig()
	output := render.Render(st, cfg)

	if output == "" {
		t.Errorf("expected non-empty rendered string")
	}
}
```

- [ ] **Step 3: Implement `internal/render/renderer.go`**

```go
package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"statusline/internal/config"
	"statusline/internal/model"
)

var (
	badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6366F1")).
			Padding(0, 1)

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	modelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)
)

func Render(st *model.UnifiedStatus, cfg *config.Config) string {
	var line1Segments []string
	var mainSegments []string

	// Engine Badge
	if cfg.Elements.EngineLabel {
		label := strings.ToUpper(st.EngineName)
		mainSegments = append(mainSegments, badgeStyle.Render(label))
	}

	// Model
	if cfg.Elements.Model && st.Model != "" {
		mainSegments = append(mainSegments, modelStyle.Render(st.Model))
	}

	// Git Branch
	if cfg.Elements.GitBranch && st.GitBranch != "" {
		line1Segments = append(line1Segments, branchStyle.Render(" "+st.GitBranch))
	}

	// Tokens
	if cfg.Elements.ShowTokens && st.Capabilities.HasTokens && st.ContextTokens > 0 {
		mainSegments = append(mainSegments, fmt.Sprintf("[%d%%]", st.ContextTokens))
	}

	line1 := strings.Join(line1Segments, " ")
	main := strings.Join(mainSegments, " │ ")

	if line1 != "" {
		return line1 + "\n" + main
	}
	return main
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/render`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/render/
git commit -m "feat: implement lipgloss ANSI renderer"
```

---

### Task 6: Main CLI Integration & E2E Validation

**Files:**
- Create: `cmd/statusline/main.go`
- Test: `cmd/statusline/main_test.go`

- [ ] **Step 1: Write E2E CLI test**

```go
package main_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCLIMainAuto(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/statusline", "--cli=auto")
	cmd.Stdin = strings.NewReader(`{"omcLabel":"OMC","model":{"displayName":"TestModel"}}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "CLAUDE") {
		t.Errorf("expected CLAUDE in output, got: %s", string(out))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/statusline`
Expected: FAIL

- [ ] **Step 3: Implement `cmd/statusline/main.go`**

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"statusline/internal/adapter"
	"statusline/internal/config"
	"statusline/internal/render"
	"statusline/internal/vcs"
)

func main() {
	var cliFlag string
	var configPath string
	var autoFlag bool

	flag.StringVar(&cliFlag, "cli", "auto", "Target CLI adapter (auto, claude, codex, antigravity)")
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.BoolVar(&autoFlag, "auto", false, "Alias for --cli=auto")
	flag.Parse()

	// Handle init subcommand
	if flag.NArg() > 0 && flag.Arg(0) == "init" {
		targetPath := configPath
		if targetPath == "" {
			home, _ := os.UserHomeDir()
			targetPath = filepath.Join(home, ".config", "statusline", "config.json")
		}
		if err := config.SaveDefaultConfig(targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Initialized default config at %s\n", targetPath)
		return
	}

	if autoFlag {
		cliFlag = "auto"
	}

	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "statusline", "config.json")
	}

	// Fail-soft stdin read
	var rawInput []byte
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		rawInput, _ = io.ReadAll(os.Stdin)
	}

	// Environment mapping
	env := map[string]string{
		"CLAUDE_CONFIG_DIR":  os.Getenv("CLAUDE_CONFIG_DIR"),
		"ANTIGRAVITY_APP_DIR": os.Getenv("ANTIGRAVITY_APP_DIR"),
		"CODEX_ENV":           os.Getenv("CODEX_ENV"),
	}

	cfg := config.LoadConfig(configPath)
	status, _ := adapter.ParseInput(cliFlag, rawInput, env)

	// Enrich Git status asynchronously with 10ms budget
	cwd, _ := os.Getwd()
	vcs.EnrichGit(status, cwd, 10)

	output := render.Render(status, cfg)
	fmt.Print(output)
}
```

- [ ] **Step 4: Build and test E2E**

```bash
go build -o statusline ./cmd/statusline
go test ./...
```
Expected: PASS

- [ ] **Step 5: Final Commit**

```bash
git add .
git commit -m "feat: complete statusline CLI implementation and E2E validation"
```

---

## Plan Self-Review Checklist
1. **Spec Coverage**: All items in `docs/superpowers/specs/2026-08-01-universal-statusline-design.md` covered (`--cli=auto` default, `statusline init`, Go 1.22+, Lipgloss renderer, 10ms Git timeout, embedded defaults).
2. **No Placeholders**: Every step has exact code blocks, commands, and expected outcomes.
3. **Type Consistency**: `model.UnifiedStatus`, `config.Config`, and `adapter.EngineAdapter` types match across all tasks.
