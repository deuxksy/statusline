package render

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"statusline/internal/config"
	"statusline/internal/model"
)

var (
	badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6366F1")).
			Padding(0, 1)

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	repoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#06B6D4"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	cwdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	hostnameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8"))

	modelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)

	thinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A855F7"))

	skillsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EC4899"))

	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAB308"))
)

func renderSegment(item string, st *model.UnifiedStatus, cfg *config.Config) string {
	switch item {
	case "hostname":
		if cfg.Elements.Hostname && st.Hostname != "" {
			return hostnameStyle.Render("@" + st.Hostname)
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
			return cwdStyle.Render(val)
		}
	case "gitRepo":
		if cfg.Elements.GitRepo && st.GitRepo != "" {
			return repoStyle.Render(st.GitRepo)
		}
	case "gitBranch":
		if cfg.Elements.GitBranch && st.GitBranch != "" {
			return branchStyle.Render(" " + st.GitBranch)
		}
	case "gitStatus":
		if cfg.Elements.GitStatus && st.GitStatus != "" {
			return statusStyle.Render(st.GitStatus)
		}
	case "engineLabel":
		if cfg.Elements.EngineLabel && st.EngineName != "" {
			label := strings.ToUpper(st.EngineName)
			return badgeStyle.Render(label)
		}
	case "model":
		if cfg.Elements.Model && st.Model != "" {
			return modelStyle.Render(st.Model)
		}
	case "thinking":
		if cfg.Elements.Thinking && st.Capabilities.HasThinking && st.ThinkingState != "" {
			return thinkingStyle.Render("🧠 " + st.ThinkingState)
		}
	case "contextBar":
		if cfg.Elements.ContextBar && cfg.Elements.ShowTokens && st.Capabilities.HasTokens && st.ContextTokens > 0 {
			return fmt.Sprintf("[%d%%]", st.ContextTokens)
		}
	case "tokens":
		if cfg.Elements.ShowTokens && !cfg.Elements.ContextBar && st.Capabilities.HasTokens && st.ContextTokens > 0 {
			return fmt.Sprintf("%d tokens", st.ContextTokens)
		}
	case "activeSkills":
		if cfg.Elements.ActiveSkills && st.Capabilities.HasSkills && len(st.ActiveSkills) > 0 {
			return skillsStyle.Render("⚡ " + strings.Join(st.ActiveSkills, ","))
		}
	case "lastTool":
		if cfg.Elements.LastTool && st.Capabilities.HasTools && st.LastTool != "" {
			return toolStyle.Render("🔧 " + st.LastTool)
		}
	}
	return ""
}

func Render(st *model.UnifiedStatus, cfg *config.Config) string {
	if st == nil || cfg == nil {
		return ""
	}

	var line1Segments []string
	for _, item := range cfg.Layout.Line1 {
		if seg := renderSegment(item, st, cfg); seg != "" {
			line1Segments = append(line1Segments, seg)
		}
	}

	var mainSegments []string
	for _, item := range cfg.Layout.Main {
		if seg := renderSegment(item, st, cfg); seg != "" {
			mainSegments = append(mainSegments, seg)
		}
	}

	// Fallback if layout iteration yields nothing but we have engine label
	if len(mainSegments) == 0 && cfg.Elements.EngineLabel && st.EngineName != "" {
		mainSegments = append(mainSegments, badgeStyle.Render(strings.ToUpper(st.EngineName)))
	}

	line1 := strings.Join(line1Segments, " ")
	main := strings.Join(mainSegments, " │ ")

	if line1 != "" {
		if main != "" {
			return line1 + " │ " + main
		}
		return line1
	}
	return main
}
