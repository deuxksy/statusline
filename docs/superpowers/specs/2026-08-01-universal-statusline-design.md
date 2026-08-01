# Universal StatusLine CLI Tool (`statusline`) Design Specification

**Date:** 2026-08-01  
**Target Runtimes:** Claude Code, Codex CLI, Antigravity CLI  
**Language:** Go (1.22+)  
**Config File:** `~/.config/statusline/config.json`  

---

## 1. Executive Summary

`statusline` is a lightweight, ultra-fast (<5ms execution time) command-line utility written in Go. It acts as an input/output pipeline filter designed to run inside various AI CLI environments (Claude Code, Codex CLI, Antigravity CLI) as their designated `statusLine` command provider.

It takes JSON data from `stdin`, environmental context, and configuration rules from `~/.config/statusline/config.json`, then outputs beautifully styled ANSI status bar text directly to `stdout`.

---

## 2. CLI Interface & Execution Modes

```bash
# Auto-detect AI CLI environment (Default: --cli=auto)
statusline
statusline --cli=auto
statusline --auto

# Explicitly specify engine adapter
statusline --cli=claude
statusline --cli=codex
statusline --cli=antigravity

# Custom configuration file path override
statusline --config=~/.config/statusline/config.json
```

### Auto-Detection Strategy (`--cli=auto` by default)
When `--cli=auto` is used (which is the **default value** if no `--cli` flag is passed), `statusline` detects the host environment using the following fallback order:
1. **Environment Variables**: Checks for `CLAUDE_CONFIG_DIR`, `CODEX_ENV`, `ANTIGRAVITY_APP_DIR`, etc.
2. **Stdin JSON Schema**: Inspects unique payload keys (e.g., `omcLabel`, `sessionHealth`).
3. **Parent Process Name**: Inspects parent process executable name via OS process tree.
4. **Fallback**: Defaults to `claude` adapter.

---

## 3. Architecture & Project Layout

```
statusline/
├── cmd/
│   └── statusline/
│       └── main.go          # Entrypoint & CLI flag parsing
├── internal/
│   ├── adapter/
│   │   ├── adapter.go       # EngineAdapter interface
│   │   ├── auto.go          # Auto-detection logic
│   │   ├── claude.go        # Claude Code parser
│   │   ├── codex.go         # Codex CLI parser
│   │   └── antigravity.go   # Antigravity CLI parser
│   ├── config/
│   │   └── config.go        # Config parser & default generator
│   ├── model/
│   │   └── status.go        # UnifiedStatus struct definition
│   └── render/
│       ├── renderer.go      # Lipgloss-based ANSI renderer
│       └── formatters.go    # Token bar, git, model formatters
├── go.mod
└── go.sum
```

---

## 4. Data Pipeline & Adapter Pattern

### Unified Data Model (`UnifiedStatus`)
```go
type UnifiedStatus struct {
    EngineName    string // "claude", "codex", "antigravity"
    Model         string // e.g. "glm-5", "claude-3-7-sonnet"
    Cwd           string // Current working directory
    Hostname      string // System hostname
    
    // Git Metadata
    GitRepo       string
    GitBranch     string
    GitStatus     string // e.g. "clean", "dirty (+2 -1)"

    // Token & Cost Usage
    ContextTokens int
    ContextLimit  int
    PromptTime    float64

    // Skill & Tool execution
    ActiveSkills  []string
    LastSkill     string
    LastTool      string
    
    // Status
    Permission    string
    ThinkingState string
}
```

### EngineAdapter Interface
```go
type EngineAdapter interface {
    Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error)
}
```

---

## 5. Configuration Schema (`~/.config/statusline/config.json`)

```json
{
  "elements": {
    "engineLabel": true,
    "model": true,
    "modelFormat": "short",
    "gitRepo": true,
    "gitBranch": true,
    "gitStatus": true,
    "cwd": true,
    "cwdFormat": "folder",
    "hostname": true,
    "contextBar": true,
    "showTokens": true,
    "activeSkills": true,
    "lastTool": true,
    "thinking": true
  },
  "thresholds": {
    "contextWarning": 70,
    "contextCritical": 85
  },
  "wrapMode": "truncate",
  "theme": "sleek_dark",
  "layout": {
    "line1": ["hostname", "cwd", "gitRepo", "gitBranch", "gitStatus"],
    "main": ["engineLabel", "model", "thinking", "contextBar", "tokens", "activeSkills", "lastTool"]
  }
}
```

- If `~/.config/statusline/config.json` does not exist, `statusline` initializes it automatically with sensible defaults.

---

## 6. Rendering & Terminal Formatting

- **Lipgloss Library**: Uses `charmbracelet/lipgloss` for styling badges, text truncation, and color palette management.
- **Dynamic Color Bar**: Token usage bar shifts dynamically from green (<70%) to yellow (70-85%) and red (>85%).
- **ANSI Width Aware Truncation**: Respects terminal `COLUMNS` width to prevent wrapping artifacts in host TUI views.
