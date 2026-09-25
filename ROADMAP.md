# ROADMAP

> statusline — AI 코딩 어시스턴트용 고성능 상태바 생성기

---

## ✅ v0.1.0 — 기반 구축 (완료)

- [x] Go 모듈 초기화 및 UnifiedStatus 모델
- [x] Config 시스템 (임베디드 기본값, `statusline init`)
- [x] CLI 어댑터 (Claude, Codex, Antigravity)
- [x] Auto-detection 판별기 (stdin JSON + 환경변수)
- [x] Git 메타데이터 직접 읽기 (repo, branch, dirty)
- [x] 10ms deadline Git status 체크
- [x] Lipgloss ANSI 렌더러
- [x] 레이아웃 시스템 (line1 + main)
- [x] Context 토큰 사용량 (bar/percentage)
- [x] Thinking/Agent 상태 표시
- [x] Active Skills / Last Tool 표시
- [x] OKF/Diátaxis 문서 체계

## ✅ v0.2.0 — Quota 표시 (완료)

- [x] Antigravity quota 파싱
  - `gemini` — Gemini 자체 모델 quota
  - `3rd` — Gemini에서 사용하는 외부 모델 quota
- [x] 잔여 비율별 색상 (초록 >50% / 노랑 20~50% / 빨강 <20%)
- [x] 출력: `📊 gemini 5h:93% wk:52% │ 3rd 5h:100% wk:22%`
- [ ] 각 프로바이더 자체 사용량 추적
  - Anthropic — Claude Code 사용량
  - OpenAI — Codex 사용량
  - Z.AI — Z.AI 사용량

## 🔲 v0.3.0 — 미구현 (Config에 선언만 존재)

- [ ] `modelFormat` — 모델명 축약 (`short` / `full`)
- [ ] `wrapMode` — 터미널 폭 초과 시 truncate 처리
- [ ] `theme` — 테마 시스템 (`sleek_dark` 외 추가 테마)
- [ ] `permission` — 권한 요청 상태 표시 (HasPermission capability 존재)

## 🔲 v0.8.0 — 다중 계정 (Multiple Accounts) 지원

- [ ] 페이로드 계정 메타데이터 추출 (`email`, `plan_tier`)
- [ ] 계정별 상태 캐시 레이어 (`~/.cache/statusline/accounts/<email>.json`)
- [ ] 다중 계정 Quota 집계 및 전환 감지
- [ ] CLI에서 현재 활성 계정 또는 위험(잔여 최소) 계정 Quota 우선 표시 옵션

## 🔲 v1.0.0 — 경량 트레이 위젯 (`fyne-io/systray`)

- [ ] **macOS** — 상단 메뉴바 위젯 (텍스트 `SetTitle` 및 아이콘 표시)
- [ ] **Windows** — 시스템 트레이 위젯 (툴팁 및 네이티브 팝업 메뉴)
- [ ] **다중 계정 드롭다운 메뉴** (계정별 Quota 분리 표시 및 활성 계정 상태 표시)
- [ ] 단일 Go 바이너리로 최소 리소스(초경량/저전력) 구동
- [ ] CLI ↔ Tray 간 공유 데이터 레이어 (IPC / 상태 파일 / 소켓 감시)

## 🔲 v2.0.0 — 리치 팝업 UI & 대시보드 (`Wails`)

- [ ] Wails 내장 트레이 기반 마이그레이션 (기존 백엔드 Go 비즈니스 로직 100% 재사용)
- [ ] 트레이 클릭 시 플로팅 팝업 창 (Webview 기반 미니 대시보드)
- [ ] 실시간 Quota 모니터링 카드 (Gemini, Anthropic, OpenAI, Z.AI)
- [ ] Context 사용량 시각화 (프로그레스 바, 실시간 차트 UI)
- [ ] OS 네이티브 알림 시스템 (Quota 고갈 위험 시 푸시 알림)

## 💡 아이디어

- [ ] Quota reset 시간 표시 (`reset_in_seconds` 활용)
- [ ] Context 임계치 색상 (`thresholds.contextWarning/Critical` 활용)
- [ ] Codex 어댑터 보강 (현재 최소 구현)
- [ ] PPID 기반 auto-detection (설계 스펙에 명시, 미구현)
- [ ] `plan_tier` / `email` 표시 옵션
- [ ] 다중 테마 프리셋 (light, solarized 등)
- [ ] 웹 대시보드 (브라우저 기반 실시간 모니터링)
