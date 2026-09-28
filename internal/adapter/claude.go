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
		ContextWindow struct {
			TotalInputTokens  int `json:"total_input_tokens"`
			ContextWindowSize int `json:"context_window_size"`
		} `json:"context_window"`
		Thinking struct {
			State string `json:"state"`
		} `json:"thinking"`
		ThinkingState string   `json:"thinkingState"`
		ActiveSkills  []string `json:"activeSkills"`
		LastTool      string   `json:"lastTool"`
		Cwd           string   `json:"cwd"`
		Permission    string   `json:"permission"`
		TerminalWidth int      `json:"terminal_width"`
	}

	if err := json.Unmarshal(input, &raw); err == nil {
		if len(raw.Model) > 0 {
			var modelStr string
			if err := json.Unmarshal(raw.Model, &modelStr); err == nil {
				st.Model = modelStr
			} else {
				var modelObj struct {
					DisplayName string `json:"display_name"` // CC 2.x 표준 stdin 필드
					OmcDisplay  string `json:"displayName"`  // OMC 래퍼 폴백
					Name        string `json:"name"`         // OMC 래퍼 폴백
					ID          string `json:"id"`
				}
				if err := json.Unmarshal(raw.Model, &modelObj); err == nil {
					switch {
					case modelObj.DisplayName != "":
						st.Model = modelObj.DisplayName
					case modelObj.OmcDisplay != "":
						st.Model = modelObj.OmcDisplay
					case modelObj.Name != "":
						st.Model = modelObj.Name
					default:
						st.Model = modelObj.ID
					}
				}
			}
		}
		// CC 2.x 표준 stdin 필드가 우선. contextBar/contextTokens는 OMC 래퍼 폴백.
		if raw.ContextWindow.TotalInputTokens > 0 {
			st.ContextTokens = raw.ContextWindow.TotalInputTokens
			st.ContextLimit = raw.ContextWindow.ContextWindowSize
		} else if raw.ContextTokens > 0 {
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
		if raw.Permission != "" {
			st.Permission = raw.Permission
		}
		if raw.TerminalWidth > 0 {
			st.TerminalWidth = raw.TerminalWidth
		}
	}
	return st, nil
}
