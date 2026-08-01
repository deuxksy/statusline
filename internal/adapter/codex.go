package adapter

import (
	"statusline/internal/model"
)

type CodexAdapter struct{}

func (c *CodexAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("codex")
	st.Model = "codex"
	return st, nil
}
