# ROADMAP

> statusline — AI 코딩 어시스턴트용 고성능 상태바 & 위젯 생성기

---

### 🎯 핵심 로드맵 전략

1. **Phase 1 (~ v0.7.0)**: **터미널 CLI 상태바(statusline) 완성**
   - Claude Code, Codex, Antigravity, Z.AI 환경에서 단일행 고성능 스트림 필터로 완벽 동작
   - 각 프로바이더별 Quota, 토큰, 사용량, 구독 플랜 및 다중 계정 파싱 로직 안정화
2. **Phase 2 (v0.8.0 ~ v0.9.0)**: **GUI 에이전트 연동 & OS 시스템 트레이 위젯 (`fyne-io/systray`)**
   - VS Code, Antigravity IDE, Codex/Claude Work 등 GUI 에이전트 사용량 측정 및 공통 캐시 적재
   - CLI/GUI 공통 데이터를 macOS 메뉴바/Windows 트레이에서 실시간 표출
3. **Phase 3 (v1.0.0)**: **공식 정식 릴리스 & 상용화 (Commercial Launch)**
   - Community(무료): CLI 상태바 / Pro(유료): GUI 위젯 + GUI 에이전트 모니터링 + 다중 계정
4. **Phase 4 (v2.0.0 ~)**: **리치 플로팅 대시보드 (`Wails`)**
   - 트레이 클릭 시 Webview 팝업 창을 통한 차트 및 상세 카드 제공

---

## ✅ v0.1.0 — 기반 구축 (완료)

- [X] Go 모듈 초기화 및 UnifiedStatus 모델
- [X] Config 시스템 (임베디드 기본값, `statusline init`)
- [X] CLI 어댑터 (Claude, Codex, Antigravity)
- [X] Auto-detection 판별기 (stdin JSON + 환경변수)
- [X] Git 메타데이터 직접 읽기 (repo, branch, dirty)
- [X] 10ms deadline Git status 체크
- [X] Lipgloss ANSI 렌더러
- [X] 레이아웃 시스템 (line1 + main)
- [X] Context 토큰 사용량 (bar/percentage)
- [X] Thinking/Agent 상태 표시
- [X] Active Skills / Last Tool 표시
- [X] OKF/Diátaxis 문서 체계

## ✅ v0.2.0 — Antigravity Quota 표시 (완료)

- [X] Antigravity quota 파싱
  - `gemini` — Gemini 자체 모델 quota
  - `3rd` — Gemini에서 사용하는 외부 모델 quota
- [X] 잔여 비율별 색상 (초록 >50% / 노랑 20~50% / 빨강 <20%)
- [X] 출력: `📊 gemini 5h:93% wk:52% │ 3rd 5h:100% wk:22%`
- [X] Claude Code 기본 어댑터 연동 (ContextTokens, Thinking, Skills)

## ✅ v0.3.0 — CLI 세부 옵션 고도화 (완료)

- [X] `modelFormat` — 모델명 축약 (`short` / `full`)
- [X] `wrapMode` — 터미널 폭 초과 시 truncate(`…`) 처리 (단일행 유지 보장)
- [X] `theme` — 테마 시스템 (`sleek_dark`, `light`, `nord`)
- [X] `permission` — 권한 요청 상태 표시 (`🔒 WAITING` 배지)

## ✅ v0.4.0 — OpenAI (Codex / Platform) 완전 지원 (완료)

- [X] OpenAI Codex 어댑터 고도화 및 컨텍스트/토큰 파싱
- [X] 구독 플랜 인식 및 표시 (ChatGPT Plus / Team / Enterprise)
- [X] OpenAI Platform Admin API 연동 및 누적 비용(`total_cost`) 실시간 추적
- [X] Rate limits 및 잔여 쿼터 실시간 표시
- [X] Codex / OpenAI 읽기 전용 능동 수집기 및 스트림 감시 구현 (`statusline collect`, `statusline watch`)

## ✅ v0.5.0 — Z.AI 완전 지원 (완료)

- [X] Z.AI provider 감지 어댑터 (`internal/adapter/zai.go`) — `ANTHROPIC_BASE_URL` 도메인 + `glm-` 모델 폴백
- [X] Auto-detection 판별기에 Z.AI provider 감지 추가 (claude 엔진 내 provider 주입)
- [X] Z.AI quota 수집 (`collect --provider=zai`) 및 캐시 브리지 — 5h 토큰/월간 MCP 잔여율 + reset 카운트다운 상태바 표시

## 🔲 v0.6.0 — Anthropic (Claude Code) 완전 지원

- [ ] Anthropic OAuth usage API 연동 (`api.anthropic.com/api/oauth/usage`, Keychain 자격 증명) — OMC HUD `usage-api.js`·claude-hud 참조
- [ ] Claude Code 페이로드의 토큰 및 누적 비용(`total_cost`) 모니터링 연동
- [ ] 세션 및 일일/월간 Quota 잔여량 파싱 및 표시
- [ ] 구독 플랜 인식 (Claude Pro / Team / Enterprise)
- [ ] Anthropic 모델별(Opus, Sonnet, Haiku) 사용량 세분화

## 🔲 v0.7.0 — 다중 계정(Multi-Accounts) & 공통 예산·캐시 인프라

