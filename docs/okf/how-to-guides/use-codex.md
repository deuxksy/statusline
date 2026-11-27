# Codex 데이터 수집 사용 가이드

현재 Codex 연동은 **조회용 CLI**와 **MCP 상태바**다. `collect`는 현재 상태를 JSON으로 가져오고, `watch`는 로컬 세션에 새로 기록되는 토큰·컨텍스트 이벤트를 출력한다. 수집 결과는 [트레이 위젯 가이드](./use-tray.md)의 위젯이 캐시를 통해 표시하고, Codex 하단 상태바는 6절의 MCP provider가 그린다.

## collect — 전 provider 수집

```bash
statusline collect                    # 전 provider 병합 (chatgpt+zai)
statusline collect --provider=chatgpt # Codex + OpenAI Platform (Admin key 조건부)
statusline collect --provider=zai     # Z.AI quota (zhipu 별칭 지원)
```

`--provider`는 서비스 계정(chatgpt, zai)을, `--cli`는 호스트 CLI(claudecode, codex, agy)를 구분한다. `anthropic`·`gemini`는 아직 미지원(exit 2).

## 1. 최신 바이너리 빌드

프로젝트 루트에서 실행한다.

```sh
go build -o statusline ./cmd/statusline
```

Windows PowerShell에서는 `go build -o statusline.exe ./cmd/statusline`으로 빌드한다. 아래의 `./statusline` 명령은 `./statusline.exe`로 실행한다.

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

Platform 조회는 Codex 구독 상태와 별도다. 키는 환경변수 `OPENAI_ADMIN_KEY`로 두거나 `~/.config/statusline/credentials.json`에 넣는다. macOS와 Linux에서는 파일 권한을 `0600`으로 설정한다. Windows의 경로는 `%USERPROFILE%\.config\statusline\credentials.json`이다. 파일 형식은 다음과 같다. 예시 값은 실제 키가 아니다.

```json
{"openai_admin_key":"<관리자 키>"}
```

조회 우선순위, 키가 없을 때의 동작, 오류 처리는 [OpenAI 데이터 출처](../reference/openai-data-sources.md#openai-platform-연결)의 연결 규칙을 따른다. `~/.config/statusline/config.json`은 표시 설정용이며 인증 키를 읽지 않는다.

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

Windows PowerShell에서는 `Get-Content -Raw -Encoding utf8 docs/samples/codex.json | ./statusline.exe --cli=codex`를 실행한다.

이 샘플의 성격은 [CLI 페이로드 명세](../reference/cli-payload-spec.md#3-codex-widget-snapshot)를 따른다. 이 경로의 렌더러는 `collect` 결과를 받지 않는다. Codex 하단에 표시하려면 6절의 MCP provider를 등록한다.

## 6. Codex MCP 상태바 설정

Windows 11의 Codex CLI `0.172.0`에서 확인한 절차다. 먼저 저장소 루트에서 바이너리를 빌드하고 `codex mcp list`로 `statusline-tickets`가 등록돼 있는지 확인한다. 이미 등록돼 있다면 `codex mcp add`는 생략한다.

**Windows PowerShell**

```powershell
go build -o statusline.exe ./cmd/statusline
codex mcp list
```

등록이 필요하면 아래 명령을 실행한다. `$python`에는 Python 실행 파일의 절대 경로를 사용한다.

```powershell
$python = python -c 'import sys; print(sys.executable)'
$provider = (Resolve-Path -LiteralPath './integrations/codex/statusline.py').Path
codex mcp add statusline-tickets -- "$python" -u "$provider"
```

**macOS·Linux**에서는 `go build -o statusline ./cmd/statusline`으로 빌드하고, 등록이 필요하면 `codex mcp add statusline-tickets -- python3 -u "$PWD/integrations/codex/statusline.py"`를 실행한다.

`codex features enable tui_extensions_statusline`을 실행하거나 `/experimental`에서 **Custom status line**을 켠다. Codex CLI를 종료하고 새 터미널에서 `codex`를 실행한 뒤 `/statusline` → **Choose a status line provider** → **Statusline · 초기화 티켓**을 선택한다.

목록에 **Default status line**만 표시되거나 `F2`에 provider 검색 시간 초과가 표시되면, 사용자 `~/.codex/config.toml`의 기존 `[tui]` 섹션에 다음 값을 설정한다. `status_line_provider`가 이미 있으면 해당 값만 교체한다.

```toml
[tui]
status_line_provider = { server = "statusline-tickets", tool = "statusline" }
```

새 터미널에서 `codex`를 실행해 하단에 `CODEX`로 시작하는 상태바가 표시되는지 확인한다. `/create-statusline`은 새 provider를 작성할 때 사용한다. 이 절차는 저장소의 provider를 등록하거나 재사용한다.

## 문제 해결

- `collect` 대신 `GENERIC`이 출력된다면 1단계의 빌드를 다시 수행한다.
- `errors.codex`가 있으면 Codex CLI 설치·로그인 상태와 app-server 실행 가능 여부를 확인한다.
- `errors.credentials`가 있으면 자격 증명 파일의 JSON 형식을 확인한다. macOS와 Linux에서는 `0600` 권한도 확인한다.
- `errors.platform`에 HTTP 403이 있으면 API Platform 조직 관리자 키와 조회 권한을 확인한다.
- `watch`에 출력이 없다면 선택된 thread에 새 기록이 생겼는지 확인한다.

API별 필드와 데이터 출처는 [OpenAI 데이터 출처](../reference/openai-data-sources.md)를 참고한다.
