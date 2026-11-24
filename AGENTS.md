> Note: Read `@./.ai/RULES.md` first before making changes.

# Codex CLI Specific Guidelines
- Keep Codex data collection read-only and run tests with `go test ./...`.
- Code changes should preserve Go standard formatting (`gofmt`).
- Statusline integration test: `echo '{"model":"codex"}' | ./statusline --cli=codex`.
- Codex sample payload test: `cat docs/samples/codex.json | ./statusline --cli=codex`.
- Codex MCP 상태바 변경 시 `tools/list`와 `resources/read`를 검증하고 새 Codex 세션에서 출력을 확인한다.
- MCP smoke test: `printf '%s\n' '{"id":1,"method":"tools/list"}' '{"id":2,"method":"resources/read","params":{"uri":"ui://statusline-tickets/footer"}}' | python3 -u integrations/codex/statusline.py | grep -c CODEX` — `1`이 출력되어야 한다. `0`이면 저장소 루트 `./statusline` 빌드 누락이므로 `go build -o statusline ./cmd/statusline`부터 실행한다.
