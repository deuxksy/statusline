package model_test

import (
	"testing"

	"statusline/internal/model"
)

func TestUnifiedStatusDefaultsGeneric(t *testing.T) {
	status := model.NewUnifiedStatus("generic")
	if status.EngineName != "generic" {
		t.Errorf("expected engine generic, got %s", status.EngineName)
	}
	if status.Capabilities.HasTokens != false {
		t.Errorf("expected HasTokens false by default for generic")
	}
	if status.Capabilities.HasSkills != false {
		t.Errorf("expected HasSkills false by default for generic")
	}
	if status.Capabilities.HasTools != false {
		t.Errorf("expected HasTools false by default for generic")
	}
	if status.Capabilities.HasThinking != false {
		t.Errorf("expected HasThinking false by default for generic")
	}
	if status.Capabilities.HasPermission != false {
		t.Errorf("expected HasPermission false by default for generic")
	}
}

func TestUnifiedStatusCapabilitiesClaudeAndAntigravity(t *testing.T) {
	engines := []string{"claude", "antigravity"}
	for _, engine := range engines {
		t.Run(engine, func(t *testing.T) {
			status := model.NewUnifiedStatus(engine)
			if status.EngineName != engine {
				t.Errorf("expected engine %s, got %s", engine, status.EngineName)
			}
			if !status.Capabilities.HasTokens {
				t.Errorf("expected HasTokens true for %s", engine)
			}
			if !status.Capabilities.HasSkills {
				t.Errorf("expected HasSkills true for %s", engine)
			}
			if !status.Capabilities.HasTools {
				t.Errorf("expected HasTools true for %s", engine)
			}
			if !status.Capabilities.HasThinking {
				t.Errorf("expected HasThinking true for %s", engine)
			}
			if !status.Capabilities.HasPermission {
				t.Errorf("expected HasPermission true for %s", engine)
			}
		})
	}
}
