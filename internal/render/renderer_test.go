package render_test

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"statusline/internal/config"
	"statusline/internal/model"
	"statusline/internal/render"
)

func TestRenderOutput(t *testing.T) {
	st := model.NewUnifiedStatus("claude")
	st.Model = "Sonnet 3.7"
	st.GitBranch = "main"
	st.ContextTokens = 45

	cfg := config.DefaultConfig()
	output := render.Render(st, cfg)

	if output == "" {
		t.Fatalf("expected non-empty rendered string")
	}

	if !strings.Contains(output, "CLAUDE") {
		t.Errorf("expected output to contain engine label 'CLAUDE', got: %q", output)
	}
	if !strings.Contains(output, "Sonnet 3.7") {
		t.Errorf("expected output to contain model 'Sonnet 3.7', got: %q", output)
	}
	if !strings.Contains(output, "main") {
		t.Errorf("expected output to contain git branch 'main', got: %q", output)
	}
	if !strings.Contains(output, "45%") {
		t.Errorf("expected output to contain token percentage '45%%', got: %q", output)
	}
}

func TestRenderGitStatusFixedSpacing(t *testing.T) {
	cfg := config.DefaultConfig()

	// 1. Dirty git repo
	stDirty := model.NewUnifiedStatus("claude")
	stDirty.GitRepo = "repo"
	stDirty.GitBranch = "main"
	stDirty.GitStatus = "*"
	outDirty := render.Render(stDirty, cfg)

	// 2. Clean git repo
	stClean := model.NewUnifiedStatus("claude")
	stClean.GitRepo = "repo"
	stClean.GitBranch = "main"
	stClean.GitStatus = ""
	outClean := render.Render(stClean, cfg)

	plainDirty := ansi.Strip(outDirty)
	plainClean := ansi.Strip(outClean)

	idxDirty := strings.Index(plainDirty, "│")
	idxClean := strings.Index(plainClean, "│")
	if idxDirty == -1 || idxClean == -1 {
		t.Fatalf("expected separator │ in output")
	}
	colDirty := ansi.StringWidth(plainDirty[:idxDirty])
	colClean := ansi.StringWidth(plainClean[:idxClean])
	if colDirty != colClean || colClean != 29 { // 28 chars line1 + 1 space before │
		t.Errorf("expected │ to visually align at column 29, got dirty=%d, clean=%d", colDirty, colClean)
	}

	// Long branch should be truncated to fit Line1Width and keep │ aligned
	stLong := model.NewUnifiedStatus("claude")
	stLong.GitRepo = "repo"
	stLong.GitBranch = "feature/a-very-long-branch-name-12345"
	plainLong := ansi.Strip(render.Render(stLong, cfg))
	idxLong := strings.Index(plainLong, "│")
	if idxLong == -1 {
		t.Fatalf("expected separator │ in long branch output")
	}
	colLong := ansi.StringWidth(plainLong[:idxLong])
	if colLong != colClean {
		t.Errorf("expected long branch │ to visually align at column %d, got %d", colClean, colLong)
	}
	if !strings.Contains(plainLong, "…") {
		t.Errorf("expected long branch to be truncated with …, got: %q", plainLong)
	}

	// 3. Unconstrained mode (Line1Width = 0)
	cfg0 := config.DefaultConfig()
	cfg0.Layout.Line1Width = 0
	plainDirty0 := ansi.Strip(render.Render(stDirty, cfg0))
	plainClean0 := ansi.Strip(render.Render(stClean, cfg0))
	if !strings.Contains(plainDirty0, "repo  main * │") {
		t.Errorf("expected dirty statusline to contain 'repo  main * │', got: %q", plainDirty0)
	}
	if !strings.Contains(plainClean0, "repo  main   │") {
		t.Errorf("expected clean statusline to contain 'repo  main   │', got: %q", plainClean0)
	}

	// 4. Non-git directory (no branch/repo) with Line1Width = 0 should not have phantom space
	stNonGit := model.NewUnifiedStatus("claude")
	stNonGit.Cwd = "/tmp/testdir"
	plainNonGit := ansi.Strip(render.Render(stNonGit, cfg0))
	if strings.Contains(plainNonGit, "testdir  │") {
		t.Errorf("expected non-git statusline not to have phantom space, got: %q", plainNonGit)
	}
}

