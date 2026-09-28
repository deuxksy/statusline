package render

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"statusline/internal/config"
	"statusline/internal/model"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

type ThemePalette struct {
	BadgeBg    lipgloss.Color
	BadgeFg    lipgloss.Color
	BranchFg   lipgloss.Color
	RepoFg     lipgloss.Color
	StatusFg   lipgloss.Color
	CwdFg      lipgloss.Color
	HostnameFg lipgloss.Color
	ModelFg    lipgloss.Color
	ThinkingFg lipgloss.Color
	SkillsFg   lipgloss.Color
	ToolFg     lipgloss.Color
	QuotaLabel lipgloss.Color
	QuotaGood  lipgloss.Color
	QuotaWarn  lipgloss.Color
	QuotaBad   lipgloss.Color
	WaitBg     lipgloss.Color
	WaitFg     lipgloss.Color
}

var themes = map[string]ThemePalette{
	"sleek_dark": {
		BadgeBg:    lipgloss.Color("#6366F1"),
		BadgeFg:    lipgloss.Color("#FFFFFF"),
		BranchFg:   lipgloss.Color("#10B981"),
		RepoFg:     lipgloss.Color("#06B6D4"),
		StatusFg:   lipgloss.Color("#F59E0B"),
		CwdFg:      lipgloss.Color("#64748B"),
		HostnameFg: lipgloss.Color("#94A3B8"),
		ModelFg:    lipgloss.Color("#3B82F6"),
		ThinkingFg: lipgloss.Color("#A855F7"),
		SkillsFg:   lipgloss.Color("#EC4899"),
		ToolFg:     lipgloss.Color("#EAB308"),
		QuotaLabel: lipgloss.Color("#8B5CF6"),
		QuotaGood:  lipgloss.Color("#10B981"),
		QuotaWarn:  lipgloss.Color("#F59E0B"),
		QuotaBad:   lipgloss.Color("#EF4444"),
		WaitBg:     lipgloss.Color("#EF4444"),
		WaitFg:     lipgloss.Color("#FFFFFF"),
	},
	"light": {
		BadgeBg:    lipgloss.Color("#4F46E5"),
		BadgeFg:    lipgloss.Color("#FFFFFF"),
		BranchFg:   lipgloss.Color("#047857"),
		RepoFg:     lipgloss.Color("#0891B2"),
		StatusFg:   lipgloss.Color("#D97706"),
		CwdFg:      lipgloss.Color("#475569"),
		HostnameFg: lipgloss.Color("#64748B"),
		ModelFg:    lipgloss.Color("#1D4ED8"),
		ThinkingFg: lipgloss.Color("#7E22CE"),
		SkillsFg:   lipgloss.Color("#BE185D"),
		ToolFg:     lipgloss.Color("#B45309"),
		QuotaLabel: lipgloss.Color("#6D28D9"),
		QuotaGood:  lipgloss.Color("#047857"),
		QuotaWarn:  lipgloss.Color("#D97706"),
		QuotaBad:   lipgloss.Color("#B91C1C"),
		WaitBg:     lipgloss.Color("#DC2626"),
		WaitFg:     lipgloss.Color("#FFFFFF"),
	},
	"nord": {
		BadgeBg:    lipgloss.Color("#5E81AC"),
		BadgeFg:    lipgloss.Color("#ECEFF4"),
		BranchFg:   lipgloss.Color("#A3BE8C"),
		RepoFg:     lipgloss.Color("#88C0D0"),
		StatusFg:   lipgloss.Color("#EBCB8B"),
		CwdFg:      lipgloss.Color("#D8DEE9"),
		HostnameFg: lipgloss.Color("#4C566A"),
		ModelFg:    lipgloss.Color("#81A1C1"),
		ThinkingFg: lipgloss.Color("#B48EAD"),
		SkillsFg:   lipgloss.Color("#D08770"),
		ToolFg:     lipgloss.Color("#EBCB8B"),
		QuotaLabel: lipgloss.Color("#B48EAD"),
		QuotaGood:  lipgloss.Color("#A3BE8C"),
		QuotaWarn:  lipgloss.Color("#EBCB8B"),
		QuotaBad:   lipgloss.Color("#BF616A"),
		WaitBg:     lipgloss.Color("#BF616A"),
		WaitFg:     lipgloss.Color("#ECEFF4"),
	},
	"monokai-dark": {
		BadgeBg:    lipgloss.Color("#FD971F"),
		BadgeFg:    lipgloss.Color("#272822"),
		BranchFg:   lipgloss.Color("#A6E22E"),
		RepoFg:     lipgloss.Color("#66D9EF"),
		StatusFg:   lipgloss.Color("#F92672"),
		CwdFg:      lipgloss.Color("#75715E"),
		HostnameFg: lipgloss.Color("#F8F8F2"),
		ModelFg:    lipgloss.Color("#66D9EF"),
		ThinkingFg: lipgloss.Color("#AE81FF"),
		SkillsFg:   lipgloss.Color("#F92672"),
		ToolFg:     lipgloss.Color("#E6DB74"),
		QuotaLabel: lipgloss.Color("#AE81FF"),
		QuotaGood:  lipgloss.Color("#A6E22E"),
		QuotaWarn:  lipgloss.Color("#FD971F"),
		QuotaBad:   lipgloss.Color("#F92672"),
		WaitBg:     lipgloss.Color("#F92672"),
		WaitFg:     lipgloss.Color("#F8F8F2"),
	},
}

