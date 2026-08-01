package render_test

import (
	"strings"
	"testing"

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
