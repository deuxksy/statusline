# OpenAI API 응답 샘플

이 디렉터리는 위젯의 OpenAI 데이터 수집을 위한 **API별 응답 예시**다. 모두 스키마 기반의 합성 fixture이며 실제 계정 응답이나 인증 정보가 아니다. 실제 조회 스냅샷은 [상위 디렉터리](../README.md)의 날짜별 폴더에 보관한다. 실제 응답에는 선택 필드가 빠지거나 `null`일 수 있다. Codex 로컬 app-server는 JSON-RPC 메서드, OpenAI Platform은 HTTPS REST endpoint를 사용한다.

## 인증

| 출처 | 인증 방법 | 위젯 처리 |
| --- | --- | --- |
| Codex 로컬 app-server | 사용자가 Codex CLI에 로그인해 저장한 계정 상태를 app-server가 사용한다. `account/read`의 `account`와 `requiresOpenaiAuth`로 상태를 확인한다. | 새 API 키를 요구하거나 Codex 인증 파일을 직접 읽지 않는다. 계정이 없고 인증이 필요하면 로그인 필요 상태를 표시한다. |
| OpenAI Platform Usage·Costs | 조직 관리자 API 키(`OPENAI_ADMIN_KEY`)를 Bearer 토큰으로 전송한다. 일반 `OPENAI_API_KEY`와 구분한다. | 별도 opt-in 연결로 둔다. 키는 프로젝트 파일·샘플·로그에 저장하지 않고 사용자 Secret Management(`sops`+`age`)에서 실행 시 주입한다. |

Codex 로그인을 처음 설정할 때는 Codex 자체 로그인 흐름(`codex login` 또는 app-server의 `account/login/start`)을 사용한다. 위젯은 로그인 절차를 대신 구현하지 않는다. 로컬 app-server와 Platform REST는 인증을 공유하지 않으므로 ChatGPT 구독 로그인만으로 조직 Usage·Costs API를 조회할 수 없다.

**연결 규칙:** `statusline collect`는 `OPENAI_ADMIN_KEY`를 우선 사용하고, 없으면 `~/.config/statusline/credentials.json`의 `openai_admin_key`를 읽는다. 파일 권한은 `0600`이어야 한다. 둘 다 없으면 Platform Usage·Costs 요청을 보내지 않고 Codex 로컬 데이터만 수집한다. 일반 `OPENAI_API_KEY`나 Codex 로그인 자격 증명으로 관리자 키를 대체하지 않는다. 관리자 인증 실패는 `errors.platform`에 기록하며 Codex 수집 결과를 유지한다. 출력은 원본 응답을 포함할 수 있으므로 공개 저장소에 복사하지 않는다. 위젯 UI 연결은 후속 로드맵 범위다.

`collect`의 `codex.latest_thread`는 최근 수정된 thread의 모델·작업 경로·추론 수준을 포함한다. `codex.latest_thread_usage`는 thread별 추정 사용량 응답이며, 백엔드가 비용을 제공하지 않으면 `threadUsage`가 `null`이다. `statusline watch`는 현재 최근 thread의 로컬 JSONL에서 새 토큰·컨텍스트 이벤트를 읽는다. app-server 알림 구독은 아직 구현하지 않았다.

| 파일 | 요청 또는 알림 | 데이터 범위 |
| --- | --- | --- |
| [codex-account.json](./codex-account.json) | `account/read` | 계정 유형, ChatGPT 플랜, 인증 필요 여부, workspace routing |
| [codex-rate-limits.json](./codex-rate-limits.json) | `account/rateLimits/read` | 구독 한도, 잔여 크레딧, 초기화 시각, 제한 상태 |
| [codex-account-usage.json](./codex-account-usage.json) | `account/usage/read` | 계정 누적·일별 토큰 사용량 |
| [codex-thread-usage.json](./codex-thread-usage.json) | `account/usage/read` + `threadId` | thread 모델별 토큰과 추정 크레딧·USD 사용량 |
| [codex-token-usage-event.json](./codex-token-usage-event.json) | `thread/tokenUsage/updated` | 실행 중 thread의 마지막·누적 토큰과 컨텍스트 크기 |
| [platform-completions-usage.json](./platform-completions-usage.json) | `GET /v1/organization/usage/completions` | API 모델 요청·토큰 집계 |
| [platform-costs.json](./platform-costs.json) | `GET /v1/organization/costs` | API 청구 비용 집계 |

## Codex app-server 필드