func getPalette(name string) ThemePalette {
	if p, ok := themes[name]; ok {
		return p
	}
	return themes["sleek_dark"]
}

func formatModelName(name, format string) string {
	if format != "short" {
		return name
	}
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "flash") && strings.Contains(lower, "med"):
		return "flash-med"
	case strings.Contains(lower, "flash"):
		return "flash"
	case strings.Contains(lower, "pro"):
		return "gemini-pro"
	case strings.Contains(lower, "sonnet"):
		if strings.Contains(lower, "3.7") || strings.Contains(lower, "3-7") {
			return "sonnet-3.7"
		}
		return "sonnet"
	case strings.Contains(lower, "opus"):
		return "opus"
	case strings.Contains(lower, "haiku"):
		if strings.Contains(lower, "3.5") || strings.Contains(lower, "3-5") {
			return "haiku-3.5"
		}
		return "haiku"
	default:
		if len(name) > 12 {
			return name[:10] + ".."
		}
		return name
	}
}

func renderSegment(item string, st *model.UnifiedStatus, cfg *config.Config, p ThemePalette) string {
	switch item {
	case "permission":
		if cfg.Elements.Permission && st.Capabilities.HasPermission && (st.Permission == "pending" || st.Permission == "waiting") {
			return lipgloss.NewStyle().Bold(true).Foreground(p.WaitFg).Background(p.WaitBg).Padding(0, 1).Render("🔒 WAITING")
		}
	case "hostname":
		if cfg.Elements.Hostname && st.Hostname != "" {
			return lipgloss.NewStyle().Foreground(p.HostnameFg).Render("@" + st.Hostname)
		}
	case "cwd":
		if cfg.Elements.Cwd && st.Cwd != "" {
			folder := filepath.Base(st.Cwd)
			// Avoid duplicating cwd folder name if gitRepo is enabled and matches
			if cfg.Elements.GitRepo && st.GitRepo != "" && folder == st.GitRepo {
				return ""
			}
			val := st.Cwd
			if cfg.Elements.CwdFormat == "folder" {
				val = folder
			}
			return lipgloss.NewStyle().Foreground(p.CwdFg).Render(val)
		}
	case "gitRepo":
		if cfg.Elements.GitRepo && st.GitRepo != "" {
			return lipgloss.NewStyle().Foreground(p.RepoFg).Render(st.GitRepo)
		}
	case "gitBranch":
		if cfg.Elements.GitBranch && st.GitBranch != "" {
			return lipgloss.NewStyle().Foreground(p.BranchFg).Render(" " + st.GitBranch)
		}
	case "gitStatus":
		if cfg.Elements.GitStatus {
			if st.GitStatus != "" {
				return lipgloss.NewStyle().Foreground(p.StatusFg).Bold(true).Render(st.GitStatus)
			}
			if st.GitBranch != "" {
				return " "
			}
		}
	case "engineLabel":
		if cfg.Elements.EngineLabel && st.EngineName != "" {
			label := strings.ToUpper(st.EngineName)
			return lipgloss.NewStyle().Bold(true).Foreground(p.BadgeFg).Background(p.BadgeBg).Padding(0, 1).Render(label)
		}
	case "model":
		if cfg.Elements.Model && st.Model != "" {
			displayName := formatModelName(st.Model, cfg.Elements.ModelFormat)
			return lipgloss.NewStyle().Foreground(p.ModelFg).Bold(true).Render(displayName)
		}
	case "thinking":
		if cfg.Elements.Thinking && st.Capabilities.HasThinking && st.ThinkingState != "" {
			return lipgloss.NewStyle().Foreground(p.ThinkingFg).Render("🧠 " + st.ThinkingState)
		}
	case "contextBar":
		if cfg.Elements.ContextBar && cfg.Elements.ShowTokens && st.Capabilities.HasTokens && st.ContextTokens > 0 {
			if st.ContextLimit > 0 {
				pct := int(math.Round(float64(st.ContextTokens) / float64(st.ContextLimit) * 100))
				text := fmt.Sprintf("%s/%s(%d%%)", humanizeTokens(st.ContextTokens), humanizeTokens(st.ContextLimit), pct)
				return "⚡ " + contextColor(pct, cfg, p).Render(text)
			}
			// ContextLimit 미제공(OMC/antigravity percent 경로) 폴백
			return fmt.Sprintf("[%d%%]", st.ContextTokens)
		}
	case "tokens":
		if cfg.Elements.ShowTokens && !cfg.Elements.ContextBar && st.Capabilities.HasTokens && st.ContextTokens > 0 {
			return fmt.Sprintf("%d tokens", st.ContextTokens)
		}
	case "activeSkills":
		if cfg.Elements.ActiveSkills && st.Capabilities.HasSkills && len(st.ActiveSkills) > 0 {
			return lipgloss.NewStyle().Foreground(p.SkillsFg).Render("⚡ " + strings.Join(st.ActiveSkills, ","))
		}
	case "lastTool":
		if cfg.Elements.LastTool && st.Capabilities.HasTools && st.LastTool != "" {
			return lipgloss.NewStyle().Foreground(p.ToolFg).Render("🔧 " + st.LastTool)
		}
	case "quota":
		if cfg.Elements.Quota && st.Capabilities.HasQuota && len(st.Quota) > 0 {
			if st.Provider == "zai" || st.Provider == "zhipu" {
				return renderZaiQuota(st.Quota, p)
			}
			return renderQuota(st.Quota, p)
		}
	}
	return ""
}

