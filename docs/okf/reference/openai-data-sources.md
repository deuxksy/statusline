# OpenAI 데이터 출처와 위젯 수집 범위

Codex 구독 상태와 OpenAI Platform API 사용량은 별도 데이터다. 같은 `openai` 라벨로 합산하거나, Codex 토큰 수에 API 단가를 곱해 청구 비용으로 표시하지 않는다. 아래 필드는 Codex CLI 0.157.0의 `codex app-server generate-json-schema --experimental` 결과와 OpenAI API 문서를 기준으로 확인했다.

API별 합성 예시와 전체 필드 목록은 [`docs/samples/openai/schema-examples/`](../../samples/openai/schema-examples/README.md)에 기록한다.

실제 조회 응답은 식별자를 제거해 [`docs/samples/openai/`](../../samples/openai/README.md)에 날짜별로 보관한다. 합성 예시와 실제 스냅샷은 하위 디렉터리로 구분한다.

| 출처 | 조회 | 가져올 수 있는 데이터 | 위젯 용도 |
| --- | --- | --- | --- |
| Codex 로컬 app-server | `account/rateLimits/read` | `rateLimits` 및 `rateLimitsByLimitId`의 `limitId`, `planType`, `primary`/`secondary`의 `usedPercent`, `windowDurationMins`, `resetsAt`, `credits`, `rateLimitReachedType`, `spendControlReached` | 구독 플랜, 5시간·주간 잔여량과 reset 시각 |
| Codex 로컬 app-server | `account/usage/read` | `summary`의 `lifetimeTokens`, `peakDailyTokens`, `currentStreakDays`, `longestStreakDays`, `longestRunningTurnSec`; `dailyUsageBuckets`의 날짜별 토큰 | 계정 사용량 추이 |
| Codex 로컬 app-server | `account/usage/read` + `threadId` | `threadUsage.estimatedUsageCreditsMicros`, 선택적 `estimatedUsageUsdMicros`, 모델·추론 수준별 토큰/추정 크레딧 사용량 | thread별 사용량과 추정 비용. 값이 `null`이면 표시하지 않음 |
| Codex 로컬 app-server | `thread/tokenUsage/updated`, `account/rateLimits/updated` 알림 | 실행 중 thread 토큰 및 한도 변경 이벤트 | 위젯 갱신. 한도 알림은 일부 필드만 올 수 있으므로 마지막 조회 결과와 병합하거나 재조회 |
| OpenAI Platform REST | `GET /v1/organization/usage/completions` | 기간별 `input_tokens`, `input_cached_tokens`, `output_tokens`, `num_model_requests`; 지정한 `group_by`에 따라 `model`, `project_id`, `api_key_id` 등 | API 호출량 집계 |
| OpenAI Platform REST | `GET /v1/organization/costs` | 기간별 `amount.value`, `amount.currency`; 지정한 `group_by`에 따라 `project_id`, `line_item` | API 청구 비용 집계 |

로컬 app-server는 Codex 프로세스의 프로토콜 API다. Platform REST API와 URL·인증 체계가 다르다. Platform 조직 사용량과 비용 조회에는 해당 조직의 관리자 API 키가 필요하며, 키를 샘플이나 로그에 저장하지 않는다. 사용량과 비용 집계는 기록 기준이 달라 정확히 일치하지 않을 수 있으므로 비용 표시는 Costs 응답을 기준으로 한다.

연결은 독립적이다. `OPENAI_ADMIN_KEY`가 있으면 우선 사용하고, 없으면 `~/.config/statusline/credentials.json`의 `openai_admin_key`를 읽는다. 이 파일은 권한 `0600`이어야 한다. 두 출처 모두에 키가 없으면 Platform Usage·Costs 요청을 생략하고 Codex 로컬 조회만 유지한다. 일반 `OPENAI_API_KEY`는 관리자 조회의 fallback으로 사용하지 않는다. 관리자 인증 실패도 Codex 데이터 수집을 중단하지 않는다.

## 한도 초기화 시각

Codex app-server의 `primary`와 `secondary`에는 각각 Unix 초 단위 `resetsAt`이 있다. `windowDurationMins`는 시간창 길이이며, 실제 초기화 시각은 `resetsAt`을 사용한다. 위젯은 이 값을 현지 시간대로 표시하고 `max(0, resetsAt - 현재 Unix 시각)`으로 남은 시간을 계산한다. `resetsAt`이 `null`이면 초기화 시각을 알 수 없는 것으로 표시한다. 초기화 시각이 지났더라도 이전 사용률을 0으로 가정하지 않고 새 `account/rateLimits/read` 결과를 받아 갱신한다.

`docs/samples/codex.json`의 수집 당시 값은 5시간 창 `1790366542`(2026-09-26 05:02:22 KST), 주간 창 `1790573243`(2026-09-28 14:27:23 KST)이다. 샘플은 고정된 스냅샷이므로 이 시각은 현재 계정 상태를 뜻하지 않는다. OpenAI Platform Usage·Costs의 일별 bucket 경계도 Codex 구독 한도 초기화 시각과 동일한 개념이 아니다.

현재 `docs/samples/codex.json`은 로컬 세션 JSONL에서 추출한 fixture다. `turn_context`의 모델·작업 경로·추론 수준과 `token_count`의 토큰·한도 스냅샷을 합쳤다. Codex가 statusline에 전달하는 STDIN 계약이 아니며, app-server 응답을 그대로 복사한 것도 아니다. 어댑터는 이 fixture의 모델, 최근·누적 토큰, 컨텍스트 크기, 5시간·주간 한도를 파싱한다. 별도 `statusline collect` 명령은 Codex app-server의 `account/read`, `account/rateLimits/read`, `account/usage/read` 응답을 `codex` 객체에 수집한다. 관리자 키가 환경변수 또는 자격 증명 파일에 있을 때만 직전 1일부터 현재까지의 Platform Completions Usage·Costs를 페이지 끝까지 조회해 `platform` 객체에 넣는다. 키가 없으면 `platform` 객체도 요청도 없다. 출처별 오류는 `errors`에 기록한다. 수집 결과에는 계정 식별 정보가 들어갈 수 있으므로 공개 파일에 저장하지 않는다. 렌더러의 Platform 비용 표시는 아직 구현되지 않았다.

`collect`는 최근 수정된 thread의 `id`, 모델, 추론 수준, 작업 경로, 상태, 수정 시각을 `codex.latest_thread`에 넣고, 같은 thread의 `account/usage/read` 응답을 `codex.latest_thread_usage`에 넣는다. 백엔드가 추정 비용을 제공하지 않으면 `threadUsage`는 `null`이며 임의 비용을 계산하지 않는다. `statusline watch`는 실행 시점의 최근 thread JSONL을 따라가며 새 `token_count`와 `turn_context` 이벤트를 JSON Lines로 출력한다. 사용자 메시지와 turn 본문은 출력하지 않는다. watch가 실행된 뒤 활성 thread가 바뀌면 다시 시작해야 한다. 이는 app-server의 `thread/tokenUsage/updated` 알림 구독이 아니라 Codex 로컬 세션 파일의 변경 감시다.

참고 자료: [Codex app-server 메서드](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/common.rs), [Codex 계정 응답 타입](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/v2/account.rs), [OpenAI Platform Usage·Costs API](https://platform.openai.com/docs/api-reference/usage).
