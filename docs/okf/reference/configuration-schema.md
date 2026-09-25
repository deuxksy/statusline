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

---

## Credential File (`credentials.json`)

Optional local credential store, loaded from `~/.config/statusline/credentials.json`. The file must have `0600` permissions (group/world-readable files are rejected). Environment variables always take precedence over the file.

| Field | Type | Used by | Precedence (highest first) |
|---|---|---|---|
| `openai_admin_key` | string | OpenAI Platform collect | `OPENAI_ADMIN_KEY` env → file |
| `zai_auth_token` | string | Z.AI quota collect & self-refresh | `ZAI_AUTH_TOKEN` → `ANTHROPIC_AUTH_TOKEN` (inherited from Claude Code) → file |
| `zai_base_url` | string | Z.AI endpoint override | `ZAI_BASE_URL` → `ANTHROPIC_BASE_URL` → file → default |

`zai_base_url` defaults to `https://api.z.ai/api/anthropic` when nothing is set. The `ZAI_*` variables exist to avoid clashing with Claude Code's own `ANTHROPIC_*` environment when both are present (e.g. an official Anthropic setup). Provider detection only accepts `api.z.ai` / `bigmodel.cn` hosts, so a non-Z.AI `ANTHROPIC_BASE_URL` never sends credentials to an unrecognized server.

```json
{
  "openai_admin_key": "sk-admin-...",
  "zai_auth_token": "...",
  "zai_base_url": "https://api.z.ai/api/anthropic"
}
```
