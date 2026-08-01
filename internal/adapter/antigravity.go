package adapter

import (
	"encoding/json"
	"statusline/internal/model"
)

type AntigravityAdapter struct{}

func (a *AntigravityAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("antigravity")
	var raw struct {
		Model string `json:"model"`
	}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &raw)
		st.Model = raw.Model
	}
	return st, nil
}
