package adapter

import (
	"encoding/json"
	"math"
	"strings"
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
		Quota map[string]struct {
			RemainingFraction float64 `json:"remaining_fraction"`
		} `json:"quota"`
		ToolConfirmationPending bool `json:"tool_confirmation_pending"`
		TerminalWidth          int  `json:"terminal_width"`
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
		// quota 파싱: gemini-*/3p-* prefix로 카테고리 그룹핑
		if len(raw.Quota) > 0 {
			categories := map[string]*model.QuotaCategory{}
			for key, q := range raw.Quota {
				var catName, period string
				if strings.HasPrefix(key, "gemini-") {
					catName = "gemini"
					period = strings.TrimPrefix(key, "gemini-")
				} else if strings.HasPrefix(key, "3p-") {
					catName = "3rd"
					period = strings.TrimPrefix(key, "3p-")
				} else {
					continue
				}
				cat, ok := categories[catName]
				if !ok {
					cat = &model.QuotaCategory{Name: catName}
					categories[catName] = cat
				}
				switch period {
				case "5h":
					cat.FiveH = q.RemainingFraction
				case "weekly":
					cat.Weekly = q.RemainingFraction
				}
			}
			// gemini 먼저, 3p 나중에 (안정적 순서)
			for _, name := range []string{"gemini", "3rd"} {
				if cat, ok := categories[name]; ok {
					st.Quota = append(st.Quota, *cat)
				}
			}
		}
		if raw.ToolConfirmationPending {
			st.Permission = "pending"
		}
		if raw.TerminalWidth > 0 {
			st.TerminalWidth = raw.TerminalWidth
		}
	}
	return st, nil
}
