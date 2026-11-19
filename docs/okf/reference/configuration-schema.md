# Reference: Configuration Schema (`config.json`)

The configuration file is loaded from `~/.config/statusline/config.json`.

---

## Field Reference

### `elements` (object)

| Field | Type | Default | Description |
|---|---|---|---|
| `engineLabel` | boolean | `true` | Show engine badge (e.g. `ANTIGRAVITY`, `CLAUDE`, `CODEX`) |
| `model` | boolean | `true` | Show model display name |
| `gitRepo` | boolean | `true` | Show top-level Git repository folder name |
| `gitBranch` | boolean | `true` | Show current Git branch (` main`) |
| `gitStatus` | boolean | `true` | Show Git dirty status (`*`) |
| `cwd` | boolean | `true` | Show current working directory folder name |
| `hostname` | boolean | `false` | Show machine hostname |
| `contextBar` | boolean | `true` | Show context usage percentage (`[5%]`) |
| `thinking` | boolean | `true` | Show agent/thinking state (`🧠 tool_use`) |

### `layout` (object)

- `line1`: List of elements to render (merged onto single line in statusbar mode).
- `main`: List of main status bar elements.

### `tray` (object)

Tray widget (`statusline tray`, v0.8.1) settings. `primary` is saved back to the config file when the main provider is switched from the tray dropdown.

| Field | Type | Default | Description |
|---|---|---|---|
| `primary` | string | `"zai"` | Provider shown in the menu bar title (`zai`, `chatgpt`). Invalid values fall back to `zai`. |

---

## Credential File (`credentials.json`)

Optional local credential store, loaded from `~/.config/statusline/credentials.json`. The file must have `0600` permissions (group/world-readable files are rejected). Environment variables always take precedence over the file.

| Field | Type | Used by | Precedence (highest first) |
|---|---|---|---|
| `openai_admin_key` | string | OpenAI Platform collect | `OPENAI_ADMIN_KEY` env → file |
| `zai_auth_token` | string | Z.AI quota collect & self-refresh | `ZAI_AUTH_TOKEN` → `ANTHROPIC_AUTH_TOKEN` (inherited from Claude Code) → file |
| `zai_base_url` | string | Z.AI endpoint override | `ZAI_BASE_URL` → `ANTHROPIC_BASE_URL` → file → default |

**Reset tickets:** remaining 1-time Reset tickets are fetched automatically during `collect` from the Z.AI console API (`GET /api/biz/customer-package-reset/list?targetType=PERSONAL`, same auth token, undocumented endpoint) and rendered as `🎟️xN(MM/DD)` next to the quota segment. When expiry is within 3 days the segment turns red (`QuotaBad`). No configuration needed; if the endpoint disappears, tickets silently drop to 0 while quota collection continues (fail-soft).

`elements.zaiTickets` (int, default 0) — remaining 1-time Reset tickets to render as `🎟️xN` next to the quota segment. The Z.AI quota API does not expose ticket state, so keep this count in sync with the dashboard manually (set to 0 when used up).

`zai_base_url` defaults to `https://api.z.ai/api/anthropic` when nothing is set. The `ZAI_*` variables exist to avoid clashing with Claude Code's own `ANTHROPIC_*` environment when both are present (e.g. an official Anthropic setup). Provider detection only accepts `api.z.ai` / `bigmodel.cn` hosts, so a non-Z.AI `ANTHROPIC_BASE_URL` never sends credentials to an unrecognized server.

```json
{
  "openai_admin_key": "sk-admin-...",
  "zai_auth_token": "...",
  "zai_base_url": "https://api.z.ai/api/anthropic"
}
```

**Token rotation note:** when the Z.AI token is rotated (env `ANTHROPIC_AUTH_TOKEN` renewed), update `zai_auth_token` in this file too. The render path spawns a detached `collect` that may run without `ANTHROPIC_AUTH_TOKEN` in its environment and falls back to this file; a stale token makes the quota endpoint return HTTP 200 with no `limits` payload (not 401), so the failure is silent — only the refresh marker (`zai.json.refresh`) is left behind and the quota segment disappears. Quick check for this exact case:

```bash
env -i HOME="$HOME" PATH="$PATH" statusline collect --provider=zai
# {"errors":{"zai":"quota/limit: limits not found"}} → stale token in credentials.json
```
