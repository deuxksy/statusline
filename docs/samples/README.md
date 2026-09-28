# Statusline 데이터 샘플

이 디렉터리는 CLI 입력 fixture와 Codex·OpenAI API 응답 샘플을 보관합니다. 파일마다 실제 조회 결과인지 합성 예시인지 구분해 기록합니다.

## Files

- [`antigravity.json`](./antigravity.json) : Real STDIN payload from Google Antigravity (agy) CLI.
- [`claude.json`](./claude.json) : Sample STDIN payload from Claude Code / OMC.
- [`codex.json`](./codex.json): 2026-09-26 로컬 Codex 세션 JSONL에서 추출한 위젯용 테스트 샘플. `turn_context`의 모델·추론 수준과 `event_msg`의 `token_count`에 있는 `info`·`rate_limits`를 합쳤으며, `product`는 식별용으로 추가했습니다. Codex가 직접 제공하는 STDIN 스키마가 아닙니다. 작업 경로는 `/workspace/statusline`으로 익명화했고 대화·세션 식별자는 제외했습니다. 사용량과 reset 시각은 수집 당시의 스냅샷입니다. Codex 어댑터는 이 값을 파싱하며, 실시간 위젯 수집은 로컬 app-server API 연결이 필요합니다.
- [`openai/`](./openai/README.md): `statusline collect`로 실제 조회한 Codex·OpenAI Platform 응답의 익명화 스냅샷.
- [`openai/schema-examples/`](./openai/schema-examples/README.md): API 스키마 기반 합성 예시. 실제 응답이 아닙니다.
- [`zai/quota-limit-response.json`](./zai/quota-limit-response.json): 2026-09-26 실측 수치 기반 Z.AI `quota/limit` API 응답 재구성 샘플 (raw 응답 미보관, 소비 필드만 포함).
- [`zai/zai-cache.json`](./zai/zai-cache.json): `~/.cache/statusline/zai.json` 캐시 포맷 샘플 (위 응답의 정규화 결과).
- [`collect/`](./collect/): `statusline collect` 실측 응답의 익명화 스냅샷 (`codex.json`, `zai.json`) — 이메일·계정 ID·thread ID·로컬 경로는 제네릭 값으로 치환.

## How to Test Statusline Output

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
