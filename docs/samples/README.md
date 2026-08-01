# CLI Statusline STDIN Payload Samples

This directory stores real STDIN JSON payloads captured from different AI CLI tools (`antigravity`, `claude`, `codex`).

## Files

- [`antigravity.json`](file:///home/crong/git/statusline/docs/samples/antigravity.json) : Real STDIN payload from Google Antigravity (agy) CLI.
- [`claude.json`](file:///home/crong/git/statusline/docs/samples/claude.json) : Sample STDIN payload from Claude Code / OMC.
- [`codex.json`](file:///home/crong/git/statusline/docs/samples/codex.json) : Sample STDIN payload from Codex CLI.

## How to Test Statusline Output

```bash
# Test Antigravity
cat docs/samples/antigravity.json | ./statusline --cli=antigravity

# Test Claude
cat docs/samples/claude.json | ./statusline --cli=claude

# Test Codex
cat docs/samples/codex.json | ./statusline --cli=codex

# Test Auto Discriminator
cat docs/samples/antigravity.json | ./statusline --cli=auto
```
