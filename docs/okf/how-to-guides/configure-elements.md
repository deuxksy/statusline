# How-To Guide: Configure Statusline Elements & Integration

This guide explains how to customize elements, disable redundant fields, and integrate `statusline` with `tmux` and Antigravity CLI.

---

## 1. Customizing `config.json`

Open `~/.config/statusline/config.json`:

```json
{
  "$schema": "1.0",
  "elements": {
    "engineLabel": true,
    "model": true,
    "gitRepo": true,
    "gitBranch": true,
    "gitStatus": true,
    "cwd": true,
    "hostname": false,
    "contextBar": true,
    "showTokens": true,
    "activeSkills": true,
    "lastTool": true,
    "thinking": true
  },
  "layout": {
    "line1": [],
    "main": [
      "cwd",
      "gitBranch",
      "gitStatus",
      "engineLabel",
      "model",
      "thinking",
      "contextBar",
      "tokens",
      "activeSkills",
      "lastTool"
    ]
  }
}
```

### Disabling Redundant Hostname

To prevent duplicating hostname when `tmux` already displays `hostname` in `status-right`:
Set `"hostname": false` in `elements`.

---

## 2. Antigravity CLI Integration

In `~/.gemini/antigravity-cli/settings.json`, set the `statusline` command:

```json
{
  "statusline": {
    "command": "/home/crong/.local/bin/statusline --cli=auto"
  }
}
```

---

## 3. Tmux Integration

In `~/.tmux.conf`:

```tmux
# Status bar configuration
set -g status-position bottom
set -g status-left "#[fg=black,bg=cyan,bold] #S #[default]"
set -g status-right "#[fg=yellow] #(hostname) #[fg=white]| %H:%M %Y-%m-%d "
```
