# Statusline 데이터 샘플

이 디렉터리는 CLI 입력 fixture와 Codex·OpenAI API 응답 샘플을 보관합니다. 파일마다 실제 조회 결과인지 합성 예시인지 구분해 기록합니다.

## Files

- [`antigravity.json`](./antigravity.json) : Real STDIN payload from Google Antigravity (agy) CLI. Email, session IDs and local paths are replaced with generic values.
- [`claude.json`](./claude.json) : Sample STDIN payload from Claude Code / OMC.
- [`codex.json`](./codex.json): 2026-09-26 로컬 Codex 세션 JSONL에서 추출한 렌더러 fixture. `product`는 식별용으로 추가했습니다. 작업 경로는 `/workspace/statusline`으로 익명화했고 대화·세션 식별자는 제외했습니다. 사용량과 reset 시각은 수집 당시의 스냅샷입니다. 구조와 출처는 [CLI 페이로드 명세](../okf/reference/cli-payload-spec.md#3-codex-widget-snapshot)를 참조하세요.
- [`openai/`](./openai/README.md): `statusline collect`로 실제 조회한 Codex·OpenAI Platform 응답의 익명화 스냅샷.
- [`openai/schema-examples/`](./openai/schema-examples/README.md): API 스키마 기반 합성 예시. 실제 응답이 아닙니다.
- [`zai/quota-limit-response.json`](./zai/quota-limit-response.json): 2026-09-26 실측 수치 기반 Z.AI `quota/limit` API 응답 재구성 샘플 (raw 응답 미보관, 소비 필드만 포함).
- [`zai/zai-cache.json`](./zai/zai-cache.json): `~/.cache/statusline/zai.json` 캐시 포맷 샘플 (위 응답의 정규화 결과).
- [`collect/`](./collect/): `statusline collect` 실측 응답의 익명화 스냅샷 (`codex.json`, `zai.json`) — 이메일·계정 ID·thread ID·로컬 경로는 제네릭 값으로 치환.

## How to Test Statusline Output

Windows PowerShell 명령은 [Windows 11 샘플 검증](../okf/tutorials/getting-started.md#windows-11-powershell)을 참고하세요.

```bash
# Test Antigravity
cat docs/samples/antigravity.json | ./statusline --cli=antigravity

# Test Claude
cat docs/samples/claude.json | ./statusline --cli=claude

# Test Codex
cat docs/samples/codex.json | ./statusline --cli=codex

# Test Auto Discriminator
cat docs/samples/antigravity.json | ./statusline --cli=auto

# Test Z.AI quota (캐시를 실제 경로에 두고 — Linux: ~/.cache, macOS: ~/Library/Caches)
mkdir -p ~/.cache/statusline && cp docs/samples/zai/zai-cache.json ~/.cache/statusline/zai.json        # Linux
mkdir -p ~/Library/Caches/statusline && cp docs/samples/zai/zai-cache.json ~/Library/Caches/statusline/zai.json  # macOS
echo '{"omcLabel":"OMC","model":{"name":"glm-5.3"}}' | ./statusline --cli=claude
```
