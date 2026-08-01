# Reference: Statusline Architecture

This document describes the software architecture and package layout of `statusline` v0.2.

---

## Package Directory Structure

```
cmd/statusline/       Entrypoint (CLI flags, stdin reader, exit codes)
internal/
├── adapter/          CLI host adapters (Antigravity, Claude, Codex, Auto Discriminator)
├── config/           Config loader and JSON defaults
├── model/            UnifiedStatus & HostCapabilities domain models
├── render/           Lipgloss ANSI status line formatting
└── vcs/              Asynchronous Git VCS enrichment (10ms deadline)
```

## Data Pipeline Flow

```
STDIN JSON
   │
   ▼
adapter.ParseInput()
   ├── Auto Discriminator (payload key / env var check)
   └── Adapter.Parse() ──► UnifiedStatus struct
   │
   ▼
vcs.EnrichGit() (Async git status with 10ms timeout)
   │
   ▼
render.Render() ──► ANSI formatted single-line string
   │
   ▼
STDOUT
```
