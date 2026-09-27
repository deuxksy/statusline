package main

import (
	"strings"
	"sync"
	"time"

	"fyne.io/systray"
	"statusline/internal/collect"
	"statusline/internal/config"
	"statusline/internal/tray"
)

// runTray — macOS 메뉴바(v0.8.1)/Windows 트레이(v0.8.5) 위젯 진입점.
// systray.Run은 main goroutine에서 직접 호출해야 한다 (AppKit 런루프 제약 — 스펙).
// 모든 경로 fail-soft: 어떤 캐시·I/O 실패로도 크래시하지 않는다.
func runTray(configPath string) int {
	cfg := config.LoadConfig(configPath)
	cachePath, err := collect.LiveCachePath()
	if err != nil {
		return 1
	}

	app := &trayApp{
		configPath:  configPath,
		cfg:         cfg,
		cachePath:   cachePath,
		refreshPath: collect.LiveRefreshPath(cachePath),
	}
	systray.Run(app.onReady, app.onExit)
	return 0
}

// trayApp — 메뉴 슬롯은 고정하고 Title을 갱신한다 (부재 provider는 "데이터 없음" 플레이스홀더).
type trayApp struct {
	configPath  string
	cfg         *config.Config
	cachePath   string
	refreshPath string

	mu     sync.Mutex // cfg.Tray.Primary 보호 (클릭 goroutine ↔ poll)
	pollMu sync.Mutex // poll 직렬화 — ticker·setPrimary poll이 겹칠 때 MenuItem.SetTitle 동시 호출 방지(systray item 필드는 lock 없는 write)
	items  map[string]*systray.MenuItem
}

func (a *trayApp) onReady() {
	a.items = map[string]*systray.MenuItem{
		"zai5h":   systray.AddMenuItem("zai 5h —", "zai 5h quota"),
		"zaiMcp":  systray.AddMenuItem("zai MCP —", "zai monthly MCP quota"),
		"chatgpt": systray.AddMenuItem("chatgpt —", "chatgpt rate limits"),
	}
	systray.AddSeparator()
	a.items["mainZai"] = systray.AddMenuItem("○ zai (main)", "메인 provider를 zai로")
	a.items["mainChatgpt"] = systray.AddMenuItem("○ chatgpt (main)", "메인 provider를 chatgpt로")
	systray.AddSeparator()
	a.items["refresh"] = systray.AddMenuItem("지금 갱신", "즉시 collect 스폰")
	a.items["quit"] = systray.AddMenuItem("종료", "트레이 종료")

	go a.watchClicks()
	a.poll()
	go a.loop()
}

func (a *trayApp) onExit() {}

func (a *trayApp) loop() {
	tick := time.NewTicker(tray.PollInterval)
	defer tick.Stop()
	for range tick.C {
		a.poll()
	}
}

func (a *trayApp) watchClicks() {
	for {
		select {
		case <-a.items["mainZai"].ClickedCh:
			a.setPrimary("zai")
		case <-a.items["mainChatgpt"].ClickedCh:
			a.setPrimary("chatgpt")
		case <-a.items["refresh"].ClickedCh:
			a.spawnCollect() // 지금 갱신 — TTL 조건 없이 즉시 (marker 가드는 준수)
		case <-a.items["quit"].ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (a *trayApp) setPrimary(primary string) {
	a.mu.Lock()
	a.cfg.Tray.Primary = primary
	a.mu.Unlock()
	_ = config.SaveConfig(a.configPath, a.cfg) // 실패 시 메모리 상태로 동작 (스펙: 재시도 없음)
	a.poll()
}

func (a *trayApp) spawnCollect() {
	tray.TryRefresh(a.refreshPath, func() error {
		return collect.SpawnSelfCollect("all")
	})
}

// poll — 캐시 읽기 → 갱신 판정·스폰 → 타이틀·메뉴 갱신. recover로 fail-soft 유지.
// pollMu로 직렬화 — systray MenuItem.SetTitle은 내부 필드를 lock 없이 쓴다(v1.12.2).
func (a *trayApp) poll() {
	a.pollMu.Lock()
	defer a.pollMu.Unlock()
	defer func() { _ = recover() }() // 트레이는 어떤 입력·I/O 실패로도 크래시하지 않는다

	now := time.Now()
	live, ok := collect.ReadLiveCacheAt(a.cachePath)
	if tray.NeedsRefresh(live, ok, now) {
		a.spawnCollect()
	}

	a.mu.Lock()
	primary := a.cfg.Tray.Primary
	a.mu.Unlock()

	title := tray.Title(primary, live)
	systray.SetTitle(title)   // macOS 메뉴바 상시 텍스트
	systray.SetTooltip(title) // Windows 트레이(v0.8.5) 폴백 표시

	if a.items == nil {
		return
	}
	a.items["zai5h"].SetTitle("zai 5h " + tray.ZaiQuotaLine(live.Zai, now, true))
	a.items["zaiMcp"].SetTitle("zai MCP " + tray.ZaiQuotaLine(live.Zai, now, false))
	if _, hasRate := live.Codex["rate_limits"]; hasRate {
		a.items["chatgpt"].SetTitle("chatgpt rate " + strings.TrimPrefix(tray.ChatGPTTitle(live.Codex), "c"))
	} else {
		a.items["chatgpt"].SetTitle("chatgpt 데이터 없음")
	}
	zaiMark, chatMark := "○", "○"
	if primary != "chatgpt" {
		zaiMark = "✓"
	} else {
		chatMark = "✓"
	}
	a.items["mainZai"].SetTitle(zaiMark + " zai (main)")
	a.items["mainChatgpt"].SetTitle(chatMark + " chatgpt (main)")
}
