package adapter

import (
	"encoding/json"
	"math"
	"statusline/internal/model"
)

type AntigravityAdapter struct{}

func (a *AntigravityAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("antigravity")
	if len(input) == 0 {
		return st, nil
	}

	var raw struct {
		Model         json.RawMessage `json:"model"`
		ContextTokens int             `json:"contextTokens"`
		ContextBar    struct {
			Percentage int `json:"percentage"`
		} `json:"contextBar"`
		ContextWindow struct {
			UsedPercentage   float64 `json:"used_percentage"`
			TotalInputTokens int     `json:"total_input_tokens"`
		} `json:"context_window"`
		Thinking struct {
			State string `json:"state"`
		} `json:"thinking"`
		ThinkingState     string   `json:"thinkingState"`
		AgentState        string   `json:"agent_state"`
		ActiveSkills      []string `json:"activeSkills"`
		ActiveSkillsSnake []string `json:"active_skills"`
		LastTool          string   `json:"lastTool"`
		LastToolSnake     string   `json:"last_tool"`
		Cwd               string   `json:"cwd"`
		Workspace         struct {
			CurrentDir string `json:"current_dir"`
		} `json:"workspace"`
	}

	if err := json.Unmarshal(input, &raw); err == nil {
		if len(raw.Model) > 0 {
			var modelStr string
			if err := json.Unmarshal(raw.Model, &modelStr); err == nil {
				st.Model = modelStr
			} else {
				var modelObj struct {
					DisplayName      string `json:"displayName"`
					DisplayNameSnake string `json:"display_name"`
					Name             string `json:"name"`
					ID               string `json:"id"`
				}
				if err := json.Unmarshal(raw.Model, &modelObj); err == nil {
					if modelObj.DisplayName != "" {
						st.Model = modelObj.DisplayName
					} else if modelObj.DisplayNameSnake != "" {
						st.Model = modelObj.DisplayNameSnake
					} else if modelObj.ID != "" {
						st.Model = modelObj.ID
					} else if modelObj.Name != "" {
						st.Model = modelObj.Name
					}
				}
			}
		}
		if raw.ContextWindow.UsedPercentage > 0 {
			st.ContextTokens = int(math.Round(raw.ContextWindow.UsedPercentage))
		} else if raw.ContextTokens > 0 {
			st.ContextTokens = raw.ContextTokens
		} else if raw.ContextBar.Percentage > 0 {
			st.ContextTokens = raw.ContextBar.Percentage
		}
		if raw.ThinkingState != "" {
			st.ThinkingState = raw.ThinkingState
		} else if raw.Thinking.State != "" {
			st.ThinkingState = raw.Thinking.State
		} else if raw.AgentState != "" {
			st.ThinkingState = raw.AgentState
		}
		if len(raw.ActiveSkills) > 0 {
			st.ActiveSkills = raw.ActiveSkills
		} else if len(raw.ActiveSkillsSnake) > 0 {
			st.ActiveSkills = raw.ActiveSkillsSnake
		}
		if raw.LastTool != "" {
			st.LastTool = raw.LastTool
		} else if raw.LastToolSnake != "" {
			st.LastTool = raw.LastToolSnake
		}
		if raw.Cwd != "" {
			st.Cwd = raw.Cwd
		} else if raw.Workspace.CurrentDir != "" {
			st.Cwd = raw.Workspace.CurrentDir
		}
	}
	return st, nil
}
