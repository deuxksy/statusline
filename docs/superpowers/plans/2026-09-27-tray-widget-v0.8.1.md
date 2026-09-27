# v0.8.1 macOS 트레이 위젯 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** macOS 메뉴바 트레이 위젯(`statusline tray`) — collect 통합 캐시(live_session.json)를 읽어 메인 provider 잔여 타이틀·드롭다운을 표시하고 TTL 만료 시 자가 스폰 갱신한다.

**Architecture:** 캐시 계층(`internal/collect/live_cache.go`)은 기존 zai_cache 패턴(atomic write·marker·TTL)을 live_session.json으로 확장하고 `runCollect`가 병합 적재한다. 표시 계층(`internal/tray`)은 systray 의존 없이 순수 함수(타이틀/메뉴 문자열, 갱신 판정)로 구성하고, `cmd/statusline/traycmd.go`만 systray(AppKit)를 직접 다룬다.

**Tech Stack:** Go 1.26 · `fyne-io/systray` (신규 의존성, traycmd/tray 패키지에 국한)

**Spec:** `docs/superpowers/specs/2026-09-27-tray-widget-design.md`

## Global Constraints

- fail-soft: 트레이 프로세스가 어떤 입력·파손·I/O 실패로도 크래시하지 않는다. 에러를 stdout에 내지 않는다.
- 상수: 폴링 5s · 캐시 TTL 5분(`FetchedAt` 기준, mtime 아님) · refresh marker `O_EXCL` 60s.
- marker 계약: 갱신 완료 시 제거, 스폰 실패 시 유지(60s 백프레셔 — 5초 폴링이 매 주기 재스폰하지 않는다). `zai_attach.go` 관행 대칭.
- 타이틀 형식: zai `z87 w45`(주간 부재/0이면 `z87`) · chatgpt `c80`(미제공 `c-`) · 캐시 부재/파손 `…`.
- `tray.primary`: 기본 `"zai"`, 유효값 `zai`|`chatgpt`. 무효값은 렌더에서 `"zai"` 폴백(경고 없음).
- `zai.json`·기존 CLI 어태치 경로는 무손상 (병행 유지, 이중 기록 의도됨).
- systray import는 `cmd/statusline`(traycmd)과 `internal/tray` 진입 구조에만 허용. `internal/collect`는 신규 외부 의존 금지.
- 완료 조건: `go test -race ./...` 전체 통과 + `go build ./...` + `GOOS=windows go build ./...`.

## Review Focus

1. **첫 실행 — 캐시 디렉터리 자체 부재**: `~/Library/Caches/statusline/`이 없어도 쓰기가 성공해야 한다 → Task 1 `TestWriteLiveCacheAtCreatesDir`(존재하지 않는 하위 경로에 쓰기).
2. **스폰 스톰**: 스폰이 계속 실패해도 5초마다 재스폰하지 않는다(marker 60s 백프레셔) → Task 5 `TestTryRefreshBackpressure`.
3. **rate_limits 스키마 변화·파손**: codex 스냅샷이 어떤 형태여도 타이틀은 `c-`로 흡수 → Task 4 `TestChatGPTTitleMalformed`.
4. **주간 0% vs 부재 구분 불가**(omitempty): `WeeklyRemaining` 0이면 `w00`이 아니라 주간 미표시로 축소 → Task 4 `TestTitleZaiNoWeekly`.
5. **동시 writer tmp 경합**: tmp 경로가 고정(`.tmp`)이라 CLI collect·트레이 스폰 완료가 겹치면 한쪽 rename 실패 가능 → Write 실패는 호출자(runCollect)가 무시(fail-soft) 계약이므로 Task 1 `TestWriteLiveCacheAt` 실패 반환 값 테스트로 고정.

---

### Task 1: LiveSession 캐시 모듈

**Files:**
- Create: `internal/collect/live_cache.go`
- Test: `internal/collect/live_cache_test.go`

**Interfaces:**
- Consumes: 기존 `TryMarkZaiRefreshAt(path string, maxAge time.Duration) bool`·`ClearZaiRefreshAt(path string)`(`zai_cache.go` — 경로 기반이라 그대로 재사용, 이름에 zai가 있어도 계약 동일)
- Produces (Task 2·4·5·6이 사용):

