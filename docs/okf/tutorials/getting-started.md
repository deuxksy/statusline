# Tutorial: Getting Started with statusline

This tutorial guides you through building, configuring, and testing `statusline`.

---

## 1. Prerequisites

- Go 1.26.5 or later installed (the version required by `go.mod`)
- Terminal emulator (e.g. WezTerm, iTerm2, Kitty) or `tmux`

## 2. Build statusline

Clone the repository and compile the binary:

```bash
git clone https://github.com/deuxksy/statusline.git
cd statusline
go build -o statusline ./cmd/statusline
```

To install the binary to your local executable path (`~/.local/bin`):

```bash
cp ./statusline ~/.local/bin/statusline
```

## 3. Initialize Configuration

Initialize the default configuration file (`~/.config/statusline/config.json`):

Skip this step if the file already exists: `init` overwrites it with the default configuration.

```bash
./statusline init
```

Output:
`Initialized default config at ~/.config/statusline/config.json`

## 4. Test statusline Pipeline

Send a sample Antigravity (`agy`) payload via STDIN:

```bash
echo '{"product":"antigravity","model":{"display_name":"Gemini 3.6 Flash (Medium)"},"context_window":{"used_percentage":5.2},"agent_state":"tool_use"}' | ./statusline --cli=auto
```

Expected Output:
`ANTIGRAVITY  │ statusline  main            │ Gemini 3.6 Flash (Medium) │ 🧠 tool_use │ [5%]`

Congratulations! You have set up and verified `statusline`.

## Windows 11 PowerShell

저장소 루트에서 `statusline.exe`를 빌드한 뒤 사용자 실행 파일 디렉터리에 복사한다. Codex MCP provider는 저장소 루트의 `statusline.exe`를 사용한다. Claude Code와 Antigravity CLI는 복사한 실행 파일을 사용한다.

```powershell
go build -o statusline.exe ./cmd/statusline
$binDir = Join-Path $env:USERPROFILE '.local\bin'
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
Copy-Item -LiteralPath '.\statusline.exe' -Destination $binDir
$configPath = Join-Path $env:USERPROFILE '.config\statusline\config.json'
if (-not (Test-Path -LiteralPath $configPath)) {
    .\statusline.exe init
}
```

각 샘플을 렌더링해 출력이 `ANTIGRAVITY`, `CLAUDE`, `CODEX`로 시작하는지 확인한다.

```powershell
Get-Content -Raw -Encoding utf8 docs/samples/antigravity.json | .\statusline.exe --cli=antigravity
Get-Content -Raw -Encoding utf8 docs/samples/claude.json | .\statusline.exe --cli=claude
Get-Content -Raw -Encoding utf8 docs/samples/codex.json | .\statusline.exe --cli=codex
```

실제 CLI에 적용하려면 [Antigravity CLI](../how-to-guides/configure-elements.md#2-antigravity-cli-integration), [Claude Code](../how-to-guides/configure-elements.md#3-claude-code-integration), [Codex CLI](../how-to-guides/use-codex.md#6-codex-mcp-상태바-설정) 설정을 따른다.
