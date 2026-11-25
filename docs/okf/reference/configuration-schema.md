# Reference: Configuration Schema (`config.json`)

The configuration file is loaded from `~/.config/statusline/config.json`.

---

## Field Reference

### Top-level

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `$schema` | string | `"1.1"` | Config version. Files marked `"1.0"` or left unversioned are migrated in memory: `cycleMode` and `sandbox` are inserted after `model` in `layout.main` |
| `wrapMode` | string | `"truncate"` | Truncate the line with `…` when it exceeds the terminal width |
| `theme` | string | `"sleek_dark"` | `sleek_dark`, `light`, `nord`, `monokai-dark` |

### `elements` (object)

| Field | Type | Default | Description |
|---|---|---|---|
| `engineLabel` | boolean | `true` | Show engine badge (e.g. `ANTIGRAVITY`, `CLAUDE`, `CODEX`) |
| `model` | boolean | `true` | Show model display name |
| `modelFormat` | string | `"full"` | `short` abbreviates known model names (`flash-med`, `sonnet-3.7`, `opus`, `haiku`); other names longer than 12 characters are cut to 10 characters followed by `..` |
| `gitRepo` | boolean | `true` | Show top-level Git repository folder name |
| `gitBranch` | boolean | `true` | Show current Git branch (` main`) |
| `gitStatus` | boolean | `true` | Show Git dirty status (`*`) |
| `cwd` | boolean | `true` | Show current working directory. Hidden when `gitRepo` already shows the same folder name |
| `cwdFormat` | string | `"folder"` | `folder` shows the base name; any other value shows the full path |
| `hostname` | boolean | `false` | Show machine hostname |
| `contextBar` | boolean | `true` | Show context usage: `⚡ 44k/1.0M(4%)` when the context window size is known, `[5%]` otherwise. Requires `showTokens` |
| `showTokens` | boolean | `true` | Show token counts. With `contextBar` off, the `tokens` layout item shows `44000 tokens` instead |
| `activeSkills` | boolean | `true` | Show active skills (`⚡ superpowers:brainstorming`) |
| `lastTool` | boolean | `true` | Show the last tool call (`🔧 view_file`) |
| `thinking` | boolean | `true` | Show agent/thinking state (`🧠 tool_use`) |
| `quota` | boolean | `true` | Show the provider quota segment (`📊 gemini 5h:75% wk:34%`, `📊 zai 5h:95%(2h40m) mo:91%(19d22h) 🎟️x2(12/22)`) |
| `permission` | boolean | `true` | Show Codex permission (`🙋`, `🤖`, `🔓`, `👁️`) and `🔒 WAITING` while a permission request is pending |
| `cycleMode` | boolean | `true` | Show Antigravity cycle mode (`👁️` plan, `🤖` accept-edits, `🔓` auto) or Codex speed (`🚶` standard, `🏃` fast) |
| `sandbox` | boolean | `true` | Show `🛡️ sandbox` when the Antigravity sandbox is enabled |

### `layout` (object)

- `line1`: List of elements to render (merged onto single line in statusbar mode).
- `main`: List of main status bar elements.
- `line1Width`: Fixed width of the `line1` group in characters (default `28`). `0` disables padding and truncation.

Item names: `hostname`, `cwd`, `gitRepo`, `gitBranch`, `gitStatus`, `permission`, `engineLabel`, `model`, `cycleMode`, `sandbox`, `thinking`, `contextBar`, `tokens`, `quota`, `activeSkills`, `lastTool`. The engine badge is always rendered first, wherever `engineLabel` appears. For the Codex engine, `permission` is rendered right after `cycleMode` (or `model`).

### `thresholds` (object)

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `contextWarning` | integer | `70` | Context usage percentage at which the `⚡` segment turns warning color |
| `contextCritical` | integer | `85` | Context usage percentage at which the `⚡` segment turns critical color |

### `tray` (object)

macOS 메뉴바와 Windows 시스템 트레이에서 실행하는 `statusline tray`의 설정이다. 위젯 메뉴에서 메인 provider를 전환하면 `primary`를 설정 파일에 저장한다.

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