```go
type LiveSession struct {
    FetchedAt time.Time                  `json:"fetchedAt"`
    Codex     map[string]json.RawMessage `json:"codex,omitempty"`
    Platform  *PlatformSnapshot          `json:"platform,omitempty"`
    Zai       *ZaiSnapshot               `json:"zai,omitempty"`
}
func LiveCachePath() (string, error)                    // UserCacheDir()/statusline/live_session.json
func LiveRefreshPath(cachePath string) string           // cachePath + ".refresh"
func ReadLiveCacheAt(path string) (LiveSession, bool)   // 부재·파손·FetchedAt zero → false
func WriteLiveCacheAt(path string, live LiveSession) error // MkdirAll + tmp+rename 0644
func MergeLiveSession(prev LiveSession, codex map[string]json.RawMessage, platform *PlatformSnapshot, zai *ZaiSnapshot, now time.Time) LiveSession
```

`MergeLiveSession` 계약: 인자가 non-nil(nil map 아님/포인터 non-nil)인 provider만 `prev` 필드를 교체하고 나머지는 유지. `FetchedAt = now`. **호출자는 성공 provider가 1개 이상일 때만 쓴다**(전 실패 시 prev를 그대로 두고 쓰지 않음).

- [ ] **Step 1: 실패 테스트 작성** — `TestWriteReadLiveCacheRoundTrip`(쓰기→읽기 필드 보존), `TestWriteLiveCacheAtCreatesDir`(`t.TempDir()+"/statusline"` 미생성 경로), `TestReadLiveCacheAtCorrupt`(파손 JSON·빈 파일 → false), `TestMergeLiveSessionPreservesOthers`(prev에 Codex 있고 zai만 갱신 → Codex 유지·Zai 교체·FetchedAt 갱신), `TestMergeLiveSessionNilKeeps`(zai nil → prev.Zai 유지)
- [ ] **Step 2: 실패 확인** — `go test -race ./internal/collect -run 'LiveCache|MergeLive' -v` → 컴파일 실패(미정의)
- [ ] **Step 3: `live_cache.go` 구현** — `ReadZaiCacheAt`/`WriteZaiCacheAt`(`zai_cache.go`)를 구조 참조해 그대로 대칭 작성. tmp는 `path+".tmp"`(zai 관행)
- [ ] **Step 4: 통과 확인** — 위 명령 PASS
- [ ] **Step 5: Commit** — `git add internal/collect/live_cache.go internal/collect/live_cache_test.go && git commit -m "feat(collect): live_session.json 통합 캐시 모듈 추가"`

### Task 2: collect 파이프라인 캐시 적재

**Files:**
- Modify: `cmd/statusline/collectcmd.go:93` (runCollect), `cmd/statusline/main.go:70-75` (호출부)
- Test: `cmd/statusline/collectcmd_test.go`

**Interfaces:**
- Consumes: Task 1 전부
- Produces: `runCollect(stdout, stderr io.Writer, providers []string, credsPath, zaiCachePath, liveCachePath string) int` — 시그니처에 `liveCachePath` 추가

적재 규칙: stdout 인코딩 성공 후, 성공 provider가 1개 이상(`out.Codex != nil || out.Platform != nil || out.Zai != nil`)이면 `ReadLiveCacheAt → MergeLiveSession → WriteLiveCacheAt`, 이어 `ClearZaiRefreshAt(LiveRefreshPath(liveCachePath))`. 전 실패면 캐시를 건드리지 않는다.

- [ ] **Step 1: 실패 테스트 작성** — 기존 runCollect 호출부(테스트 전체)에 `liveCachePath` 인자 갱신. 신규 `TestRunCollectWritesLiveCache`(zai 성공 시나리오 — 기존 zai httptest 패턴 재사용 — 종료 후 live_session.json에 `zai` 키·`fetchedAt` 존재), `TestRunCollectAllFailureKeepsCache`(사전에 캐시 파일 기록 → 자격 증명 없는 collect → 캐시 바이트 불변)
- [ ] **Step 2: 실패 확인** — `go test -race ./cmd/statusline -run RunCollect -v`
- [ ] **Step 3: 구현** — runCollect 말미(stdout 인코드 직후)에 적재 규칙 삽입, main.go collect 분기에서 `collect.LiveCachePath()` 전달
- [ ] **Step 4: 통과 확인** — `go test -race ./...`
- [ ] **Step 5: Commit** — `git commit -m "feat(collect): runCollect live_session 병합 적재 및 marker 제거"`

### Task 3: TrayConfig + 저장 API