// humanizeTokens — 토큰수 축약: 57340 → "57k", 200000 → "200k", 1000000 → "1M", 1234567 → "1.2M"
func humanizeTokens(n int) string {
	switch {
	case n >= 1_000_000:
		m := float64(n) / 1_000_000
		if m == math.Trunc(m) {
			return fmt.Sprintf("%dM", int(m))
		}
		return fmt.Sprintf("%.1fM", m)
	case n >= 1_000:
		return fmt.Sprintf("%dk", n/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// contextColor — 컨텍스트 사용률(%)에 따라 색상 스타일 반환
func contextColor(pct int, cfg *config.Config, p ThemePalette) lipgloss.Style {
	if pct >= cfg.Thresholds.ContextCritical {
		return lipgloss.NewStyle().Foreground(p.QuotaBad)
	}
	if pct >= cfg.Thresholds.ContextWarning {
		return lipgloss.NewStyle().Foreground(p.QuotaWarn)
	}
	return lipgloss.NewStyle().Foreground(p.QuotaGood)
}

// quotaColor — 잔여 비율에 따라 색상 스타일 반환
func quotaColor(fraction float64, p ThemePalette) lipgloss.Style {
	pct := fraction * 100
	if pct > 50 {
		return lipgloss.NewStyle().Foreground(p.QuotaGood)
	} else if pct >= 20 {
		return lipgloss.NewStyle().Foreground(p.QuotaWarn)
	}
	return lipgloss.NewStyle().Foreground(p.QuotaBad)
}

// renderQuota — 카테고리별 5h/wk 렌더링
func renderQuota(categories []model.QuotaCategory, p ThemePalette) string {
	labelStyle := lipgloss.NewStyle().Foreground(p.QuotaLabel).Bold(true)
	now := time.Now()
	var parts []string
	for _, cat := range categories {
		label := labelStyle.Render(cat.Name)
		fiveHStr := fmt.Sprintf("5h:%d%%", int(math.Round(cat.FiveH*100)))
		if cd := formatResetCountdown(cat.FiveHResetsAt, now); cd != "" {
			fiveHStr += "(" + cd + ")"
		}
		weeklyStr := fmt.Sprintf("wk:%d%%", int(math.Round(cat.Weekly*100)))
		if cd := formatResetCountdown(cat.WeeklyResetsAt, now); cd != "" {
			weeklyStr += "(" + cd + ")"
		}
		fiveH := quotaColor(cat.FiveH, p).Render(fiveHStr)
		weekly := quotaColor(cat.Weekly, p).Render(weeklyStr)
		parts = append(parts, fmt.Sprintf("%s %s %s", label, fiveH, weekly))
	}
	return "📊 " + strings.Join(parts, " │ ")
}

func Render(st *model.UnifiedStatus, cfg *config.Config) string {
	if st == nil || cfg == nil {
		return ""
	}

	palette := getPalette(cfg.Theme)

	var line1Segments []string
	for _, item := range cfg.Layout.Line1 {
		if seg := renderSegment(item, st, cfg, palette); seg != "" {
			line1Segments = append(line1Segments, seg)
		}
	}

	var mainSegments []string
	for _, item := range cfg.Layout.Main {
		if seg := renderSegment(item, st, cfg, palette); seg != "" {
			mainSegments = append(mainSegments, seg)
		}
	}

	// Fallback if layout iteration yields nothing but we have engine label
	if len(mainSegments) == 0 && cfg.Elements.EngineLabel && st.EngineName != "" {
		badge := lipgloss.NewStyle().Bold(true).Foreground(palette.BadgeFg).Background(palette.BadgeBg).Padding(0, 1)
		mainSegments = append(mainSegments, badge.Render(strings.ToUpper(st.EngineName)))
	}

	line1 := strings.Join(line1Segments, " ")
	if cfg.Layout.Line1Width > 0 && line1 != "" {
		curWidth := ansi.StringWidth(line1)
		if curWidth < cfg.Layout.Line1Width {
			line1 = line1 + strings.Repeat(" ", cfg.Layout.Line1Width-curWidth)
		} else if curWidth > cfg.Layout.Line1Width {
			line1 = ansi.Truncate(line1, cfg.Layout.Line1Width, "…")
		}
	}
	main := strings.Join(mainSegments, " │ ")

	result := ""
	if line1 != "" {
		if main != "" {
			result = line1 + " │ " + main
		} else {
			result = line1
		}
	} else {
		result = main
	}

	// wrapMode 처리 (truncate)
	if cfg.WrapMode == "truncate" && st.TerminalWidth > 0 {
		w := ansi.StringWidth(result)
		if w > st.TerminalWidth {
			result = ansi.Truncate(result, st.TerminalWidth, "…")
		}
	}

	return result
}

// formatResetCountdown — Unix(ms) 리셋 시각을 "18d19h"/"4h46m" 카운트다운으로. 경과·무효는 ""
func formatResetCountdown(unixMs int64, now time.Time) string {
	if unixMs <= 0 {
		return ""
	}
	d := time.UnixMilli(unixMs).Sub(now)
	if d <= 0 {
		return ""
	}
	totalHours := int(d.Hours())
	if days := totalHours / 24; days > 0 {
		return fmt.Sprintf("%dd%dh", days, totalHours%24)
	}
	return fmt.Sprintf("%dh%dm", totalHours, int(d.Minutes())%60)
}

// formatZaiQuota — Z.AI quota 순수 텍스트 형태: "zai 5h:97%(4h46m) mo:66%(18d19h)"
// 카운트다운은 reset 값이 있을 때만 붙는다. 0%도 유효값이다.
func formatZaiQuota(cats []model.QuotaCategory, now time.Time) string {
	if len(cats) == 0 {
		return ""
	}
	cat := cats[0]
	fiveH := fmt.Sprintf("5h:%d%%", int(math.Round(cat.FiveH*100)))
	if cd := formatResetCountdown(cat.FiveHResetsAt, now); cd != "" {
		fiveH += "(" + cd + ")"
	}
	mcp := fmt.Sprintf("mo:%d%%", int(math.Round(cat.Monthly*100)))
	if cd := formatResetCountdown(cat.MonthlyResetsAt, now); cd != "" {
		mcp += "(" + cd + ")"
	}
	return cat.Name + " " + fiveH + " " + mcp
}

// renderZaiQuota — formatZaiQuota의 스타일 버전 (📊 prefix, 라벨·비율별 색상)
func renderZaiQuota(cats []model.QuotaCategory, p ThemePalette) string {
	if len(cats) == 0 {
		return ""
	}
	now := time.Now()
	cat := cats[0]
	label := lipgloss.NewStyle().Foreground(p.QuotaLabel).Bold(true).Render(cat.Name)
	fiveH := fmt.Sprintf("5h:%d%%", int(math.Round(cat.FiveH*100)))
	if cd := formatResetCountdown(cat.FiveHResetsAt, now); cd != "" {
		fiveH += "(" + cd + ")"
	}
	mcp := fmt.Sprintf("mo:%d%%", int(math.Round(cat.Monthly*100)))
	if cd := formatResetCountdown(cat.MonthlyResetsAt, now); cd != "" {
		mcp += "(" + cd + ")"
	}
	fiveHPart := quotaColor(cat.FiveH, p).Render(fiveH)
	mcpPart := quotaColor(cat.Monthly, p).Render(mcp)
	return "📊 " + label + " " + fiveHPart + " " + mcpPart
}
