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
