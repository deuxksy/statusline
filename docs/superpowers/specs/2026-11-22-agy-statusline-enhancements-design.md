# Antigravity (`agy`) Statusline 확장 설계

## 1. 목적 (Intent & Why)
`agy` CLI가 stdin JSON으로 이미 전달하는 `cycle_mode`, `sandbox.enabled`, `context_window.context_window_size`, `model.effort`를 상태바에 반영해, 작업 중 현재 권한/계획 모드·샌드박스 활성 여부·컨텍스트 사용량/한도·추론 강도를 한눈에 확인한다.

## 2. 범위 (Goals / Non-Goals / Anti-Goals)
- **Goals**:
  1. `cycle_mode` (`plan` / `accept-edits` / `auto`) 배지 표시
  2. `sandbox.enabled` 활성 시 `🛡️ sandbox` 표시
  3. `context_window`의 `context_window_size` 매핑으로 `⚡ 44k/1M(4%)` 포맷 및 임계치 색상 적용
  4. `model.effort`를 `model` 세그먼트에 통합 표시 (중복 방지)
  5. 구버전 `config.json`에서 누락된 신규 기본 세그먼트를 `layout.main`에 자동 보완
  6. `engineLabel`(`ANTIGRAVITY` 등)을 상태바 맨 앞(`line1` 이전)에 배치
- **Non-Goals**: `cost`, `subagents`, `vim`, `task_count` 등 기타 `agy` 필드 및 `tray` 위젯 UI 변경.
- **Anti-Goals**: `"Gemini 4 Hydrogen (Medium)"`처럼 이미 effort가 포함된 모델명에 `(med)`를 중복 표기하지 않는다. 최신 스키마에서 사용자가 의도적으로 제거한 세그먼트는 강제 복원하지 않는다.

## 3. 접근 방식 (Approaches)
- **기존 `config.json` 호환성**:
  - 대안 A: `DefaultConfig()`만 갱신하고 기존 사용자는 수동 편집에 의존 → 업그레이드 직후 기능이 안 보이는 문제.
  - **채택 (방식 B)**: 현재 스키마 버전을 `"1.1"`로 올리고, `LoadConfig` 시 `$schema`가 `"1.1"` 미만(`""` 또는 `"1.0"`)인 설정 파일에 한해 `layout.main`에 `"cycleMode"`, `"sandbox"`가 없으면 `"model"` 뒤에 자동 삽입하고 스키마를 `"1.1"`로 간주한다.
- **`ReasoningEffort` 표시**:
  - 대안 2: 별도 세그먼트(`reasoningEffort`) 추가 → `│` 구분자 증가 및 모델명과 중복 위험.
  - **채택 (방식 1)**: `model` 세그먼트 렌더링 시 `st.ReasoningEffort`를 `(low)`/`(med)`/`(high)`/`(max)` 형태로 접미 표기하되, 포맷된 모델명(대소문자 무시)에 해당 effort 문자열(예: `medium`/`med`, `low`, `high`, `max`)이 이미 포함돼 있으면 생략한다.

## 4. 상세 설계 (Components)
- **`internal/model/status.go`**: `CycleMode string`, `Sandbox bool` 필드 (반영됨).
- **`internal/adapter/antigravity.go`**:
  - `cycle_mode` → `st.CycleMode`, `sandbox.enabled` → `st.Sandbox` (반영됨).
  - `model.effort` → `st.ReasoningEffort` 매핑.
  - `context_window.context_window_size > 0`일 때 `st.ContextLimit = context_window_size` 설정, `st.ContextTokens`는 `used_percentage` 기반 환산(`round(used_percentage / 100 * context_window_size)`) 우선, 없으면 `total_input_tokens` 폴백. `context_window_size == 0`이면 기존 퍼센트 폴백 유지.
- **`internal/config/config.go`**:
  - `DefaultConfig().Schema`를 `"1.1"`로 상향, `Elements.CycleMode`/`Elements.Sandbox` 및 `Layout.Main` 기본값 포함.
  - `LoadConfig`에서 구버전 스키마(`""` 또는 `"1.0"`) 감지 시 `layout.main`에 누락된 `"cycleMode"`, `"sandbox"` 보완.
- **`internal/render/renderer.go`**:
  - `cycleMode` (`📋 plan`, `✏️ accept-edits`, `⚡ auto`), `sandbox` (`🛡️ sandbox`) 렌더링 (반영됨).
  - `model` 세그먼트에 `formatModelWithEffort(displayName, st.ReasoningEffort)` 적용.
  - `Render`에서 `engineLabel`을 `line1` 앞 첫 세그먼트로 배치하고 `line1`/`main` 순회 시 중복 렌더링 방지.
