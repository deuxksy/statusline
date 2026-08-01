package render

import (
	"fmt"
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

	modelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)
)

func Render(st *model.UnifiedStatus, cfg *config.Config) string {
	if st == nil || cfg == nil {
		return ""
	}

	var line1Segments []string
	var mainSegments []string

	// Engine Badge
	if cfg.Elements.EngineLabel && st.EngineName != "" {
		label := strings.ToUpper(st.EngineName)
		mainSegments = append(mainSegments, badgeStyle.Render(label))
	}

	// Model
	if cfg.Elements.Model && st.Model != "" {
		mainSegments = append(mainSegments, modelStyle.Render(st.Model))
	}

	// Git Branch
	if cfg.Elements.GitBranch && st.GitBranch != "" {
		line1Segments = append(line1Segments, branchStyle.Render(" "+st.GitBranch))
	}

	// Tokens
	if cfg.Elements.ShowTokens && st.Capabilities.HasTokens && st.ContextTokens > 0 {
		mainSegments = append(mainSegments, fmt.Sprintf("[%d%%]", st.ContextTokens))
	}

	line1 := strings.Join(line1Segments, " ")
	main := strings.Join(mainSegments, " │ ")

	if line1 != "" {
		return line1 + "\n" + main
	}
	return main
}
