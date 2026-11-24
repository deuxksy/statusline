
# statusline Project Rules

## Build & Test Commands

- Build: `go build -o statusline ./cmd/statusline`
- Test: `go test -v ./...`
- Test single package: `go test -v ./internal/adapter`
- Quota collect (one-shot): `./statusline collect --provider=zai`
- Tray widget (dev): `mise run tray` (builds the binary first)

## Development Principles

- KISS & DRY: The render path is a pure I/O stream filter — no network calls, no daemons (5ms budget). `tray` and `collect` are separate processes, never linked into rendering; the render path only fire-and-forget spawns `collect` when the quota cache is stale (5min TTL).
- Fail-soft: Never emit raw tracebacks or error logs to stdout; output clean minimal fallback on pipe/JSON errors.
- Target execution speed: <5ms rendering budget for the statusline rendering path.
