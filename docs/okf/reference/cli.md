# Reference: CLI

`statusline` 단일 바이너리의 구문·option·subcommand·exit code·환경변수·파일을 정리한다. 설정 값은 [설정 파일 스키마](./configuration-schema.md), STDIN payload는 [CLI 페이로드 명세](./cli-payload-spec.md)를 참조한다.

## 구문

```text
statusline [--cli=<engine> | --auto] [--config=<path>] [init | collect [--provider=<provider>] | watch | tray]
```

subcommand가 없으면 렌더 모드로 동작한다. STDIN으로 JSON을 한 번 읽어 한 줄을 출력하고 종료한다.

## Option

| Option | 기본값 | 설명 |
| :--- | :--- | :--- |
| `--cli=<engine>` | `auto` | 렌더할 host engine. `auto`, `claude`, `codex`, `antigravity`. 별칭 `claudecode`는 `claude`, `agy`는 `antigravity`로 바꾼다 |
| `--auto` | 없음 | `--cli=auto`의 별칭 |
| `--config=<path>` | `~/.config/statusline/config.json` | 설정 파일 경로. `init`은 이 경로에 기본 설정을 쓴다 |

`auto`는 STDIN JSON과 환경변수로 engine을 정한다. 판별 순서는 다음과 같다.

1. `omcLabel`이 있거나 `CLAUDE_CONFIG_DIR`이 설정되어 있으면 `claude`
2. `antigravity` 문자열이 있거나 `ANTIGRAVITY_APP_DIR`이 설정되어 있으면 `antigravity`
3. `product`가 `codex`인 JSON이거나 `CODEX_ENV`이 설정되어 있으면 `codex`
4. 그 밖에는 `generic`

정의되지 않은 engine 이름을 `--cli`에 넣으면 `generic`으로 그린다. 오류가 아니다.

## Subcommand

| Subcommand | 역할 | 언제 끝나나 |
| :--- | :--- | :--- |
| 없음 | STDIN JSON을 한 줄로 렌더 | STDIN을 다 읽으면 |
| `init` | 기본 설정 파일을 `--config` 경로에 기록 | 즉시 |
| `collect` | provider의 잔여량·사용량을 JSON 한 건으로 수집 | 즉시, 최대 10초 |
| `watch` | 최근 Codex thread의 새 token·context event를 JSON Lines로 출력 | `Ctrl+C` |
| `tray` | macOS 메뉴바 위젯 | 드롭다운에서 종료하면 |

### 렌더 모드

- 빈 입력이나 깨진 JSON도 오류 없이 기본 한 줄을 출력한다.
- STDIN이 terminal이면 입력을 읽지 않고 기본 줄만 출력한다.
- stdout에는 결과 한 줄만 쓴다. 오류와 사용법은 stderr에 쓴다.

### `init`

- directory를 `0700`, 파일을 `0600`으로 만든다.
- 이미 있는 파일은 백업 없이 덮어쓴다.
- 출력은 `Initialized default config at <path>`이다.

### `collect`

| `--provider` 값 | 수집 범위 |
| :--- | :--- |
| 생략 또는 `all` | `chatgpt`와 `zai` |
| `chatgpt` | Codex app-server 로컬 조회. 관리자 key가 있으면 OpenAI Platform도 함께 |
| `zai`, `zhipu` | Z.AI quota와 reset ticket |
| `anthropic`, `gemini` | 미지원, exit 2 |

- `--provider`는 `--provider=zai`처럼 `=`를 붙인 형식만 받는다.
- `--cli`는 무시하고 stderr에 경고를 남긴다.
- stdout은 한 줄 JSON이다. 최상위 key는 `codex`, `platform`, `zai`, `errors`이고 값이 없는 key는 생략된다. `errors`의 key는 source 이름(`credentials`, `codex`, `platform`, `zai`)이다.
- `zai` 단독 수집인데 자격 증명이 없으면 stdout 없이 exit 1이다.
- 결과를 `live_session.json`에 병합 기록한다. 성공한 source만 반영된다.

### `watch`

- `codex app-server`의 `thread/list`로 최근 thread의 로컬 JSONL 경로를 찾아 파일 끝에서부터 따라간다.
- 1초마다 파일을 확인해 새 `token_count`, `turn_context` event를 JSON 한 줄씩 출력한다. 사용자 message와 turn 본문은 출력하지 않는다.
- 최근 thread가 없거나 `codex`를 실행할 수 없으면 exit 1이다. 다른 thread로 옮기면 다시 시작한다.

### `tray`

- 5초마다 `live_session.json`을 읽어 menu bar의 title과 menu를 갱신한다.
- cache가 없거나 5분이 지났으면 `collect --provider=all`을 background로 실행한다.
- 표시 규칙은 [트레이 위젯 사용하기](../how-to-guides/use-tray.md)를 따른다.

## Exit code

| Code | 경우 |
| :--- | :--- |
| `0` | 정상 종료. 렌더 모드는 입력이 비어도 `0`이다 |
| `1` | `init` 실패. 자격 증명 없는 `zai` 단독 `collect`. stdout 인코딩 실패. `watch` 실패. `tray`가 cache 경로를 정하지 못함 |
| `2` | 정의되지 않은 option이나 값이 없는 option. `collect`의 알 수 없거나 미지원인 `--provider` |

## 환경변수

| 변수 | 쓰는 곳 | 역할 |
| :--- | :--- | :--- |
| `CLAUDE_CONFIG_DIR`, `ANTIGRAVITY_APP_DIR`, `CODEX_ENV` | `auto` 판별 | 설정되어 있으면 해당 engine |
| `ANTHROPIC_BASE_URL` | 렌더, `collect` | host가 `api.z.ai` 또는 `bigmodel.cn`이면 Z.AI provider로 판정한다. `ZAI_BASE_URL`이 없을 때 Z.AI endpoint로도 쓴다 |
| `ANTHROPIC_AUTH_TOKEN` | 렌더, `collect` | `ZAI_AUTH_TOKEN`이 없을 때 Z.AI token으로 쓴다 |
| `ZAI_AUTH_TOKEN` | 렌더, `collect` | Z.AI token. 가장 우선한다 |
| `ZAI_BASE_URL` | `collect` | Z.AI endpoint 재정의. 가장 우선한다 |
| `OPENAI_ADMIN_KEY` | `collect` | OpenAI Platform 관리자 key. `credentials.json`보다 우선한다 |
| `COLUMNS` | 렌더 | payload에 terminal 폭이 없을 때 쓴다 |

## 파일

| 경로 | 용도 |
| :--- | :--- |
| `~/.config/statusline/config.json` | 표시 설정. `--config`로 바꿀 수 있다 |
| `~/.config/statusline/credentials.json` | 자격 증명. 권한이 `0600`이어야 한다 |
| `<user cache dir>/statusline/zai.json` | Z.AI quota cache. 렌더가 읽는다 |
| `<user cache dir>/statusline/zai.json.refresh` | 렌더 경로의 갱신 marker. 60초 동안 중복 실행을 막는다 |
| `<user cache dir>/statusline/live_session.json` | `collect` 결과 cache. `tray`가 읽는다 |
| `<user cache dir>/statusline/live_session.json.refresh` | `tray`의 갱신 marker. 60초 동안 중복 실행을 막는다 |

user cache dir은 `os.UserCacheDir()` 값이다. Linux는 `~/.cache`, macOS는 `~/Library/Caches`이다.
