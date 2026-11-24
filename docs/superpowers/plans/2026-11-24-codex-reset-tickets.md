# Codex 초기화 티켓 구현 계획

**목표** Codex CLI 상태바에서 초기화 티켓 잔여 수와 가장 이른 만료일을 `🎟️x2(12/22)` 형식으로 표시한다.

**설계** [Universal StatusLine 설계](../specs/2026-08-01-universal-statusline-design.md#codex-초기화-티켓-표시)

- [x] 공식 `create-statusline` Python 예제를 기반으로 로컬 stdio provider를 작성한다.
- [x] 세션 스냅샷을 기존 Go 렌더러로 전달하고 stdout을 Codex UI 트리로 변환한다.
- [x] 로컬 app-server의 `account/rateLimits/read`를 백그라운드에서 조회해 티켓을 표시한다.
- [x] Codex의 Custom status line 기능을 확인하고 MCP provider를 등록한다.
- [x] MCP 입출력, 티켓 표시 규칙, 실제 TUI를 검증한다.
