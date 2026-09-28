# Reference: CLI STDIN Payload Specifications

This reference specifies the exact STDIN JSON payload structures received from AI CLI host environments.

---

## 1. Antigravity (`agy`) Payload Spec

Sample location: [`docs/samples/antigravity.json`](file:///Users/crong/git/statusline/docs/samples/antigravity.json)

```json
{
  "product": "antigravity",
  "version": "1.2.11",
  "cwd": "/Users/crong/git/statusline",
  "session_id": "b2443d60-4e81-4fb4-b229-d084868b9f98",
  "conversation_id": "b2443d60-4e81-4fb4-b229-d084868b9f98",
  "model": {
    "id": "Gemini 3.8 Flash (Medium)",
    "display_name": "Gemini 3.8 Flash (Medium)",
    "effort": "medium"
  },
  "workspace": {
    "current_dir": "/Users/crong/git/statusline",
    "project_dir": "/Users/crong/git/statusline"
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
  "email": "deuxksy@gmail.com"
}
```

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

---

## 3. Codex Widget Snapshot

Sample: [`docs/samples/codex.json`](../../samples/codex.json). This is a normalized local-session snapshot for the future widget, not a Codex CLI STDIN payload. `model`, `cwd`, and `effort` come from the latest `turn_context`; `info` and `rate_limits` come from the latest `event_msg` `token_count`. The events can have different timestamps.

- `info.total_token_usage`: cumulative session tokens; it can exceed one context window.
- `info.last_token_usage`: latest reported model call, used as an approximation for the context percentage.
- `info.model_context_window`: model context capacity.
- `rate_limits.primary` and `secondary`: used percentages, window durations, and Unix reset times. The adapter renders remaining percentages by subtracting from 100.
- Cost is unavailable in this local sample. OpenAI Platform organization costs and Codex subscription limits are separate data sources.
