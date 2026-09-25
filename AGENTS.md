> Note: Read `@./.ai/RULES.md` first before making changes.

# Codex CLI Specific Guidelines
- Keep Codex data collection read-only and run tests with `go test ./...`.
- Code changes should preserve Go standard formatting (`gofmt`).
- Statusline integration test: `echo '{"model":"codex"}' | ./statusline --cli=codex`.
