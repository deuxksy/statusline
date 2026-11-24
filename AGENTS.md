> Note: Read `@./.ai/RULES.md` first before making changes.

# Codex CLI Specific Guidelines
- Keep Codex data collection read-only and run tests with `go test ./...`.
- Code changes should preserve Go standard formatting (`gofmt`).
- Statusline integration test: `echo '{"model":"codex"}' | ./statusline --cli=codex`.
- Codex sample payload test: `cat docs/samples/codex.json | ./statusline --cli=codex`.
- Codex MCP 상태바 변경 시 `tools/list`와 `resources/read`를 검증하고 새 Codex 세션에서 출력을 확인한다.
