# Explanation: Core Design Decisions in statusline v0.2

This document explains the architectural principles and design tradeoffs behind `statusline` v0.2.

---

## 1. Single-Line Statusbar Output Rationale

CLI tools (such as Antigravity CLI and Claude Code) display status updates on a single-line statusbar HUD at the bottom of the screen.

When statusline outputs multiple lines separated by `\n`, host CLI widgets often truncate or discard all lines except the last line. To prevent losing critical context (such as Git branch, directory, or model info), `statusline` v0.2 formats all active segments into a single cohesive line joined by ` │ `.

## 2. Fail-Soft Principle & Performance Budget (<5ms)

- **Fail-Soft**: If STDIN is empty, corrupted, or unexpected, statusline never outputs tracebacks or stderr crashes. It returns a clean minimal fallback (`ANTIGRAVITY` badge or standard status).
- **10ms VCS Timeout**: Git commands (`git rev-parse`, `git status`) are executed with a strict 10ms context deadline so that statusline rendering never slows down terminal interactive responsiveness.
- **KISS & DRY**: No persistent background daemons, zero external database dependencies. Pure stream filter CLI.
