# Z.AI Quota 지원 설계 (v0.5.0)

> **Source**: 설계 브레인스토밍 세션 (2026-09-26) — `ROADMAP.md` v0.5.0 항목 구체화

## 목차

- [배경](#배경)
- [목표 및 비목표](#목표-및-비목표)
- [검증된 사실](#검증된-사실)
- [결정된 접근법](#결정된-접근법)
- [상세 설계](#상세-설계)
- [테스트 전략](#테스트-전략)
- [Out of Scope](#out-of-scope)

## 배경

Z.AI Coding Plan은 자체 CLI 호스트가 아닌 **Claude Code의 백엔드**로 동작한다 (`ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic`). 따라서 로드맵의 "Z.AI CLI 전용 어댑터"는 실제로는 두 번째 문제다:

1. Claude Code 엔진 흐름 안에서 Z.AI provider 감지
2. stdin payload에 존재하지 않는 quota 데이터의 수집→캐시→렌더 브리지

렌더 경로는 <5ms 예산 때문에 API 호출이 불가하므로, quota 표시에는 캐시 계층이 필수다. 이 캐시 계층은 v0.7.0(다중 계정)·v0.8.0(통합 수집기)의 기반이 된다.

## 목표 및 비목표

**목표**

- Z.AI 백엔드에서 구동 중임을 자동 감지하고 모델명(glm-5.3)과 함께 quota를 상태바에 표시
- 5시간 토큰 잔여율 + 월간 MCP 사용 잔여율 실시간 표시
- 별도 설정 없이 자동 갱신 (out-of-box)

**비목표**

- 다중 계정 (v0.7.0)
- model-usage / tool-usage 시간별 상세 수집 (활용처 미정, YAGNI)
- 백그라운드 데몬 (프로젝트 원칙 위반)

## 검증된 사실

zai-coding-plugins 0.0.1 소스 + OMC HUD `usage-api.js` (claude-hud 기반) 분석 + 실계정 실측 (2026-09-26, ZAI/Pro):

- 인증: env `ANTHROPIC_AUTH_TOKEN`을 `Authorization` 헤더에 raw로 전달 (Bearer 없음)
- 플랫폼 판별: `ANTHROPIC_BASE_URL` 도메인 — `api.z.ai` → ZAI, `open.bigmodel.cn`/`dev.bigmodel.cn` → ZHIPU (동일 API 형태)
- `GET {base}/api/monitor/usage/quota/limit` (파라미터 없음):
  - 응답 envelope: `{data: {limits: [...]}}` (또는 최상단 `limits`)
  - `limits[].type == "TOKENS_LIMIT"` — 토큰 버킷. **여러 개일 수 있음** (`unit` 코드로 분류, `unit=6`이 주간 — OMC 주석: observed, undocumented). 5시간 버킷은 `unit!=6` 중 `nextResetTime` 가장 빠른 것
  - `limits[].type == "TIME_LIMIT"` — 월간 MCP, `percentage`(사용률), `currentValue`, `usage`(총한도), `usageDetails`(도구별)
  - 각 항목에 `nextResetTime` (Unix 밀리초) — reset 카운트다운 소스
  - 실측 예: MCP 34% 사용 (341/1000, search-prime 303 / web-reader 33 / zread 5), 5h 1% 사용
- OMC HUD 검증 캐시 전략 (운영 실적 있음): 실패 15초 / 네트워크 오류 2분 / 429 백오프 5분 / 최대 15분 경과 데이터 폐기
- Antigravity 렌더러는 **잔여율** 컨벤션(`5h:93%`)이므로 `잔여 = 1 − percentage` 변환 필요. OMC HUD는 사용률 컨벤션 — 같은 슬롯의 숫자가 의미가 반대라는 점 유의

## 결정된 접근법

갱신 전략 3안(자가 갱신 / 수동·훅 / watch 롱폴링) 비교 후 **스테일-리드 + 자가 갱신** 확정:

- 렌더 경로는 캐시 파일만 읽고 즉시 렌더 (~0.1ms)
- 캐시 mtime이 TTL(5분) 경과 시 detached `collect` 프로세스를 비동기 spawn — 렌더 차단 없음
- one-shot 프로세스이므로 데몬이 아님 (프로젝트 원칙 준수)

## 상세 설계

### 감지 (Detection)

- `UnifiedStatus`에 `Provider string` 필드 추가 (`""` / `"zai"` / `"zhipu"`). engine은 계속 `claude`
- 신규 `internal/adapter/zai.go`: `ProviderFromEnv(env, modelName)` 헬퍼
  - 1순위: `ANTHROPIC_BASE_URL` 도메인 판별
  - 2순위(폴백): 모델명 `glm-` 접두사
- `ParseInput`이 claude 엔진 확정 후 provider 주입

### 수집 (`collect --provider=zai`)

- GET `{base}/api/monitor/usage/quota/limit` 1회. base는 `ANTHROPIC_BASE_URL`에서 `protocol://host` 추출
- 매핑:
  - `TOKENS_LIMIT`(`unit != 6` 중 `nextResetTime` 가장 빠른 항목) → `QuotaCategory.FiveH = 1 − percentage/100`, `FiveHResetsAt = nextResetTime` (Unix ms)
  - `TIME_LIMIT` → 신규 `Monthly = 1 − percentage/100`, 신규 `MonthlyResetsAt = nextResetTime`
  - 주간 토큰 버킷(`unit=6`)은 파싱만 기록하고 표시하지 않는다 (해당 플랜 존재 여부가 가변적 — YAGNI)
- SSRF 가드: `ANTHROPIC_BASE_URL` 파싱 후 hostname이 알려진 provider 도메인과 정확히 일치할 때만 호출
- 헤더: `Authorization`(raw token), `Accept-Language: en-US,en`

### 캐시 브리지

- 경로: `~/.cache/statusline/zai.json` (XDG). 형식: `{"fetched_at": <RFC3339>, "limits": [...]}`
- 시크릿 미포함 — 토큰은 캐시에 저장하지 않는다. atomic write (temp+rename)
- 렌더: 캐시 stat+read → TTL 5분 경과 && `ANTHROPIC_AUTH_TOKEN` 존재 → detached spawn (`Setsid` 등 플랫폼별 분리), stdout 버림
- 캐시 부재/파손/token 미설정 → quota 미표시 (fail-soft)

### 렌더링

- OMC HUD 동등 표시 (reset 카운트다운 포함): `📊 zai 5h:97%(4h46m) mo:66%(18d19h)` — 라벨은 `Provider` 값 사용 (`zai`/`zhipu`)
- 카운트다운 포맷: `>1일 → "{d}d{h}h"`, `1시간~1일 → "{h}h{m}m"`, 경과 시 생략 (OMC 규칙 동일)
- 숫자는 **잔여율** 컨벤션 유지 (Antigravity quota와 동일 — OMC HUD의 사용률 컨벤션과 반대)
- 색상 임계치 동일: 잔여 >50% 초록, 20~50% 노랑, <20% 빨강
- `Provider != ""` → `HasQuota = true`

### 에러 핸들링

- API 401/403/5xx → collect는 오류 JSON 출력(기존 패턴), 캐시는 갱신하지 않고 이전 값 유지
- 렌더 경로의 모든 실패는 quota 미표시로 흡수 — stdout 오염 없음 (fail-soft 원칙)

## 테스트 전략

- 감지: env 조합 매트릭스 + `glm-` 폴백 유닛테스트
- 수집: `httptest` 서버로 quota/limit 목업 → limits 매핑·잔여 변환 검증
- 캐시: TTL 판정(경과/미경과), atomic write, 파손 파일 fail-soft
- 렌더: golden test (`📊 zai 5h:97%(4h46m) mo:66%(18d19h)` 및 색상 임계치, 카운트다운 포맷 경계 — 1일 기준/경과 시 생략)
- 통합: 캐시 파일 준비 + `cat payload | ./statusline --cli=claude` → quota 라인 포함 확인

## 참조 소스

| 소스 | 참조 내용 | 적용 |
| :--- | :--- | :--- |
| zai-coding-plugins `query-usage.mjs` (Z.AI 공식) | 엔드포인트·인증 헤더·응답 envelope | v0.5.0 수집 |
| OMC HUD `usage-api.js` (claude-hud 기반) | `nextResetTime`/`unit` 스키마, 버킷 분류, 캐시 백오프 전략, SSRF 가드, 자격증명 origin 고정 | v0.5.0 수집·캐시 |
| OMC HUD `usage-api.js` — 3rd-party dispatch | `ANTHROPIC_BASE_URL` hostname 기반 z.ai/ZHIPU/MiniMax/Kimi 분기 패턴 | v0.5.0 감지 + 향후 provider 확장 |
| OMC HUD `usage-api.js` — Anthropic OAuth | `api.anthropic.com/api/oauth/usage`, Keychain 자격 증명, 토큰 리프레시 | v0.6.0 Anthropic 지원 |
| 실계정 실측 (2026-09-26) | MCP 34%(341/1000), 5h 1% | 스키마·테스트 픽스처 |

## Out of Scope

- `model-usage` / `tool-usage` 엔드포인트 — v1.0+/v2.0.0 대시보드에서 상세 사용량 리포트로 활용 (`glm-plan-usage:usage-query` 스타일 참조)
- 주간 토큰 버킷(`unit=6`) 표시 — 파싱만 하고 표시하지 않음 (플랜별 존재 가변)
- 다중 계정 캐시 분리 (`~/.cache/statusline/accounts/...`) — v0.7.0
- `statusline watch` 확장