func TestRenderCapabilityFiltering(t *testing.T) {
	st := model.NewUnifiedStatus("generic")
	st.Model = "GenericModel"
	st.ContextTokens = 80 // Should be ignored because Capabilities.HasTokens is false for generic

	cfg := config.DefaultConfig()
	output := render.Render(st, cfg)

	if strings.Contains(output, "80%") {
		t.Errorf("expected token percentage to be hidden for host without HasTokens capability, got: %q", output)
	}
}

func TestRenderConfigElementsDisabled(t *testing.T) {
	st := model.NewUnifiedStatus("claude")
	st.Model = "Claude 3.7"
	st.GitBranch = "feature-branch"
	st.ContextTokens = 50

	cfg := config.DefaultConfig()
	cfg.Elements.EngineLabel = false
	cfg.Elements.Model = false
	cfg.Elements.GitBranch = false
	cfg.Elements.ShowTokens = false

	output := render.Render(st, cfg)

	if strings.Contains(output, "CLAUDE") {
		t.Errorf("expected engine label to be hidden when EngineLabel=false, got: %q", output)
	}
	if strings.Contains(output, "Claude 3.7") {
		t.Errorf("expected model to be hidden when Model=false, got: %q", output)
	}
	if strings.Contains(output, "feature-branch") {
		t.Errorf("expected git branch to be hidden when GitBranch=false, got: %q", output)
	}
	if strings.Contains(output, "50%") {
		t.Errorf("expected token percentage to be hidden when ShowTokens=false, got: %q", output)
	}
}

func TestRenderEmptyStatus(t *testing.T) {
	st := model.NewUnifiedStatus("generic")
	cfg := config.DefaultConfig()

	output := render.Render(st, cfg)
	if !strings.Contains(output, "GENERIC") {
		t.Errorf("expected engine label 'GENERIC' in empty generic status, got: %q", output)
	}
}

func TestRenderQuotaEnabled(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Quota = []model.QuotaCategory{
		{Name: "gemini", FiveH: 0.93, Weekly: 0.53},
		{Name: "3rd", FiveH: 1.0, Weekly: 0.22},
	}

	cfg := config.DefaultConfig()
	output := render.Render(st, cfg)

	if !strings.Contains(output, "📊") {
		t.Errorf("expected quota emoji '📊' in output, got: %q", output)
	}
	if !strings.Contains(output, "gemini") {
		t.Errorf("expected 'gemini' category label, got: %q", output)
	}
	if !strings.Contains(output, "3rd") {
		t.Errorf("expected '3rd' category label, got: %q", output)
	}
	if !strings.Contains(output, "5h:93%") {
		t.Errorf("expected '5h:93%%' in output, got: %q", output)
	}
	if !strings.Contains(output, "wk:53%") {
		t.Errorf("expected 'wk:53%%' in output, got: %q", output)
	}
	if !strings.Contains(output, "wk:22%") {
		t.Errorf("expected 'wk:22%%' in output, got: %q", output)
	}
}

func TestRenderQuotaWithCountdown(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Quota = []model.QuotaCategory{
		{
			Name:           "gemini",
			FiveH:          0.93,
			Weekly:         0.53,
			FiveHResetsAt:  time.Now().Add(4*time.Hour + 30*time.Minute + 30*time.Second).UnixMilli(),
			WeeklyResetsAt: time.Now().Add(72*time.Hour + 30*time.Second).UnixMilli(),
		},
		{
			Name:          "3rd",
			FiveH:         0.0,
			Weekly:        0.22,
			FiveHResetsAt: time.Now().Add(1*time.Hour + 15*time.Minute + 30*time.Second).UnixMilli(),
		},
	}

	cfg := config.DefaultConfig()
	output := render.Render(st, cfg)

	for _, want := range []string{"5h:93%(4h30m)", "wk:53%(3d0h)", "3rd", "5h:0%(1h15m)", "wk:22%"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got: %q", want, output)
		}
	}
}

