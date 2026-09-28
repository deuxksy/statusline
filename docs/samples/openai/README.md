# OpenAI 실제 조회 스냅샷

`statusline collect`가 실제 Codex app-server와 OpenAI Platform API에서 반환받은 값을 날짜별로 보관한다. 날짜별 폴더는 **실제 응답**이며, [`schema-examples/`](./schema-examples/README.md)는 스키마 기반 합성 예시다. `docs/samples/codex.json`은 로컬 JSONL에서 정규화한 렌더러 fixture로 또 다른 형식이다.

## 2026-09-26 스냅샷

- 수집 시각: 2026-09-26 01:22:28 KST (`2026-09-25T16:22:28Z`)
- 명령: `./statusline collect`
- 출처: Codex 로컬 app-server의 `account/read`, `account/rateLimits/read`, `account/usage/read`, `thread/list`; OpenAI Platform의 Completions Usage·Costs GET
- 범위: Platform Usage·Costs는 수집 시각 기준 직전 1일부터 조회했다.
- 익명화: 이메일, 계정·workspace·thread·credit ID, 로컬 작업 경로, Platform 사용자·프로젝트·API 키 ID 및 line item 값을 `<redacted>`로 치환했다. 수치, `null`, reset 시각, plan type, 모델명, 응답 구조는 유지했다. 인증 키는 수집 결과에 포함되지 않는다.

| 파일 | 출처 |
| --- | --- |
| [codex-account.json](./2026-09-26/codex-account.json) | `account/read` |
| [codex-rate-limits.json](./2026-09-26/codex-rate-limits.json) | `account/rateLimits/read` |
| [codex-account-usage.json](./2026-09-26/codex-account-usage.json) | 계정 `account/usage/read` |
| [codex-latest-thread.json](./2026-09-26/codex-latest-thread.json) | 최근 `thread/list` 항목에서 본문·preview를 제외한 메타데이터 |
| [codex-latest-thread-usage.json](./2026-09-26/codex-latest-thread-usage.json) | 최근 thread ID로 조회한 `account/usage/read` |
| [platform-completions-usage.json](./2026-09-26/platform-completions-usage.json) | `GET /v1/organization/usage/completions` |
| [platform-costs.json](./2026-09-26/platform-costs.json) | `GET /v1/organization/costs` |

이 시점의 `latest_thread_usage.threadUsage`는 `null`이고 Platform 두 응답의 `results`도 비어 있다. 따라서 이 스냅샷에는 thread별 추정 비용이나 Platform 청구 금액 예시가 없다. Codex 구독 한도와 Platform 비용은 서로 다른 데이터이며 합산하지 않는다.
