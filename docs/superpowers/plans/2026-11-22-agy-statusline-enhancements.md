# Antigravity (`agy`) Statusline Enhancements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Antigravity (`agy`) statusline enhancements: schema-based `config.json` migration for `cycleMode`/`sandbox`, detailed `context_window` token/limit mapping, and deduplicated `ReasoningEffort` rendering in the `model` segment.

**Architecture:** `cycleMode` and `sandbox` model/adapter/config/render primitives are already in place. This plan adds three focused, independent improvements across `internal/config`, `internal/adapter`, and `internal/render`, keeping the render path pure I/O within the <5ms budget.

**Tech Stack:** Go, `charmbracelet/lipgloss`, `charmbracelet/x/ansi`, Go `testing`.

**Spec:** [2026-11-22-agy-statusline-enhancements-design.md](file:///Users/crong/git/statusline/docs/superpowers/specs/2026-11-22-agy-statusline-enhancements-design.md)

## Global Constraints

- Render path is a pure I/O stream filter (<5ms budget, no network/daemon calls).
- Do not duplicate effort when `model` display name already contains it (case-insensitive, e.g., `"Gemini 4 Hydrogen (Medium)"`, `"flash-med"`).
- Do not re-add `"cycleMode"` or `"sandbox"` if `$schema` is `"1.1"` or newer.
- Preserve `gofmt` formatting and pass `go test ./...`.

## Review Focus

- Legacy `config.json` (`$schema` empty or `"1.0"`) with custom `layout.main` missing `"model"` → append `"cycleMode"`, `"sandbox"` cleanly without panicking.
- Legacy `config.json` that already includes `"cycleMode"` → do not insert a duplicate entry.
- Modern `config.json` (`$schema: "1.1"`) where user intentionally removed `"sandbox"` from `layout.main` → keep it removed.
- `antigravity` payload with `context_window.context_window_size == 0` → keep existing percentage fallback (`st.ContextLimit == 0`).
- `modelFormat: "short"` producing `"flash-med"` with `effort: "medium"` → do not render `"flash-med (med)"`.

---

### Task 1: Config Schema `"1.1"` Migration (`internal/config`)

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Produces: `DefaultConfig().Schema == "1.1"`, `LoadConfig(path string) *Config` migrating `layout.main` for `$schema < "1.1"`.

- [ ] **Step 1: Write failing tests in `internal/config/config_test.go`**
  - `TestLoadConfigMigratesLegacyLayout`:
    - Legacy config (`{"$schema":"1.0","layout":{"main":["permission","engineLabel","model","thinking"]}}`) → `cfg.Layout.Main` becomes `["permission","engineLabel","model","cycleMode","sandbox","thinking"]` and `cfg.Schema == "1.1"`.
    - Legacy config with no `$schema` and `layout.main` without `"model"` (`["engineLabel","thinking"]`) → appends `"cycleMode"`, `"sandbox"` without panic.
    - Legacy config already containing `"cycleMode"` → does not duplicate `"cycleMode"`, only inserts missing `"sandbox"`.
    - Current schema (`{"$schema":"1.1","layout":{"main":["engineLabel","model"]}}`) → preserves `["engineLabel","model"]` untouched.

- [ ] **Step 2: Run test to verify it fails**
  - Run: `go test -v ./internal/config -run TestLoadConfigMigratesLegacyLayout`
  - Expected: FAIL (schema is still `"1.0"` and `layout.main` is not migrated).

- [ ] **Step 3: Implement schema `"1.1"` and migration in `internal/config/config.go`**
  - Set `DefaultConfig().Schema = "1.1"`.
  - Add helper `migrateLayoutMain(main []string, schema string) []string`: if `schema == "" || schema == "1.0"`, ensure `"cycleMode"` and `"sandbox"` are present (inserting after `"model"` if found, else appending).
  - Call `migrateLayoutMain` in `LoadConfig` and set `cfg.Schema = "1.1"` when migrated.

- [ ] **Step 4: Run tests to verify pass**
  - Run: `go test -v ./internal/config`
  - Expected: PASS.

---

### Task 2: Antigravity `context_window` Limit & `model.effort` Parsing (`internal/adapter`)

**Files:**
- Modify: `internal/adapter/antigravity.go`
- Test: `internal/adapter/adapter_test.go`

**Interfaces:**
- Produces: `st.ContextLimit` (`int`), `st.ContextTokens` (`int`), `st.ReasoningEffort` (`string`) on `*model.UnifiedStatus` from `AntigravityAdapter.Parse`.

- [ ] **Step 1: Write failing tests in `internal/adapter/adapter_test.go`**
  - Update/add `TestAntigravityAdapterContextWindowAndEffort`:
    - Parse `docs/samples/antigravity.json`: assert `status.ContextLimit == 1048576`, `status.ContextTokens == 44068` (or `int(math.Round(4.2026519775390625 / 100 * 1048576))` == `44068`), and `status.ReasoningEffort == "medium"`.
    - Payload with `context_window_size: 0` and `used_percentage: 45`: assert `status.ContextLimit == 0` and `status.ContextTokens == 45`.
    - Payload with `context_window_size: 200000`, `used_percentage: 0`, `total_input_tokens: 12000`: assert `status.ContextLimit == 200000` and `status.ContextTokens == 12000`.

- [ ] **Step 2: Run test to verify it fails**
  - Run: `go test -v ./internal/adapter -run TestAntigravityAdapterContextWindowAndEffort`
  - Expected: FAIL (`ContextLimit` is `0`, `ReasoningEffort` is `""`).

- [ ] **Step 3: Implement parsing in `internal/adapter/antigravity.go`**
  - Add `ContextWindowSize int` to `raw.ContextWindow` (`json:"context_window_size"`).
  - Add `Effort string` to `modelObj` (`json:"effort"`), assign `st.ReasoningEffort = modelObj.Effort`.
  - When `raw.ContextWindow.ContextWindowSize > 0`:
    - `st.ContextLimit = raw.ContextWindow.ContextWindowSize`
    - If `raw.ContextWindow.UsedPercentage > 0`: `st.ContextTokens = int(math.Round(raw.ContextWindow.UsedPercentage / 100.0 * float64(raw.ContextWindow.ContextWindowSize)))`
    - Else if `raw.ContextWindow.TotalInputTokens > 0`: `st.ContextTokens = raw.ContextWindow.TotalInputTokens`
  - Keep existing fallback when `ContextWindowSize == 0`.

- [ ] **Step 4: Run tests to verify pass**
  - Run: `go test -v ./internal/adapter`
  - Expected: PASS (including `TestAntigravityAdapterRealPayload` where `context_window_size` is omitted).

---

### Task 3: Render `ReasoningEffort` in `model` Segment (`internal/render`)

**Files:**
- Modify: `internal/render/renderer.go`
- Test: `internal/render/renderer_test.go`

**Interfaces:**
- Consumes: `st.Model`, `st.ReasoningEffort`, `cfg.Elements.ModelFormat`
- Produces: `formatModelWithEffort(displayName, effort string) string` used by `renderSegment("model", ...)`.

- [ ] **Step 1: Write failing tests in `internal/render/renderer_test.go`**
  - `TestRenderModelWithReasoningEffort`:
    - `st.Model = "gpt-6-astra"`, `st.ReasoningEffort = "low"` → contains `"gpt-6-astra (low)"`.
    - `st.Model = "gpt-6-astra"`, `st.ReasoningEffort = "medium"` → contains `"gpt-6-astra (med)"`.
    - `st.Model = "Gemini 4 Hydrogen (Medium)"`, `st.ReasoningEffort = "medium"` → contains `"Gemini 4 Hydrogen (Medium)"` and NOT `"(med)"`.
    - `st.Model = "Gemini 3.8 Flash (Medium)"`, `cfg.Elements.ModelFormat = "short"`, `st.ReasoningEffort = "medium"` → `"flash-med"` without `"(med)"`.
    - `st.Model = "Gemini 4 Hydrogen"`, `st.ReasoningEffort = ""` → `"Gemini 4 Hydrogen"` without parentheses.

- [ ] **Step 2: Run test to verify it fails**
  - Run: `go test -v ./internal/render -run TestRenderModelWithReasoningEffort`
  - Expected: FAIL (`gpt-6-astra` does not include `(low)`/`(med)`).

- [ ] **Step 3: Implement `formatModelWithEffort(displayName, effort string) string` in `internal/render/renderer.go`**
  - Normalize `effort` (`"medium"` → `"med"`, `"low"`/`"high"`/`"max"` as-is).
  - If `effort == ""` or `strings.Contains(strings.ToLower(displayName), strings.ToLower(effort))` or `strings.Contains(strings.ToLower(displayName), short)` → return `displayName`.
  - Otherwise return `displayName + " (" + short + ")"`.
  - Call `formatModelWithEffort(displayName, st.ReasoningEffort)` in `case "model":` of `renderSegment`.

- [ ] **Step 4: Run full verification**
  - Run: `go test -v ./internal/render`
  - Expected: PASS.

---

### Task 4: Place `engineLabel` First Before `line1` (`internal/render`)

**Files:**
- Modify: `internal/render/renderer.go`
- Test: `internal/render/renderer_test.go`

**Interfaces:**
- Consumes: `cfg.Elements.EngineLabel`, `st.EngineName`
- Produces: `Render(st, cfg)` outputting `<engineLabel> │ <line1> │ <main>` without duplicating `engineLabel` in `main`.

- [ ] **Step 1: Update/write failing tests in `internal/render/renderer_test.go`**
  - Update `TestRenderGitStatusFixedSpacing` so `line1` alignment check either disables `EngineLabel` or checks the separator after `line1` (`strings.LastIndex`).
  - Add `TestRenderEngineLabelFirst`:
    - `st := model.NewUnifiedStatus("antigravity")`, `st.GitRepo = "statusline"`, `st.GitBranch = "main"`, `st.Model = "Gemini 4 Hydrogen"`
    - Assert `ansi.Strip(render.Render(st, config.DefaultConfig()))` starts with `" ANTIGRAVITY "` (before `"statusline"`), and `"ANTIGRAVITY"` appears only once (`strings.Count(plain, "ANTIGRAVITY") == 1`).

- [ ] **Step 2: Run test to verify it fails**
  - Run: `go test -v ./internal/render -run TestRenderEngineLabelFirst`
  - Expected: FAIL (`statusline` appears before `ANTIGRAVITY`).

- [ ] **Step 3: Update `Render` in `internal/render/renderer.go`**
  - Render `engineSeg := renderSegment("engineLabel", st, cfg, palette)` before `line1`.
  - Skip `item == "engineLabel"` when iterating `cfg.Layout.Line1` and `cfg.Layout.Main` so it is never duplicated.
  - Assemble non-empty parts (`engineSeg`, `line1`, `main`) with `" │ "` separator before `wrapMode` truncation.

- [ ] **Step 4: Run full verification**
  - Run: `gofmt -l internal/ && go test ./... && go build -o statusline ./cmd/statusline && cat docs/samples/antigravity.json | ./statusline --cli=antigravity`
  - Expected: Clean `gofmt`, all tests PASS, integration output starts with `ANTIGRAVITY` and shows `✏️ accept-edits` and `⚡ 44k/1M(4%)`.
