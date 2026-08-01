# statusline Project Rules

## Build & Test Commands
- Build: `go build -o statusline ./cmd/statusline`
- Test: `go test -v ./...`
- Test single package: `go test -v ./internal/adapter`

## Codebase Architecture
- Entrypoint: `cmd/statusline/main.go`
- Domain Models: `internal/model/` (UnifiedStatus, HostCapabilities)
- Config System: `internal/config/` (Embedded defaults, `~/.config/statusline/config.json`)
- Host Adapters: `internal/adapter/` (Claude, Codex, Antigravity, Auto Discriminator)
- VCS Enrichment: `internal/vcs/` (Git 10ms deadline context timeout)
- Renderer: `internal/render/` (Lipgloss ANSI status line formatting)

## Development Principles
- KISS & DRY: Simple input/output CLI stream filter, no persistent background daemons.
- Fail-soft: Never emit raw tracebacks or error logs to stdout; output clean minimal fallback on pipe/JSON errors.
- Target execution speed: <5ms rendering budget.
