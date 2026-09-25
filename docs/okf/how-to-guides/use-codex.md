# Codex 데이터 수집 사용 가이드

현재 Codex 연동은 **조회용 CLI**다. `collect`는 현재 상태를 JSON으로 가져오고, `watch`는 로컬 세션에 새로 기록되는 토큰·컨텍스트 이벤트를 출력한다. 수집 결과를 상태바나 OS 위젯에 자동 표시하는 기능은 아직 없다.

## 1. 최신 바이너리 빌드

프로젝트 루트에서 실행한다.

```sh
go build -o statusline ./cmd/statusline
```

코드를 수정한 뒤 다시 빌드하지 않으면 이전 바이너리는 `collect`와 `watch`를 알지 못해 `GENERIC` 또는 `CODEX │ codex` 상태바를 출력할 수 있다.

## 2. Codex 계정·한도·사용량 조회

먼저 `codex login`으로 Codex CLI에 로그인한 상태여야 한다. 프로젝트는 Codex 인증 파일을 직접 읽거나 로그인 상태를 변경하지 않는다.

```sh
./statusline collect
```

출력은 `codex.account`, `codex.rate_limits`, `codex.usage`, `codex.latest_thread`, `codex.latest_thread_usage`와 선택적 `platform.usage`, `platform.costs`로 나뉜다. 값 대신 수집 여부만 확인하려면 다음처럼 실행한다.

```sh
./statusline collect | jq '{codex: (.codex // {} | keys), platform: (.platform // {} | keys), errors: (.errors // {} | keys)}'
```

`collect` 원본에는 계정 식별 정보와 로컬 경로가 포함될 수 있으므로 공개 문서나 이슈에 그대로 붙여 넣지 않는다. `latest_thread_usage.threadUsage`가 `null`이면 해당 thread의 추정 비용을 백엔드가 제공하지 않는 상태다.

## 3. OpenAI Platform Usage·Costs 연결

Platform 조회는 Codex 구독 상태와 별도다. `OPENAI_ADMIN_KEY`가 있으면 이를 우선 사용하고, 없으면 `~/.config/statusline/credentials.json`에서 `openai_admin_key`를 읽는다. 파일 권한은 `0600`이어야 한다. 현재 로컬 설정 파일의 구조는 다음과 같다. 예시 값은 실제 키가 아니다.

```json
{"openai_admin_key":"<관리자 키>"}
```

관리자 키가 없으면 Platform 요청을 보내지 않고 Codex 로컬 데이터만 수집한다. 일반 `OPENAI_API_KEY`는 대체 키로 사용하지 않는다. Platform 응답 오류는 `errors.platform`에 표시되며 Codex 수집 결과와 분리된다. `~/.config/statusline/config.json`은 표시 설정용이며 인증 키를 읽지 않는다.

## 4. 실행 중 토큰·컨텍스트 변경 보기

```sh
./statusline watch
```

실행 시점에 가장 최근 수정된 Codex thread의 로컬 JSONL 파일 끝을 따라간다. 이후 새 `token_count` 또는 `turn_context` 기록이 생기면 JSON 한 줄씩 출력한다. 사용자 메시지와 turn 본문은 출력하지 않는다. 새 이벤트가 없으면 화면도 조용하다. `Ctrl+C`로 종료한다. 다른 thread로 작업을 옮기면 `watch`를 다시 시작한다.

이 명령은 **app-server 알림 구독이 아니라 로컬 파일 감시**다. `thread/tokenUsage/updated`·`account/rateLimits/updated` 알림 수신은 아직 구현하지 않았다.

## 5. 상태바 렌더러 확인

`--cli=codex`는 STDIN JSON을 한 줄로 렌더링하는 기존 경로다. `collect`·`watch`에는 필요하지 않다.

```sh
./statusline --cli=codex < docs/samples/codex.json
```

이 샘플은 로컬 세션에서 정규화한 fixture이며 app-server 원본 응답이 아니다. 현재 렌더러는 `collect` 결과를 자동으로 받아 표시하지 않는다. Codex 내장 `/statusline` 명령도 이 바이너리로 교체되지 않는다.

## 문제 해결

- `collect` 대신 `GENERIC`이 출력된다면 1단계의 빌드를 다시 수행한다.
- `errors.codex`가 있으면 Codex CLI 설치·로그인 상태와 app-server 실행 가능 여부를 확인한다.
- `errors.credentials`가 있으면 자격 증명 파일의 JSON 형식과 `0600` 권한을 확인한다.
- `errors.platform`에 HTTP 403이 있으면 API Platform 조직 관리자 키와 조회 권한을 확인한다.
- `watch`에 출력이 없다면 선택된 thread에 새 기록이 생겼는지 확인한다.

API별 필드와 데이터 출처는 [OpenAI 데이터 출처](../reference/openai-data-sources.md)를 참고한다.