**Files:**
- Modify: `internal/config/config.go`, `internal/config/config_test.go`

**Interfaces:**
- Produces: `type TrayConfig struct { Primary string \`json:"primary"\` }`, `Config.Tray TrayConfig`, `SaveConfig(path string, cfg Config) error`(main 전환 저장용 — 기존 저장 경로와 동일한 인코딩 사용), `LoadConfig`의 기본값 `Tray{Primary: "zai"}`(파일에 tray 없거나 빈 값일 때)

- [ ] **Step 1: 실패 테스트 작성** — `TestLoadConfigTrayDefault`(빈/부재 config → `cfg.Tray.Primary == "zai"`), `TestLoadConfigTrayValue`(명시 값 로드), `TestSaveConfigRoundTrip`(Save→Load 보존)
- [ ] **Step 2: 실패 확인** — `go test -race ./internal/config -v`
- [ ] **Step 3: 구현** — `SaveDefaultConfig`의 기본 문서에도 `"tray": {"primary": "zai"}` 포함
- [ ] **Step 4: 통과 확인** — 동일 명령 PASS
- [ ] **Step 5: Commit** — `git commit -m "feat(config): tray.primary 설정 및 저장 API 추가"`

### Task 4: 타이틀·메뉴 렌더 순수 함수

**Files:**
- Create: `internal/tray/render.go`, `internal/tray/render_test.go`

**Interfaces:**
- Consumes: Task 1 `LiveSession`·`ZaiSnapshot`(`TokenRemaining`, `WeeklyRemaining` — 0~1.0)
- Produces (Task 6 사용):

```go
func Title(primary string, live collect.LiveSession) string
func ChatGPTTitle(codex map[string]json.RawMessage) string // rate_limits 미제공·파손 → "c-"
func MenuLines(live collect.LiveSession, primary string, now time.Time) []string
```

`Title` 계약: primary 무효값 → `"zai"` 폴백. `live.Zai == nil`(또는 live 전체 빈) → `"…"`. zai: `"z" + pct(TokenRemaining)`, `WeeklyRemaining > 0`이면 ` + " w" + pct(WeeklyRemaining)`. pct = clamp(0,100, round(v×100)). chatgpt: `ChatGPTTitle(live.Codex)` — rate_limits에서 잔여 최솟값, 실패 `"c-"`(경로·스키마는 방어적 파싱). `MenuLines`: provider별 상세(zai 5h/MCP 잔여·리셋 로컬타임 `HH:MM`/`MM-DD`, 항목 부재 시 생략) + 메인 전환 표시(`✓ zai`/`○ chatgpt`) + `지금 갱신` + `종료`.

- [ ] **Step 1: 실패 테스트 작성** — `TestTitleZai`(0.87/0.45 → `z87 w45`), `TestTitleZaiNoWeekly`(0.87/0 → `z87`), `TestTitleInvalidPrimaryFallback`(`"gpt"` → zai 렌더), `TestTitleMissing`(빈 LiveSession → `…`), `TestChatGPTTitleMalformed`(nil·빈 map·가짜 JSON → `c-`), `TestChatGPTTitleWithRateLimits`(테스트 픽스처 rate_limits JSON 주입 → `c80` 형태 — 파서가 항상 `c-` 폴백만 반환하는지 감지), `TestMenuLines`(zai 항목·전환 마크·생략 확인)
- [ ] **Step 2: 실패 확인** — `go test -race ./internal/tray -v` → 패키지 없음/컴파일 실패
- [ ] **Step 3: 구현** — systray import 금지(순수 함수). rate_limits 실측 전까지 파싱은 최소 방어적 스켈레톤으로 `c-` 폴백 우세
- [ ] **Step 4: 통과 확인** — 동일 명령 PASS
- [ ] **Step 5: Commit** — `git commit -m "feat(tray): 타이틀·메뉴 렌더 순수 함수 추가"`

### Task 5: 갱신 감시자 (폴링·marker·스폰)

**Files:**
- Create: `internal/tray/watch.go`, `internal/tray/watch_test.go`

**Interfaces:**
- Consumes: Task 1(`ReadLiveCacheAt`, `LiveRefreshPath`, `TryMarkZaiRefreshAt`/`ClearZaiRefreshAt`), 기존 `collect.SpawnSelfCollect(provider string) error`
- Produces (Task 6 사용):

