# 트레이 위젯 사용하기

macOS 메뉴바와 Windows 시스템 트레이에 AI 구독 잔여량을 표시합니다. 위젯은 `collect`가 기록한 통합 캐시(`live_session.json`)를 읽습니다.

## 사전 준비

1. 빌드: `go build -o statusline ./cmd/statusline`
2. 수집 자격 증명 — [Codex 데이터 수집 가이드](./use-codex.md)의 인증 절차를 먼저 완료하세요. 표시되는 데이터는 `collect` 결과에 의존합니다.

## 위젯 실행

```bash
./statusline tray
```

Windows PowerShell에서는 다음과 같이 실행합니다.

```powershell
go build -o statusline.exe ./cmd/statusline
.\statusline.exe tray
```

macOS에서는 메뉴바 타이틀을 클릭하고, Windows에서는 트레이 아이콘을 우클릭해 메뉴를 엽니다. 종료하려면 메뉴에서 "종료"를 선택하세요.

Windows 아이콘은 provider 코드와 대표 잔여율을 함께 표시합니다. `z87`은 z.ai 5시간 잔여율 87%, `c48`은 ChatGPT 최소 잔여율 48%를 뜻합니다. 데이터가 없으면 `z–` 또는 `c–`가 표시됩니다. 아이콘에 마우스를 올리면 아래 표와 같은 전체 요약을 볼 수 있습니다. 아이콘이 숨겨진 영역에 있으면 Windows 작업 표시줄 설정에서 표시할 수 있습니다.

## macOS 타이틀과 Windows 툴팁 규칙

| 표시 | 의미 |
| :--- | :--- |
| `z87 w45` | zai 5h 잔여 87% · 주간 잔여 45% (주간 값이 없으면 `z87`로 축소) |
| `c48` | chatgpt 최소 잔여 48% (5h/주간 창 중 사용률 최댓값 기준) |
| `c-` | chatgpt rate 정보 없음 |
| `…` | 캐시 없음 (첫 수집 전) |

## 드롭다운 메뉴

- **provider별 상세** — zai 5h/MCP 잔여와 리셋 시간, reset ticket 보유 수와 만료 일자(`zai tickets 2장 · 12/22`, 만료 3일 이하면 ⚠️ 접두), chatgpt rate. 데이터가 없는 provider는 "데이터 없음"으로 표시됩니다.
- **메인 전환** (`✓`/`○`) — 타이틀에 노출할 provider를 바꿉니다. 선택은 `config.json`의 `tray.primary`로 저장되어 재시작 후에도 유지됩니다.
- **지금 갱신** — 캐시 TTL과 무관하게 즉시 수집을 실행합니다(동시 실행 방지 가드는 준수).
- **종료**

## 자동 갱신 구조

갱신 주기와 중복 실행 방지는 [CLI 레퍼런스](../reference/cli.md#tray)의 `tray` 항목을 따릅니다.

## 문제 해결

- **"데이터 없음"이 계속 표시될 때** — 터미널에서 `./statusline collect --provider=all`을 실행해 stdout의 `errors` 항목으로 원인(자격 증명·네트워크)을 확인하세요.
- **타이틀이 `…`에서 바뀌지 않을 때** — 첫 수집이 실패했을 가능성이 있습니다. 위 명령으로 수집이 성공하는지 먼저 확인하세요.
