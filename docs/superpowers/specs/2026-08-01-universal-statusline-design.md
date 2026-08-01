# Universal StatusLine CLI Tool (`statusline`) Design Specification

**Date:** 2026-08-01  
**Target Runtimes:** Claude Code, Codex CLI, Antigravity CLI  
**Language:** Go (1.22+)  
**Config File:** `~/.config/statusline/config.json`  

---

## 1. Executive Summary

`statusline` is a lightweight, ultra-fast (<5ms target execution time) command-line utility written in Go. It acts as an input/output pipeline filter designed to run inside AI CLI environments (Claude Code, Codex CLI, Antigravity CLI) as their designated status bar provider.

It takes JSON stream payloads from `stdin` (or host environments), evaluates host capability matrices, applies user preferences from `~/.config/statusline/config.json`, and outputs beautifully styled ANSI status bar text directly to `stdout`.

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

# Initialize default configuration file (~/.config/statusline/config.json)
statusline init

# Custom configuration file path override
statusline --config=~/.config/statusline/config.json
```

### Auto-Detection Strategy (`--cli=auto` by default)
When `--cli=auto` is used (which is the default when no `--cli` flag is passed), `statusline` detects the host environment using the following fallback pipeline:
1. **Stdin JSON Schema Discriminator**: Inspects payload key structures (e.g., `omcLabel` / `sessionHealth` $\rightarrow$ `claude`, `antigravity_version` $\rightarrow$ `antigravity`).
2. **Environment Variables**: Checks for `CLAUDE_CONFIG_DIR`, `ANTIGRAVITY_APP_DIR`, `CODEX_ENV`.
3. **Parent Process (PPID) Name**: Inspects parent executable via `/proc` or OS process tree.
4. **Fail-Soft Fallback**: If ambiguous, falls back to `generic` minimal statusline mode without crashing or emitting stderr noise to the host TUI.

---

## 3. Architecture & Project Layout

```
statusline/
├── cmd/
│   └── statusline/
│       └── main.go          # Entrypoint, CLI flag parsing & init subcommand
├── internal/
│   ├── adapter/
│   │   ├── adapter.go       # EngineAdapter interface & Host Capabilities
│   │   ├── auto.go          # Discriminator & auto-detection engine
│   │   ├── claude.go        # Claude Code JSON parser
│   │   ├── codex.go         # Codex CLI parser & config generator mode
│   │   └── antigravity.go   # Antigravity CLI JSON parser
│   ├── config/
│   │   └── config.go        # Config parser, embedded defaults, init command
│   ├── model/
│   │   └── status.go        # UnifiedStatus struct & Capability flags
│   ├── vcs/
│   │   └── git.go           # Fast Git enrichment with 10ms deadline
│   └── render/
│       ├── renderer.go      # Lipgloss-based ANSI renderer
│       └── formatters.go    # Token bar, git, model formatters
├── go.mod
└── go.sum
```

---

## 4. Data Pipeline & Host Capability Matrix

### Unified Data Model (`UnifiedStatus`)
```go
type HostCapabilities struct {
    HasTokens     bool
    HasSkills     bool
    HasTools      bool
    HasThinking   bool
    HasPermission bool
}

type UnifiedStatus struct {
    EngineName    string // "claude", "codex", "antigravity", "generic"
    Model         string // e.g. "glm-5", "claude-3-7-sonnet"
    Cwd           string // Current working directory
    Hostname      string // System hostname
    
    // Git Metadata (Enriched asynchronously with 10ms timeout)
    GitRepo       string
    GitBranch     string
    GitStatus     string // e.g. "clean", "dirty (+2 -1)"

    // Token & Cost Usage (Optional per host)
    ContextTokens int
    ContextLimit  int
    PromptTime    float64

    // Skill & Tool Execution (Optional per host)
    ActiveSkills  []string
    LastSkill     string
    LastTool      string
    
    // Status (Optional per host)
    Permission    string
    ThinkingState string

    Capabilities  HostCapabilities
}
```

### Host Capability Matrix
| Feature | Claude Code | Codex CLI | Antigravity CLI | Generic Fallback |
| :--- | :---: | :---: | :---: | :---: |
| `model` | ✅ | ✅ | ✅ | ❌ |
| `cwd` / `git` | ✅ | ✅ | ✅ | ✅ |
| `tokens` | ✅ | ⚠️ (Optional) | ✅ | ❌ |
| `activeSkills` | ✅ | ❌ | ✅ | ❌ |
| `lastTool` | ✅ | ❌ | ✅ | ❌ |
| `thinking` | ✅ | ❌ | ✅ | ❌ |

- Widgets whose capability flag is `false` for the current host are gracefully omitted during rendering without leaving empty brackets or gaps.

---

## 5. Configuration Schema (`~/.config/statusline/config.json`)

```json
{
  "$schema": "1.0",
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

### Config File Lifecycle & Safety
- **No File Writes in Hot Path**: If `~/.config/statusline/config.json` does not exist, `statusline` uses **Embedded Default Config** in memory.
- **Explicit Init**: `statusline init` creates `~/.config/statusline/config.json` with safe permissions (`0600`).

---

## 6. Rendering, Performance & Fail-Soft Guardrails

1. **Lipgloss Rendering**: Uses `charmbracelet/lipgloss` for ANSI color formatting and text styling.
2. **Git Enrichment Timeout**: Git status querying is capped at **10ms deadline**. If git status takes longer (e.g. in huge monorepos), git metrics segment is safely skipped for that frame.
3. **Fail-Soft Error Policy**: If invalid JSON or a broken pipe occurs on `stdin`, `statusline` outputs a clean minimal status string (or empty line) instead of writing error traces to `stderr`, preventing TUI layout corruption in host applications.
4. **ANSI Width Truncation**: Sanitizes control characters and respects terminal `COLUMNS` width to prevent multi-line wrapping glitches.
