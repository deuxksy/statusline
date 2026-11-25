> Note: Read `.ai/RULES.md` from the repository root before making changes.

# Codex CLI Specific Guidelines
- Keep Codex data collection read-only.
- Statusline integration test: `echo '{"model":"codex"}' | ./statusline --cli=codex`.
- Codex sample payload test: `cat docs/samples/codex.json | ./statusline --cli=codex`.
- Codex MCP 상태바 변경 시 `tools/list`와 `resources/read`를 검증하고 새 Codex 세션에서 출력을 확인한다.
- MCP smoke test: `mise run mcp-smoke` — task가 `./statusline`을 먼저 빌드한 뒤 검사하고 `1`을 출력해야 한다.
