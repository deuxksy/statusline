# Tutorial: Getting Started with statusline v0.2

This tutorial guides you through building, configuring, and testing `statusline` v0.2.

---

## 1. Prerequisites

- Go 1.21+ installed
- Terminal emulator (e.g. WezTerm, iTerm2, Kitty) or `tmux`

## 2. Build statusline

Clone the repository and compile the binary:

```bash
cd /home/crong/git/statusline
go build -o statusline ./cmd/statusline
```

To install the binary to your local executable path (`~/.local/bin`):

```bash
cp ./statusline ~/.local/bin/statusline
```

## 3. Initialize Configuration

Initialize the default configuration file (`~/.config/statusline/config.json`):

```bash
./statusline init
```

Output:
`Initialized default config at /home/crong/.config/statusline/config.json`

## 4. Test statusline Pipeline

Send a sample Antigravity (`agy`) payload via STDIN:

```bash
echo '{"product":"antigravity","model":{"display_name":"Gemini 3.6 Flash (Medium)"},"context_window":{"used_percentage":5.2},"agent_state":"tool_use"}' | ./statusline --cli=auto
```

Expected Output:
`statusline  main * │  ANTIGRAVITY  │ Gemini 3.6 Flash (Medium) │ 🧠 tool_use │ [5%]`

Congratulations! You have set up and verified `statusline` v0.2.
