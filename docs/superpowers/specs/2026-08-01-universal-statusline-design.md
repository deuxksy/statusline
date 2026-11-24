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

### CLI별 상태바 등록 방식

| CLI | 등록 위치 | 호출 방식 |
| --- | --- | --- |
| Claude Code | `~/.claude/settings.json`의 `statusLine.command` | `statusline --cli=claude`를 실행하고 STDIN으로 JSON을 전달합니다. stdout을 상태바에 표시합니다. |
| Antigravity CLI (agy) | `~/.gemini/antigravity-cli/settings.json`의 `statusline.command` | `statusline --cli=antigravity`를 실행하고 agy 형식의 JSON을 STDIN으로 전달합니다. stdout을 상태바에 표시합니다. |
| Codex CLI 0.171.0 | `config.toml`의 `tui.status_line_provider` | MCP 서버와 도구 이름을 지정합니다. Codex가 MCP resource를 읽어 반환된 UI 트리를 상태바에 표시합니다. |

Codex의 MCP provider는 상태바를 교체할 때 사용하는 설정입니다. `statusline --cli=codex`는 STDIN JSON으로 Go 렌더러를 테스트하는 데 사용할 수 있습니다. [Codex 0.171.0 설정 스키마](https://raw.githubusercontent.com/openai/codex/rust-v0.171.0/codex-rs/core/config.schema.json#L5845)

### Codex 초기화 티켓 표시

Codex 기본 상태바에 `/usage`에서 확인하는 **초기화 티켓 잔여 수와 보유 티켓 중 가장 이른 만료일**을 선택 항목으로 추가합니다. Claude Code의 Z.ai 표시 형식을 따라 사용량 바로 뒤에 `🎟️x2(12/22)`를 표시합니다. 사용량 데이터가 없으면 티켓만 표시합니다. 잔여 수가 양수이고 만료일을 확인할 수 없으면 `🎟️x2`로 표시합니다. [Z.ai 렌더러](../../../internal/render/renderer.go#L448) · [Codex 초기화 티켓 안내](https://help.openai.com/en/articles/20001498-how-resets-work-in-codex)

Codex의 공식 커스텀 MCP provider를 사용합니다. MCP 연결부는 세션 스냅샷을 기존 Go 렌더러에 전달하고 전경색, 배지 배경색, 굵기와 고정 폭 여백을 Codex UI 트리로 변환합니다. 레이아웃은 `statusline` 설정을 따릅니다. 티켓 정보는 로컬 app-server의 읽기 전용 `account/rateLimits/read`를 MCP 프로세스의 첫 상태바 요청에서 백그라운드로 한 번 조회합니다. 조회한 수량은 프로세스가 다시 시작될 때 갱신됩니다. [공식 create-statusline 가이드](https://github.com/openai/codex/blob/rust-v0.171.0/codex-rs/skills/src/assets/samples/create-statusline/SKILL.md)

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
