# Anthropic (Claude Code) OAuth Usage 수집 설계

> **Status**: Draft
> **Date**: 2026-09-27
> **Scope**: `--provider=anthropic` 수집 파이프라인, CC payload `cost.total_cost_usd` 파싱, 상태바 quota 세그먼트

## 목차

- [배경](#배경)
- [목표 및 비목표](#목표-및-비목표)
- [API 계약](#api-계약)
- [설계](#설계)
- [오류 계약](#오류-계약)
- [테스트 전략](#테스트-전략)
- [리스크 및 트레이드오프](#리스크-및-트레이드오프)
- [향후 작업](#향후-작업)

## 배경

ROADMAP v0.6.0. Claude Code(claude 엔진)는 v0.5.1까지 토큰·컨텍스트 파싱만 지원하며 quota 수집 경로가 없다. Anthropic OAuth usage API(`api.anthropic.com/api/oauth/usage`)가 5h/7d 창·모델별 주간 quota·extra 사용량·구독 플랜을 제공하므로, v0.4.0(Codex)·v0.5.0(Z.AI)이 확립한 collect→캐시→어태치 파이프라인에 동일 패턴으로 편입한다.

## 목표 및 비목표

**목표**

1. `statusline collect --provider=anthropic` 수집 — OAuth 자격 증명(Keychain 우선, file/env 폴백)
2. 기본 병합(`all`)에 anthropic 포함 — 무자격 시 조용 스킵(zai 관행 대칭)
3. 상태바 기본 세그먼트: 5h/weekly 잔여율 + 세션 누적 비용(`cost.total_cost_usd`)
4. 플랜 인식(`subscriptionType`), 모델별 quota(sonnet/opus/미인식 패밀리), extra 사용량 — config opt-in
5. 만료 토큰 refresh + write-back(기존 필드 병합) — OMC 풀 패리티

**비목표**

1. 403 setup-token 폴백(`/v1/messages` ratelimit-unified 헤더) — YAGNI. 정식 OAuth 로그인 사용자는 해당 없음
2. 연속 폴링 백오프 상태기계(OMC 지수 백오프·per-identity backoffs) — 수동 1회성 collect에 불필요. v0.8.0 데몬 시점 재검토
3. 세션/일일 단위 사용량 이력 — API가 5h/7d 창만 제공
4. 다중 계정 — v0.7.0 범위
5. gemini 수집 — [collect-all 스펙](./2026-09-26-collect-all-providers-design.md) 결정 유지

## API 계약

2026-09-27 기준 OMC HUD `usage-api.ts`(claude-hud 기반) 분석. **Unverified — 구현 전 라이브 실측으로 확정한다.**

| 항목 | 값 |
| :--- | :--- |
| Endpoint | `GET https://api.anthropic.com/api/oauth/usage` |
| 인증 | `Authorization: Bearer <accessToken>` + `anthropic-beta: oauth-2025-04-20` |
| User-Agent | `claude-code/<semver>` 필수에 가까움 — bare/누락 시 시간당 ~1회 버킷으로 429 throttle(실측 retry-after 348s). 임의 버전 지어쓰기 금지, 실측값만 |
| 응답 | `five_hour{utilization, resets_at}`, `seven_day{…}`, `seven_day_sonnet`, `seven_day_opus`, `extra_usage{spent_usd, limit_usd, used_credits, monthly_limit, currency, decimal_places}`, `limits[]`(kind=weekly_scoped, scope.model.display_name) — utilization은 0~100 |
| Refresh | `POST https://platform.claude.com/v1/oauth/token` — `grant_type=refresh_token`, public client_id(기본 상수 + `CLAUDE_CODE_OAUTH_CLIENT_ID` env override) |
| 자격 증명 구조 | `{ claudeAiOauth: { accessToken, expiresAt, refreshToken, subscriptionType, rateLimitTier } }` — flat 변형 존재 |

## 설계

### 자격 증명 모듈 (`internal/collect/anthropic.go`)

로드 순서(OMC `getCredentials` port):

1. `CLAUDE_CODE_OAUTH_TOKEN` env — setup-token. 무기한 취급, refresh/write-back 없음
2. macOS Keychain: `/usr/bin/security find-generic-password -s <service> [-a <username>] -w`(timeout 2s, acct=username 우선 → service-only). `CLAUDE_CONFIG_DIR` 설정 시 service = `Claude Code-credentials-<sha256(env 문자열)[:8]>`, 미설정 시 `Claude Code-credentials`
3. `$CLAUDE_CONFIG_DIR` 또는 `~/.claude/.credentials.json` 폴백

**만료 대응**: `expiresAt <= now`면 refresh 시도. 성공 시 write-back — Keychain은 기존 JSON을 읽어 `claudeAiOauth` 토큰 필드만 병합 후 `add-generic-password -U`, file은 tmp+rename(0600). env 소스는 미기록. refresh 실패 시 `errors["anthropic"]="auth"`(fail-soft).

**보안**: `api.anthropic.com`·`platform.claude.com`은 소스에 고정 — env로 호스트 주입 불가(z.ai와 달리 SSRF 가드 불필요). 토큰은 stdout·로그·캐시 미기록.

### UA 해결

1. 캐시(`anthropic.json`)의 `userAgentVersion`
2. `claude --version` subprocess(timeout 2s, semver 추출)
3. 둘 다 없으면 UA 생략 → throttled bucket 429 → `errors` 기록(fail-soft)

UA 검증은 OMC `buildUserAgent` port: 앵커드 `/^\d+\.\d+\.\d+[A-Za-z0-9.+-]*$/` + 128자 제한. 통과 시에만 `claude-code/<ver>` 헤더 전송. version은 **수집 성공 시 캐시 파일에 함께 기록**(획득과 병행해 성공 캐시에만 반영 — 실패 시 재획득).

### 수집기 (`RunAnthropicCollect`)

- `GET api.anthropic.com/api/oauth/usage` — Bearer + beta 헤더 + UA(선택). runCollect 공유 ctx(10s), `http.Client` timeout 5s(platform/zai 고루틴 관행 대칭)
- 응답 분기: 200 → 파싱·캐시 갱신 / 429 → `errors` 기록, 캐시 stale 유지 / 403 → `errors["anthropic"]="forbidden (setup-token)"` / 기타·타임아웃 → `errors` 기록
- **호스트 고정 원칙**: `ANTHROPIC_BASE_URL`과 무관하게 항상 `api.anthropic.com` — env가 호스트를 결정하지 않는다. CC를 z.ai로 구동해도 Anthropic OAuth 잔여는 독립 수집된다(계정 축 특성)

`AnthropicSnapshot`(출력 JSON `anthropic` 키):

```go
type AnthropicSnapshot struct {
    FiveHour      AnthropicBucket   // {Percent, ResetsAt}
    SevenDay      AnthropicBucket
    SonnetWeekly  *AnthropicBucket  // flat seven_day_sonnet 우선
    OpusWeekly    *AnthropicBucket  // flat seven_day_opus 우선
    ScopedBuckets []AnthropicScoped // 미인식 패밀리(Fable 등) — display_name, percent, resetsAt, isActive
    ExtraUsage    *AnthropicExtra   // 3경로: enterprise credit / org overage / Pro metered
    Plan          string            // subscriptionType
    RateLimitTier string
}
```

파싱 규칙(OMC `parseUsageResponse` port):

- utilization clamp 0~100
- sonnet/opus: flat 필드가 신뢰 데이터 — 존재하면 `limits[]` 폴백 미갱신, 부재 시에만 `kind=weekly_scoped`에서 display_name substring(sonnet/opus) 매칭으로 갭 필
- 미인식 패밀리는 `ScopedBuckets` generic 항목 — 신규 티어 무배포 대응
- `extra_usage` 3경로: enterprise(`used_credits`+`monthly_limit`, `currency`/`decimal_places` 스케일 해석) / Max·Pro org overage(credit형) / Pro metered(`spent_usd`/`limit_usd`)
- 유효값 0개면 null → `errors`

### 캐시·어태치 (`anthropic_cache.go` / `anthropic_attach.go`)

- 캐시: `os.UserCacheDir()/statusline/anthropic.json`(`zai.json` 대칭 — macOS `~/Library/Caches/statusline/`). 구조 `{ fetchedAt, snapshot, userAgentVersion }`. 쓰기는 collect만(tmp+rename), 읽기는 attach만
- `AttachAnthropicQuota(st, ttl)`: 게이트 `st.EngineName == "claude" && st.Provider == ""`(zai/zhipu 감지 시 Z.AI quota가 우선 — 상호배타). 캐시 mtime TTL 5분. stale 시 `TryMarkAnthropicRefreshAt`(marker 60s) + `SpawnSelfCollect("anthropic")` 백그라운드 갱신 — Z.AI 패턴 완전 대칭
- `st.Quota` 주입 — `QuotaCategory` 구조 변경 없음(`Weekly` 필드 재사용):
  - `anth` 카테고리: FiveH + Weekly(잔여율 변환 `(100 − utilization) / 100` — 기존 잔여율 색상 계약과 대칭)
  - `sonnet`/`opus`/scoped 패밀리: 별도 카테고리(Weekly만) — 어태치는 전량 주입, 표시 선택은 렌더 config
- `UnifiedStatus` 확장 3필드: `CostUsd float64`, `PlanTier string`, `ExtraUsage *ExtraUsageInfo{SpentUsd, LimitUsd float64; ResetsAt int64}`

### 어댑터 (`internal/adapter/claude.go`)

raw struct에 `cost.total_cost_usd` 파싱 추가 → `st.CostUsd`. 0이면 렌더 미표시.

### 렌더러 (`internal/render/`)

- 기본: `anth 5h:38% wk:59%`(잔여율, 기존 임계색 초록 >50% / 노랑 20~50% / 빨강 <20%) + `$0.42`(CostUsd, Context 세그먼트 인접)
- opt-in config(`configuration-schema.md` 갱신):
  - `anthropic.showModelWeekly`(default false) — `sonnet wk:70% opus wk:88%`(+scoped 패밀리)
  - `anthropic.showExtraUsage`(false) — `$2.3/$10`
  - `anthropic.showPlan`(false) — `anth[pro]` 배지

```text
기본:   🧠 Sonnet 4.6 │ ⚡ 57k/1M(6%) $0.42 │ anth 5h:38% wk:59% │ main
풀옵션: 🧠 Sonnet 4.6 │ anth[pro] 5h:38% wk:59% sonnet wk:70% opus wk:88% $2.3/$10 $0.42 │ main
```

### CLI (`cmd/statusline/collectcmd.go`)

- `selectProviders`: `""`/`all` → `{chatgpt, zai, anthropic}`, `anthropic` → `{anthropic}`, `anthropic`을 not-yet 분기에서 제거(gemini만 유지)
- `collectOutput`에 `Anthropic *collect.AnthropicSnapshot` 필드 추가
- runCollect 고루틴 1개 추가 — 결과 병합은 기존 메인 고루틴 단일 병합 구조 유지

### 구현 범위

| 파일 | 변경 |
| :--- | :--- |
| `internal/collect/anthropic.go` | 신규 — 자격 증명·refresh·write-back·UA·수집·파싱 |
| `internal/collect/anthropic_cache.go` | 신규 — 캐시·refresh marker |
| `internal/collect/anthropic_attach.go` | 신규 — 렌더 브리지 |
| `internal/adapter/claude.go` | `cost.total_cost_usd` 파싱 |
| `internal/model/status.go` | `CostUsd`, `PlanTier`, `ExtraUsageInfo` 필드 |
| `internal/render/` | quota 세그먼트·config 옵션 |
| `cmd/statusline/main.go`, `collectcmd.go` | 배선 |
| `docs/okf/reference/configuration-schema.md` | 옵션 문서화 |
| `docs/samples/collect/` | 실측 익명화 샘플 |

## 오류 계약

| 상황 | stdout | stderr | exit |
| :--- | :--- | :--- | :--- |
| `all` — anthropic 무자격 | `anthropic` 키 생략 | 없음 | 0 |
| `all` — 조회 실패(429/403/네트워크) | `errors.anthropic` 기록 | 없음 | 0 |
| `--provider=anthropic` 단독 — 무자격 | 없음 | 오류 메시지 | 1 |
| `--provider=anthropic` 단독 — 조회 실패 | `{"errors":{"anthropic":…}}` | 없음 | 0 |
| 토큰 만료 + refresh 실패 | `errors.anthropic="auth"` | 없음 | 0 |
| 어태치 — TTL 초과/캐시 부재 | 세그먼트 생략(렌더) | 없음 | - |

## 테스트 전략

TDD. `go test -race ./...` 포함 전체 통과가 완료 조건.

| 대상 | 검증 항목 |
| :--- | :--- |
| 자격 증명 | 로드 우선순위, `claudeAiOauth`/flat 파싱, 만료 경계, refresh 성공/실패(httptest), write-back 미지 필드 보존·file 0600 |
| UA | 앵커드 정규식(유효 semver/가짜 거부/128자), 캐시 히트→subprocess 스킵, exec stub, 생략 분기 |
| 파싱 | clamp, flat 우선→`limits[]` 폴백(존재 시 미갱신), scoped generic, extra 3경로+currency/decimal_places, 전무 효율→null |
| 수집기 | httptest 200/429/403/timeout |
| collect 통합 | 단독 무자격 exit 1, `all` 무자격 조용 스킵 exit 0, 3-provider 병렬 부분 실패 errors |
| 어태치 | 게이트(claude+zai 상호배타), TTL, 잔여율 변환(utilization 62 → 0.38), spawn marker |
| 렌더러 | 기본/opt-in 스냅샷, 임계색 매핑 |
| 어댑터 | `cost.total_cost_usd` 있음/없음/0 |

**실측 검증(구현 후, 사용자 승인 하에)**: 라이브 `collect --provider=anthropic`(Keychain 비밀값 읽기 발생 — 명시적 승인 필요), 익명화 샘플 저장, `claude --version` 출력 형태 확인.

## 리스크 및 트레이드오프

| 항목 | 내용 | 대응 |
| :--- | :--- | :--- |
| 비공개 API 계약 | 문서화되지 않은 엔드포인트 — 스키마 변경 가능 | OMC/claude-hud 생태계 참조 유지, 파싱은 방어적(clamp·스킵·미인식 패밀리 generic) |
| Keychain 비대칭형 읽기 | `security -w`가 첫 회 승인 다이얼로그를 띄울 수 있음 | 구현 후 실측. 거부 시 안내 메시지 |
| `claude --version` 출력 형태 가정 | Unverified — 형식이 다르면 파싱 실패 | UA 생략 분기가 안전한 폴백(429 기록), 실측으로 형태 확정 |
| CC 소유 크레덴셜 write-back | 타 도구의 자격 증명을 갱신·재기록 | OMC 검증 경로 port, 미지 필드 병합 보존, env 소스 미기록 |
| UA 부재 시 1/hour throttle | 첫 수집이 429로 실패 가능 | `claude --version` 우선 실행으로 사실상 회피. 429는 errors 기록 후 재시도 |

## 향후 작업

1. 403 setup-token 폴백(`/v1/messages` ratelimit-unified 헤더) — setup-token 사용자 발생 시
2. v0.8.0 데몬 도입 시 연속 폴링 백오프 상태기계 이식
3. 다중 계정 라벨링 — v0.7.0