```go
const (
    PollInterval = 5 * time.Second
    CacheTTL     = 5 * time.Minute
    MarkerTTL    = 60 * time.Second
)
func NeedsRefresh(live collect.LiveSession, ok bool, now time.Time) bool // !ok || FetchedAt+CacheTTL <= now
func TryRefresh(refreshPath string, spawn func() error, now time.Time) bool
```

`TryRefresh` 계약: `TryMarkZaiRefreshAt(refreshPath, MarkerTTL)` 실패(진행 중) → false, 스폰 없음. 성공 → `spawn()` 호출(에러 무시) → true. **spawn 실패 시 marker 제거하지 않는다**(60s 백프레셔). marker 완료 제거는 collect 측(Task 2)이 담당.

- [ ] **Step 1: 실패 테스트 작성** — `TestNeedsRefresh`(!ok → true, TTL 내 → false, TTL 경과 → true), `TestTryRefreshSpawns`(marker 없음 → spawn 1회 호출·true), `TestTryRefreshBackpressure`(marker 존재 → spawn 0회·false, marker mtime이 60s 경과 → 재획득·spawn 1회), `TestTryRefreshSpawnFailKeepsMarker`(spawn이 에러 → marker 파일 존재 유지)
- [ ] **Step 2: 실패 확인** — `go test -race ./internal/tray -run 'NeedsRefresh|TryRefresh' -v`
- [ ] **Step 3: 구현** — 상수 3종·두 함수만 (goroutine 없음 — 루프는 Task 6 진입점이 소유)
- [ ] **Step 4: 통과 확인** — `go test -race ./...`
- [ ] **Step 5: Commit** — `git commit -m "feat(tray): 갱신 조건 판정 및 marker 백프레셔 감시자 추가"`

### Task 6: traycmd 배선 + macOS 실측

**Files:**
- Create: `cmd/statusline/traycmd.go`
- Modify: `cmd/statusline/main.go`(watch 분기 뒤에 tray 분기), `go.mod`(fyne-io/systray)

**Interfaces:**
- Consumes: Task 3(`LoadConfig`/`SaveConfig`, `Tray.Primary`), Task 4(`Title`/`MenuLines`), Task 5(`PollInterval`, `NeedsRefresh`, `TryRefresh`), Task 1(`LiveCachePath`/`LiveRefreshPath`/`ReadLiveCacheAt`), 기존 `collect.SpawnSelfCollect("all")`
- Produces: `runTray(configPath string) int` — main이 `os.Exit(runTray(configPath))`

구조: `systray.Run(onReady, onExit)`를 main goroutine에서 직접 호출. onReady에서 메뉴 등록(provider 항목·메인 전환 토글·지금 갱신·종료) + `time.NewTicker(PollInterval)` goroutine 시작(채널 종료). 각 틱: `ReadLiveCacheAt → NeedsRefresh → TryRefresh(마커, spawn)` 및 타이틀·메뉴 갱신. 메인 전환: `SaveConfig`로 `Tray.Primary` 저장 후 즉시 재렌더. "지금 갱신": `TryRefresh` 재사용(marker 가드 준수). 모든 경로 recover+fail-soft.

- [ ] **Step 1: 의존성 추가** — `go get github.com/fyne-io/systray@latest && go mod tidy` — 빌드 확인 `go build ./...`(cgo·빌드 복잡성 게이트, 실패 시 대안 검토 보고)
- [ ] **Step 2: 배선 구현** — traycmd.go + main.go 분기(`flag.Arg(0) == "tray"`, stdin 읽기 전). systray 관련 유닛 테스트는 작성하지 않는다(헤드리스 불가 — 스펙 테스트 전략)
- [ ] **Step 3: 전체 검증** — `go build -o statusline ./cmd/statusline && go test -race ./... && GOOS=windows go build ./...` 모두 PASS
- [ ] **Step 4: macOS 수동 실측** — `./statusline tray` 실행 → ① 첫 실행 시 `…` 후 수집되어 `z..` 타이틀 전환 ② 드롭다운 provider 상세·리셋 시간 ③ "지금 갱신" 동작 ④ 메인 전환 후 재시작 시 유지 ⑤ 캐시 삭제 후 복구(부재 조건) ⑥ 종료. 결과(관찰·타이틀 스크린샷)를 커밋 메시지에 요약
- [ ] **Step 5: Commit** — `git commit -m "feat(tray): macOS 메뉴바 트레이 위젯 v0.8.1 배선"`
