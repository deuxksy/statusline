# Windows 시스템 트레이 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**목표** `statusline.exe tray`에서 선택한 provider의 잔여율을 한눈에 확인할 수 있게 한다.

**구조** 공통 `tray.Title`을 데이터 원천으로 사용한다. Windows에서만 요약의 첫 숫자를 ICO로 렌더링하며, 전체 요약 툴팁과 메뉴 갱신은 기존 `trayApp.poll` 경로를 이용한다.

**기술 스택** Go 표준 라이브러리, `fyne.io/systray` v1.12.2.

**설계 문서** [Windows 시스템 트레이 설계](../specs/2026-11-25-windows-tray-design.md)

## 공통 제약

- 아이콘은 `0`~`100` 또는 `–`를 표시하며 ICO에는 16×16과 32×32 이미지를 포함한다. 고유 아이콘은 최대 102개다.
- 툴팁은 기존 요약의 전체 문자열을 사용한다. 메뉴, 5초 폴링, 캐시 TTL, 중복 수집 방지와 macOS 텍스트 표시를 유지한다.
- 별도 의존성 없이 구현한다.

## 리뷰 중점

- `0`과 데이터 없음이 아이콘에서 구분되는지 Task 1에서 확인한다.
- `100` 세 자리가 모두 16×16 안에 들어가는지 Task 1에서 확인한다.
- 비정상 요약을 누락 아이콘으로 처리하는지 Task 1에서 확인한다.
- 주간 값만 달라지거나 provider 간 잔여율이 같을 때 아이콘 재설정 없이 툴팁이 바뀌는지 Task 2에서 확인한다.
- 이전 잔여율로 돌아올 때 해당 아이콘이 다시 설정되는지 Task 2에서 확인한다.

### Task 1. 요약에서 숫자를 선택하고 ICO 생성

**파일** `internal/tray/icon.go`, `internal/tray/icon_test.go` 추가.

**인터페이스** `func IconLabel(title string) string` 및 `func Icon(label string) []byte`. `Icon` 입력 계약은 `IconLabel`이 반환하는 102개 값이다.

- [x] `TestIconLabel`을 작성한다. `z87 w45 → 87`, `c48 → 48`, `z0 → 0`, `c100 → 100`, `…/c-/z101/bad/빈 문자열 → –`를 검증한다.
- [x] `TestIcon`을 작성한다. `0`, `100`, `–` 모두 ICO 항목 수 `2`, 크기 `16/32`, 유효한 오프셋·길이와 32bpp를 검증한다. 래스터 픽셀을 확인해 `100`의 각 글리프가 16×16 안에서 구분되고 두 크기에서 `0`과 `–`가 서로 다른지 확인한다.
- [x] `go test ./internal/tray -run '^TestIcon' -count=1`을 실행해 구현 누락으로 실패하는 것을 확인한다.
- [x] `IconLabel`과 `Icon`을 구현한다. 제목 첫 필드의 `z` 또는 `c` 접두사를 분리한 뒤 정수 값을 검증하고 정규화한다. 숫자와 대시용 작은 비트맵 글리프를 대비가 높은 배경에 가운데 정렬하고 각 해상도의 32bpp DIB를 ICO에 담는다.
- [x] `go test ./internal/tray -run '^TestIcon' -count=1`을 실행해 통과를 확인한다.
- [x] 해당 두 파일을 `feat: Windows 트레이 숫자 아이콘 렌더링`으로 커밋한다.

### Task 2. Windows 트레이 연결 및 검증

**파일** `cmd/statusline/traycmd.go`, `docs/okf/how-to-guides/use-tray.md`, `ROADMAP.md` 수정. `cmd/statusline/traycmd_test.go` 추가.

**인터페이스** Task 1의 `tray.IconLabel(string) string`, `tray.Icon(string) []byte`를 사용한다. `func (a *trayApp) updateWindowsTray(title string, setTooltip func(string), setIcon func([]byte))`를 추가한다.

- [x] `TestUpdateWindowsTray`를 작성한다. `z87 w45`, `z87 w44`, `c87`, `c48`, `z87 w44`, `…`를 순서대로 전달하면 툴팁은 입력을 그대로 받고, 아이콘 콜백은 `87`, `48`, `87`, `–`의 ICO를 각각 한 번 받는지 검증한다.
- [x] `go test ./cmd/statusline -run '^TestUpdateWindowsTray$' -count=1`을 실행해 구현 누락으로 실패하는 것을 확인한다.
- [x] `updateWindowsTray`를 구현한다. 매 호출에서 전체 툴팁을 전달하고, 이전 아이콘 값과 비교해 변경됐을 때만 `setIcon`을 호출한다. `poll`의 기존 잠금 아래 Windows에서만 호출한다. Windows 이외의 OS는 기존 `SetTitle`과 `SetTooltip`을 사용한다.
- [x] `go test ./cmd/statusline -run '^TestUpdateWindowsTray$' -count=1`을 실행해 통과를 확인한다.
- [x] 변경한 Go 파일에 `gofmt`를 적용한다.
- [x] `go vet ./...`와 `go test ./...`을 실행한다.
- [x] `go build -o "$env:TEMP\statusline-tray-check.exe" ./cmd/statusline`을 Windows에서 실행한다.
- [x] Windows 데스크톱에서 임시 바이너리를 `tray`로 실행하여 확인 가능한 항목을 검증한다. 실측 범위는 아래 검증 기록을 참조한다.
- [x] 사용 가이드에 Windows 실행 명령, 표시 규칙, 숨겨진 아이콘 설정을 반영한다. 로드맵은 검증한 범위에 맞춰 갱신한다.
- [x] Task 2 변경과 이 구현 계획을 `feat: Windows 시스템 트레이 잔여량 표시`로 커밋한다.

## 검증 기록

- 아이콘 렌더링과 갱신 테스트는 구현 전 미정의 오류로 실패했고 구현 후 통과했다. 기본 Go 빌드 캐시에 쓰기 권한이 없어 검증 명령은 임시 `GOCACHE`를 사용했다.
- 격리된 테스트 캐시로 Windows에서 숫자 아이콘 파일 생성, Shell 등록, 네이티브 우클릭 메뉴, provider 설정 저장 및 종료를 확인했다. 재실행 후에도 설정이 유지되고 메뉴로 종료됐다.
- Windows 메뉴의 수동 갱신 명령으로 refresh marker가 생성되고, 연속 요청에서 재생성되지 않음을 확인했다. 실제 provider 호출은 자격 증명과 PATH를 격리하여 생략했다.
- 툴팁 문자열과 16×16 글리프의 픽셀 배치는 자동 테스트로 확인했다. 실제 작업 표시줄에서 숫자와 툴팁을 눈으로 읽는 검증은 수행하지 못했다.

## Provider 코드 표시 변경

초기 구현 검증 뒤 Windows 아이콘에 `z` 또는 `c`도 표시하도록 범위가 수정됐다. `z100`과 `c100`의 16×16 글리프 배치, `z–`와 `c–` 누락 상태, 동일한 잔여율에서 provider 변경 시 아이콘 갱신을 회귀 테스트로 확인했다. 주간 값만 바뀌면 툴팁이 갱신된다. Windows 프로세스에서 잔여율이 100%인 두 provider를 전환해 서로 다른 ICO가 생성되고 오류 없이 적용되는 것도 확인했다.
