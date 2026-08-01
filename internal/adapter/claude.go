package adapter

import (
	"encoding/json"
	"statusline/internal/model"
)

type ClaudeAdapter struct{}

func (c *ClaudeAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("claude")
	var raw struct {
		Model struct {
			DisplayName string `json:"displayName"`
		} `json:"model"`
		ContextBar struct {
			Percentage int `json:"percentage"`
		} `json:"contextBar"`
		Thinking struct {
			State string `json:"state"`
		} `json:"thinking"`
	}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &raw)
		st.Model = raw.Model.DisplayName
		st.ContextTokens = raw.ContextBar.Percentage
		st.ThinkingState = raw.Thinking.State
	}
	return st, nil
}