func TestRenderQuotaDisabled(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Quota = []model.QuotaCategory{
		{Name: "gemini", FiveH: 0.93, Weekly: 0.53},
	}

	cfg := config.DefaultConfig()
	cfg.Elements.Quota = false
	output := render.Render(st, cfg)

	if strings.Contains(output, "📊") {
		t.Errorf("expected quota to be hidden when Quota=false, got: %q", output)
	}
}

func TestRenderQuotaCapabilityFiltering(t *testing.T) {
	st := model.NewUnifiedStatus("claude") // claude에는 HasQuota=false
	st.Quota = []model.QuotaCategory{
		{Name: "gemini", FiveH: 0.93, Weekly: 0.53},
	}

	cfg := config.DefaultConfig()
	output := render.Render(st, cfg)

	if strings.Contains(output, "📊") {
		t.Errorf("expected quota to be hidden for engine without HasQuota capability, got: %q", output)
	}
}

func TestRenderModelFormat(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Model = "Gemini 3.6 Flash (Medium)"

	cfg := config.DefaultConfig()
	cfg.Elements.ModelFormat = "short"
	output := render.Render(st, cfg)
	if !strings.Contains(output, "flash-med") {
		t.Errorf("expected short model name 'flash-med', got: %q", output)
	}

	cfg.Elements.ModelFormat = "full"
	outputFull := render.Render(st, cfg)
	if !strings.Contains(outputFull, "Gemini 3.6 Flash (Medium)") {
		t.Errorf("expected full model name, got: %q", outputFull)
	}
}

func TestRenderPermissionPending(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Permission = "pending"

	cfg := config.DefaultConfig()
	cfg.Elements.Permission = true
	output := render.Render(st, cfg)
	if !strings.Contains(output, "WAITING") && !strings.Contains(output, "승인 대기") {
		t.Errorf("expected permission pending indicator, got: %q", output)
	}

	cfg.Elements.Permission = false
	outputDisabled := render.Render(st, cfg)
	if strings.Contains(outputDisabled, "WAITING") || strings.Contains(outputDisabled, "승인 대기") {
		t.Errorf("expected permission indicator to be hidden when disabled, got: %q", outputDisabled)
	}
}

func TestRenderThemes(t *testing.T) {
	st := model.NewUnifiedStatus("claude")
	st.Model = "Claude 3.7"

	themes := []string{"sleek_dark", "light", "nord", "monokai-dark"}
	for _, theme := range themes {
		t.Run(theme, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Theme = theme
			output := render.Render(st, cfg)
			if output == "" {
				t.Fatalf("expected non-empty output for theme %s", theme)
			}
			if !strings.Contains(output, "CLAUDE") {
				t.Errorf("expected CLAUDE in theme %s, got: %q", theme, output)
			}
		})
	}
}

func TestRenderWrapModeTruncate(t *testing.T) {
	st := model.NewUnifiedStatus("antigravity")
	st.Model = "Gemini 3.6 Flash (Medium)"
	st.GitBranch = "feature/very-very-long-branch-name-that-exceeds-terminal-width"
	st.TerminalWidth = 30 // 매우 좁은 너비 지정

	cfg := config.DefaultConfig()
	cfg.WrapMode = "truncate"
	output := render.Render(st, cfg)

	if output == "" {
		t.Fatalf("expected non-empty output")
	}
	// 터미널 줄바꿈 방지: 출력에 개행(\n)이 없어야 함
	if strings.Contains(output, "\n") {
		t.Errorf("expected single line output without newlines, got: %q", output)
	}
}
