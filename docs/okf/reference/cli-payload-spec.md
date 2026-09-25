# Reference: CLI STDIN Payload Specifications

This reference specifies the exact STDIN JSON payload structures received from AI CLI host environments.

---

## 1. Antigravity (`agy`) Payload Spec

Sample location: [`docs/samples/antigravity.json`](file:///home/crong/git/statusline/docs/samples/antigravity.json)

```json
{
  "product": "antigravity",
  "version": "1.1.8",
  "cwd": "/home/crong/git/statusline",
  "session_id": "b18fda04-43ae-4c34-b7f3-836db300be59",
  "conversation_id": "b18fda04-43ae-4c34-b7f3-836db300be59",
  "model": {
    "id": "Gemini 3.6 Flash (Medium)",
    "display_name": "Gemini 3.6 Flash (Medium)",
    "effort": "medium"
  },
  "workspace": {
    "current_dir": "/home/crong/git/statusline",
    "project_dir": "/home/crong/git/statusline"
  },
  "context_window": {
    "total_input_tokens": 50527,
    "total_output_tokens": 9665,
    "context_window_size": 1048576,
    "used_percentage": 4.81863,
    "remaining_percentage": 95.18137
  },
  "quota": {
    "gemini-5h": { "remaining_fraction": 0.933, "reset_in_seconds": 12740 },
    "gemini-weekly": { "remaining_fraction": 0.528, "reset_in_seconds": 335605 }
  },
  "agent_state": "tool_use",
  "plan_tier": "Google AI Pro",
  "email": "deuxksy@gmail.com"
}
```

---

## 2. Claude Code Payload Spec

Sample location: [`docs/samples/claude.json`](file:///home/crong/git/statusline/docs/samples/claude.json)

```json
{
  "omcLabel": "OMC",
  "model": { "displayName": "Claude 3.7 Sonnet" },
  "contextBar": { "percentage": 45 },
  "thinking": { "state": "thinking" },
  "activeSkills": ["superpowers:brainstorming"],
  "lastTool": "view_file"
}
```

---

## 3. Codex Widget Snapshot

Sample: [`docs/samples/codex.json`](../../samples/codex.json). This is a normalized local-session snapshot for the future widget, not a Codex CLI STDIN payload. `model`, `cwd`, and `effort` come from the latest `turn_context`; `info` and `rate_limits` come from the latest `event_msg` `token_count`. The events can have different timestamps.

- `info.total_token_usage`: cumulative session tokens; it can exceed one context window.
- `info.last_token_usage`: latest reported model call, used as an approximation for the context percentage.
- `info.model_context_window`: model context capacity.
- `rate_limits.primary` and `secondary`: used percentages, window durations, and Unix reset times. The adapter renders remaining percentages by subtracting from 100.
- Cost is unavailable in this local sample. OpenAI Platform organization costs and Codex subscription limits are separate data sources.
