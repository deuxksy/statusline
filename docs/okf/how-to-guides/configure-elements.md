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

In `~/.gemini/antigravity-cli/settings.json`, set the `statusLine` command:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.local/bin/statusline --cli=auto"
  }
}
```

**Windows 11**에서는 `%USERPROFILE%\.gemini\antigravity-cli\settings.json`의 `statusLine`을 다음처럼 설정한다. `C:\Users\me`를 실제 사용자 경로로 바꾸고 JSON 문자열의 역슬래시를 이중으로 표기한다. 저장 후 `agy`를 재시작해 하단을 확인한다.

```json
{
  "statusLine": {
    "type": "command",
    "command": "C:\\Users\\me\\.local\\bin\\statusline.exe --cli=antigravity",
    "enabled": true
  }
}
```

---

## 3. Claude Code Integration

In `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.local/bin/statusline --cli=claude"
  }
}
```

**Windows 11**에서는 `%USERPROFILE%\.claude\settings.json`의 `statusLine`을 다음처럼 설정한다. `C:/Users/me`를 실제 사용자 경로로 바꾸고 새 Claude Code 세션에서 하단을 확인한다.

```json
{
  "statusLine": {
    "type": "command",
    "command": "C:/Users/me/.local/bin/statusline.exe --cli=claude"
  }
}
```

---

## 4. Tmux Integration

In `~/.tmux.conf`:

```tmux
# Status bar configuration
set -g status-position bottom
set -g status-left "#[fg=black,bg=cyan,bold] #S #[default]"
set -g status-right "#[fg=yellow] #(hostname) #[fg=white]| %H:%M %Y-%m-%d "
```
