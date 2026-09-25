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

- [x] Antigravity quota 파싱 (gemini/3p 카테고리 그룹핑)
- [x] 대상 프로바이더: **Gemini**, **Anthropic**, **OpenAI**, **Z.AI**
  - `gemini` — Gemini 모델 전용 quota
  - `3p` — Anthropic, OpenAI, Z.AI 등 서드파티 공유 quota
- [x] 잔여 비율별 색상 (초록 >50% / 노랑 20~50% / 빨강 <20%)
- [x] 출력: `📊 gemini 5h:93% wk:52% │ 3p 5h:100% wk:22%`

## 🔲 v0.3.0 — 미구현 (Config에 선언만 존재)

- [ ] `modelFormat` — 모델명 축약 (`short` / `full`)
- [ ] `wrapMode` — 터미널 폭 초과 시 truncate 처리
- [ ] `theme` — 테마 시스템 (`sleek_dark` 외 추가 테마)
- [ ] `permission` — 권한 요청 상태 표시 (HasPermission capability 존재)

## 🔲 v1.0.0 — GUI 위젯 지원

- [ ] 데스크톱 위젯 (macOS 메뉴바 / 시스템 트레이)
- [ ] 실시간 quota 대시보드 (Gemini, Anthropic, OpenAI, Z.AI)
- [ ] Context 사용량 시각화 (프로그레스 바, 차트)
- [ ] 알림 시스템 (quota 임계치 도달 시 알림)
- [ ] CLI → GUI 공유 데이터 레이어 (JSON 파이프 + 소켓/IPC)

## 💡 아이디어

- [ ] Quota reset 시간 표시 (`reset_in_seconds` 활용)
- [ ] Context 임계치 색상 (`thresholds.contextWarning/Critical` 활용)
- [ ] Codex 어댑터 보강 (현재 최소 구현)
- [ ] PPID 기반 auto-detection (설계 스펙에 명시, 미구현)
- [ ] `plan_tier` / `email` 표시 옵션
- [ ] 다중 테마 프리셋 (light, solarized 등)
- [ ] 웹 대시보드 (브라우저 기반 실시간 모니터링)