- [ ] 동일 프로바이더 다중 계정(회사용/개인용) 식별 및 라벨링
- [ ] 계정별 독립 캐시 레이어 (`~/.cache/statusline/accounts/<provider>/<profile>.json`)
- [ ] 일일/월간 예산 한도(Budget Limit) 설정 및 임계치 도달 경고 표시
- [ ] 비용 표시 포맷 옵션 (통화 단위 `USD`/`KRW`, 소수점 자릿수 등)
- [ ] CLI 상태바에서 활성 계정 또는 위험(잔여 최소) 계정 선택 노출 옵션
  - 예: `oai[work] 5h:80% │ oai[pers] 5h:20% ⚠️`

## 🔲 v0.8.0 — 통합 수집기 & GUI 에이전트 모니터링 (Local Watcher & Daemon) 🌟

- [ ] **통합 `collect` 서브커맨드 확장**: 현재 Codex/Platform 전용인 `statusline collect`를 Antigravity, Claude, Codex, Platform 등 전 프로바이더 통합 수집(`--provider=all|antigravity|claude|codex`)으로 확장
- [ ] **Antigravity IDE** 로컬 세션 및 브레인 로그(`~/.gemini/antigravity/...`) 감시 어댑터
- [ ] **Codex Work** 로컬 app-server 세션 및 작업 공간(Workspace) 쿼터/사용량 파싱
- [ ] **Claude Work** (Enterprise/Desktop) 로컬 세션 및 사용량 이벤트 감시
- [ ] **VS Code** 확장 프로그램 작업 공간(`workspaceStorage`) 및 AI 확장 상태 감시
- [ ] GUI/CLI 수집 데이터를 공통 캐시(`~/.cache/statusline/live_session.json`)로 표준화 적재
- [ ] 백그라운드 수집 데몬 프로세스 지원 (`statusline daemon` 또는 `statusline collect --watch-gui`)

## 🔲 v0.9.0 — GUI 트레이 위젯 최초 도입 (`fyne-io/systray`)

- [ ] **macOS** — 상단 메뉴바 위젯 (텍스트 `SetTitle` 및 아이콘 표시)
- [ ] **Windows** — 시스템 트레이 위젯 (툴팁 및 네이티브 팝업 메뉴)
- [ ] 단일 Go 바이너리로 최소 리소스(초경량/저전력) 구동
- [ ] CLI ↔ Tray 간 공유 데이터 레이어 (IPC / 상태 파일 / 소켓 감시)
- [ ] 다중 계정 및 CLI/GUI 에이전트 통합 드롭다운 메뉴 (계정별 Quota 및 활성 상태 분리 표시)

## 🔲 v1.0.0 — 공식 정식 릴리스 & 상용화 (Commercial Launch) 💵

- [ ] **공식 첫 정식 릴리스 (General Availability)**
- [ ] **인증 및 자격 증명 보안 저장**:
  - `~/.config/statusline/credentials.json` (권한 `0600` 강제, 초고속 경량 파싱)
  - 다중 계정 API 키 및 라이선스 키 보관
- [ ] **상용 라이선스 & 유료화 모델 (Monetization)**:
  - CLI 기본 단일행 상태바: 무료 (Community)
  - GUI 트레이 위젯 + 다중 계정 + GUI 에이전트(VS Code/Antigravity/Codex/Claude Work) 통합 모니터링 + 알림: 유료 (Pro License)
- [ ] **라이선스 관리 및 검증 시스템**:
  - 라이선스 키 활성화 커맨드 (`statusline activate <license-key>`)
  - 오프라인/온라인 라이선스 검증 레이어 (Lemon Squeezy / Polar / Gumroad 연동)
- [ ] **릴리스 패키징 및 배포**:
  - macOS 서명(Notarization) & Windows 코드 사이닝
  - 공식 웹사이트/결제 페이지 오픈 및 배포 파이프라인 (GoReleaser)

## 🔲 v2.0.0 — 리치 팝업 UI & 대시보드 (`Wails`)

- [ ] Wails 내장 트레이 기반 마이그레이션 (기존 백엔드 Go 비즈니스 로직 100% 재사용)
- [ ] **OS 네이티브 자격 증명(키체인) 마이그레이션**:
  - macOS Keychain / Windows Credential Manager (`zalando/go-keyring`) 연동
- [ ] 트레이 클릭 시 플로팅 팝업 창 (Webview 기반 미니 대시보드)
- [ ] 실시간 Quota 모니터링 카드 (Gemini, Anthropic, OpenAI, Z.AI)
- [ ] 상세 사용량 리포트 (모델별/시간별/도구별 — Z.AI `model-usage`·`tool-usage` API, `glm-plan-usage:usage-query` 참조)
- [ ] Context 사용량 시각화 (프로그레스 바, 실시간 차트 UI)
- [ ] OS 네이티브 알림 시스템 (Quota 고갈 위험 시 푸시 알림)

## 💡 아이디어

- [ ] 제3자 provider 어댑터 확장 (MiniMax, Kimi/Moonshot) — Z.AI 어댑터 dispatch 패턴 재사용, OMC HUD `usage-api.js` 참조

- [ ] Quota reset 시간 표시 (`reset_in_seconds` 활용)
- [ ] Context 임계치 색상 (`thresholds.contextWarning/Critical` 활용)
- [ ] Codex 어댑터 보강 (현재 최소 구현)
- [ ] PPID 기반 auto-detection (설계 스펙에 명시, 미구현)
- [ ] `plan_tier` / `email` 표시 옵션
- [ ] 다중 테마 프리셋 (light, solarized 등)
- [ ] 웹 대시보드 (브라우저 기반 실시간 모니터링)
