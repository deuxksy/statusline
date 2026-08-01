package adapter

import (
	"encoding/json"
	"statusline/internal/model"
)

type ClaudeAdapter struct{}

func (c *ClaudeAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("claude")
	if len(input) == 0 {
		return st, nil
	}

	var raw struct {
		Model      json.RawMessage `json:"model"`
		ContextBar struct {
			Percentage int `json:"percentage"`
		} `json:"contextBar"`
		ContextTokens int `json:"contextTokens"`
		Thinking      struct {
			State string `json:"state"`
		} `json:"thinking"`
		ThinkingState string   `json:"thinkingState"`
		ActiveSkills  []string `json:"activeSkills"`
		LastTool      string   `json:"lastTool"`
		Cwd           string   `json:"cwd"`
	}

	if err := json.Unmarshal(input, &raw); err == nil {
		if len(raw.Model) > 0 {
			var modelStr string
			if err := json.Unmarshal(raw.Model, &modelStr); err == nil {
				st.Model = modelStr
			} else {
				var modelObj struct {
					DisplayName string `json:"displayName"`
					Name        string `json:"name"`
				}
				if err := json.Unmarshal(raw.Model, &modelObj); err == nil {
					if modelObj.DisplayName != "" {
						st.Model = modelObj.DisplayName
					} else {
						st.Model = modelObj.Name
					}
				}
			}
		}
		if raw.ContextTokens > 0 {
			st.ContextTokens = raw.ContextTokens
		} else if raw.ContextBar.Percentage > 0 {
			st.ContextTokens = raw.ContextBar.Percentage
		}
		if raw.ThinkingState != "" {
			st.ThinkingState = raw.ThinkingState
		} else if raw.Thinking.State != "" {
			st.ThinkingState = raw.Thinking.State
		}
		if len(raw.ActiveSkills) > 0 {
			st.ActiveSkills = raw.ActiveSkills
		}
		if raw.LastTool != "" {
			st.LastTool = raw.LastTool
		}
		if raw.Cwd != "" {
			st.Cwd = raw.Cwd
		}
	}
	return st, nil
}
