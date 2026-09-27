
# statusline Project Rules

## Build & Test Commands

- Build: `go build -o statusline ./cmd/statusline`
- Test: `go test -v ./...`
- Test single package: `go test -v ./internal/adapter`

## Development Principles

- KISS & DRY: Simple input/output CLI stream filter, no persistent background daemons.
- Fail-soft: Never emit raw tracebacks or error logs to stdout; output clean minimal fallback on pipe/JSON errors.
- Target execution speed: <5ms rendering budget for the statusline rendering path.