- `account/read`: `account`는 `apiKey`, `chatgpt`, `amazonBedrock` 유형 중 하나다. ChatGPT 유형은 `email`과 `planType`을 포함한다. 응답에는 `requiresOpenaiAuth`, `workspaceRouting`도 있다. `email`, account ID, thread ID는 실제 샘플에 넣지 않는다.
- `account/rateLimits/read`: `rateLimits` 단일 bucket과 선택적 `rateLimitsByLimitId` 다중 bucket, `accountId`, `ordinaryUsageAllowed`, `rateLimitResetCredits`, `rateLimitUpsell`을 반환한다. bucket 필드는 `limitId`, `limitName`, `normalModelSlug`, `planType`, `primary`, `secondary`, `credits`, `individualLimit`, `spendControlReached`, `rateLimitReachedType`이다. `primary`/`secondary`에는 `usedPercent`, `windowDurationMins`, `resetsAt`이 있다. `credits`에는 `hasCredits`, `unlimited`, `balance`가 있다. reset credit 요약에는 `availableCount`, `credits[]`; 각 credit에는 `id`, `status`, `resetType`, `grantedAt`, `expiresAt`, `title`, `description`이 있다.
- `account/usage/read`: `summary`에는 `lifetimeTokens`, `peakDailyTokens`, `longestRunningTurnSec`, `currentStreakDays`, `longestStreakDays`; `dailyUsageBuckets[]`에는 `startDate`, `tokens`가 있다. `threadId`를 전달하면 `threadUsage`가 제공될 수 있으며 `threadId`, `estimatedUsageCreditsMicros`, 선택적 `estimatedUsageUsdMicros`, `groups[]`를 포함한다. group에는 `model`, `reasoningEffort`, `speed`, `inputTokens`, `cachedInputTokens`, `netNewInputTokens`, `outputTokens`, `totalTokens`, `estimatedUsageCreditsMicros`가 있다.
- `thread/tokenUsage/updated`: `threadId`, `turnId`, `tokenUsage`를 보낸다. `tokenUsage`는 `last`, `total`, 선택적 `modelContextWindow`로 구성된다. 토큰 집계 필드는 `inputTokens`, `cachedInputTokens`, 선택적 `cacheWriteInputTokens`, `outputTokens`, `reasoningOutputTokens`, `totalTokens`다.
- `account/rateLimits/updated`: 변경된 `rateLimits`만 오는 sparse 알림이다. 누락 필드가 기존 값을 지운다는 뜻은 아니다. 최신 `account/rateLimits/read` 스냅샷에 병합하거나 재조회한다.
- `thread/list`·`thread/read`: 위젯에서 세션별 상태가 필요하면 사용한다. `thread/list`는 `data[]`, `nextCursor`, `backwardsCursor`를 반환한다. thread 항목에는 `id`, `sessionId`, `name`, `model`, `reasoningEffort`, `modelProvider`, `cwd`, `status`, `createdAt`, `updatedAt`, `gitInfo`, `cliVersion` 등이 있다. `thread/read`는 `thread`를 반환한다. 메시지 본문이 들어갈 수 있는 `turns`·`preview`는 위젯용 캐시에 보관하지 않는다.

`resetsAt`은 Unix 초이며 `null`이면 알 수 없는 시각이다. `estimatedUsageCreditsMicros`는 크레딧의 백만분의 1 단위이고, `estimatedUsageUsdMicros`는 USD의 백만분의 1 단위다. 둘 다 추정치이며 Platform Costs의 청구 금액과 합산하지 않는다.

## OpenAI Platform 필드

- 두 REST 응답 모두 `object: "page"`, `data[]`의 `start_time`·`end_time`·`results[]`, `has_more`, `next_page`를 제공한다. `next_page`가 있으면 다음 페이지를 조회한다.
- Completions 사용량 결과의 필수 필드: `object`, `input_tokens`, `output_tokens`, `num_model_requests`. 선택 필드: `api_key_id`, `batch`, `input_audio_tokens`, `input_cache_write_tokens`, `input_cached_audio_tokens`, `input_cached_image_tokens`, `input_cached_text_tokens`, `input_cached_tokens`, `input_image_tokens`, `input_text_tokens`, `input_uncached_tokens`, `model`, `output_audio_tokens`, `output_image_tokens`, `output_text_tokens`, `project_id`, `service_tier`, `user_id`. `model` 등 분류 키는 해당 `group_by`를 지정해야 채워진다.
- Costs 결과의 필드: `object`, 선택적 `amount.currency`·`amount.value`, `api_key_id`, `line_item`, `project_id`, `quantity`, `quantity_unit`. `quantity_unit`에는 `tokens`, `1000_tokens`, `duration_seconds`, `duration_minutes`, `duration_hours`, `gibibyte_hours`, `images`, `characters` 등이 있다.
- Completions 조회는 필수 `start_time`과 선택적 `end_time`, `bucket_width`(`1m`/`1h`/`1d`), `group_by`, `project_ids`, `user_ids`, `api_key_ids`, `models`, `batch`, `limit`, `page` 등을 사용한다. Costs 조회는 필수 `start_time`과 선택적 `end_time`, `bucket_width`(`1d`), `group_by`(`project_id`/`line_item`/`api_key_id`), `project_ids`, `api_key_ids`, `limit`, `page` 등을 사용한다.
- Platform에는 Completions 외에도 audio speeches, audio transcriptions, code interpreter sessions, embeddings, file search calls, images, moderations, vector stores, web search calls의 별도 Usage endpoint가 있다. 이들 endpoint는 각 서비스의 사용량을 반환한다. 위젯이 해당 서비스를 표시하기 전에는 Completions 토큰이나 Costs 금액으로 환산하지 않는다.

출처: [Codex app-server 프로토콜](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/common.rs), [Codex 계정 스키마](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/v2/account.rs), [Codex app-server 인증](https://github.com/openai/codex/blob/main/codex-rs/app-server/README.md), [OpenAI Admin API](https://developers.openai.com/api/docs/guides/admin-apis), [OpenAI Completions Usage](https://developers.openai.com/api/reference/resources/admin/subresources/organization/subresources/usage/methods/completions), [OpenAI Costs](https://developers.openai.com/api/reference/resources/admin/subresources/organization/subresources/usage/methods/costs).
