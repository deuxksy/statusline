# Reference: CLI STDIN Payload Specifications

This reference specifies the exact STDIN JSON payload structures received from AI CLI host environments.

---

## 1. Antigravity (`agy`) Payload Spec

샘플 위치는 [`docs/samples/antigravity.json`](../../samples/antigravity.json)입니다.

```json
{
  "product": "antigravity",
  "version": "1.2.11",
  "cwd": "/workspace/statusline",
  "session_id": "00000000-0000-4000-8000-000000000000",
  "conversation_id": "00000000-0000-4000-8000-000000000000",
  "model": {
    "id": "Gemini 3.8 Flash (Medium)",
    "display_name": "Gemini 3.8 Flash (Medium)",
    "effort": "medium"
  },
  "workspace": {
    "current_dir": "/workspace/statusline",
    "project_dir": "/workspace/statusline"
  },
  "context_window": {
    "total_input_tokens": 44068,
    "total_output_tokens": 4420,
    "context_window_size": 1048576,
    "used_percentage": 4.20265,
    "remaining_percentage": 95.79735
  },
  "quota": {
    "gemini-5h": { "remaining_fraction": 0.752, "reset_in_seconds": 15236 },
    "gemini-weekly": { "remaining_fraction": 0.340, "reset_in_seconds": 400398 }
  },
  "agent_state": "working",
  "cycle_mode": "accept-edits",
  "plan_tier": "Google AI Pro",
  "email": "user@example.com"
}
```

어댑터가 읽는 field와 우선순위:

- `model`: `displayName` → `display_name` → `id` → `name`. 추론 수준은 `effort`
- 컨텍스트: `context_window.context_window_size`가 있으면 `used_percentage`를 토큰 수로 환산하고(0이면 `total_input_tokens`), 없으면 `used_percentage`를 백분율로 쓴다. 둘 다 없으면 `contextTokens`, `contextBar.percentage`
- 작업 상태: `thinkingState` → `thinking.state` → `agent_state`
- `activeSkills` → `active_skills`, `lastTool` → `last_tool`, `cwd` → `workspace.current_dir`
- `quota.<key>`: `remaining_fraction`, `reset_time` → `resetTime`(RFC3339). 둘 다 없으면 `reset_in_seconds` → `resetInSeconds`를 현재 시각에 더한다. key prefix가 `gemini-`이면 `gemini`, `3p-`이면 `3rd` segment가 된다
- `cycle_mode`, `sandbox.enabled`, `tool_confirmation_pending`, `terminal_width`

---

## 2. Claude Code Payload Spec

Sample location: [`docs/samples/claude.json`](../../samples/claude.json) — OMC 래퍼 형식(`displayName`) 샘플. Native CC 필드(`display_name`/`id`)는 상기 스펙 참조.

```json
{
  "model": { "id": "glm-5.3", "display_name": "GLM-5.3" },
  "contextBar": { "percentage": 45 },
  "thinking": { "state": "thinking" },
  "activeSkills": ["superpowers:brainstorming"],
  "lastTool": "view_file"
}
```

Native CC stdin은 `model.id`/`model.display_name`(snake_case)을 보낸다. 모델명 우선순위: `display_name` → `displayName`(OMC 래퍼) → `name` → `id`.

추가로 읽는 field: `context_window.total_input_tokens`, `context_window.context_window_size`, `contextTokens`, `thinkingState`, `cwd`, `permission`, `terminal_width`.

---

## 3. Codex Widget Snapshot

Sample: [`docs/samples/codex.json`](../../samples/codex.json). This is a normalized local-session snapshot for the future widget, not a Codex CLI STDIN payload. `model`, `cwd`, and `effort` come from the latest `turn_context`; `info` and `rate_limits` come from the latest `event_msg` `token_count`. The events can have different timestamps.

- `info.total_token_usage`: cumulative session tokens; it can exceed one context window.
- `info.last_token_usage`: latest reported model call, used as an approximation for the context percentage.
- `info.model_context_window`: model context capacity.
- `rate_limits.primary` and `secondary`: `used_percent`, `window_minutes`, and `resets_at` (Unix seconds). The adapter reads `used_percent` and `resets_at` and renders remaining percentages as `100 - used_percent`.
- Cost is unavailable in this local sample. OpenAI Platform organization costs and Codex subscription limits are separate data sources.
- Renderer snapshot fields sent by the MCP statusline provider (`integrations/codex/statusline.py`): `statusline_snapshot`, `speed`, `permission`, `context_used`, `context_token_count`, `reset_tickets`, `reset_tickets_expire`. Setup is in [Codex MCP 상태바 설정](../how-to-guides/use-codex.md#6-codex-mcp-상태바-설정).
