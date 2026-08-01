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
