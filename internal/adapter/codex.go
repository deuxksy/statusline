package adapter

import (
	"encoding/json"
	"math"
	"statusline/internal/model"
)

type CodexAdapter struct{}

func (c *CodexAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("codex")
	st.Model = "codex"
	if len(input) == 0 {
		return st, nil
	}

	var raw struct {
		Model  string `json:"model"`
		Cwd    string `json:"cwd"`
		Effort string `json:"effort"`
		Info   struct {
			TotalTokenUsage    model.TokenUsage `json:"total_token_usage"`
			LastTokenUsage     model.TokenUsage `json:"last_token_usage"`
			ModelContextWindow int              `json:"model_context_window"`
		} `json:"info"`
		RateLimits struct {
			Primary *struct {
				UsedPercent float64 `json:"used_percent"`
				ResetsAt    int64   `json:"resets_at"`
			} `json:"primary"`
			Secondary *struct {
				UsedPercent float64 `json:"used_percent"`
				ResetsAt    int64   `json:"resets_at"`
			} `json:"secondary"`
		} `json:"rate_limits"`
	}
	if err := json.Unmarshal(input, &raw); err != nil {
		return st, nil
	}
	if raw.Model != "" {
		st.Model = raw.Model
	}
	st.Cwd = raw.Cwd
	st.ReasoningEffort = raw.Effort
	st.TotalTokenUsage = raw.Info.TotalTokenUsage
	st.LastTokenUsage = raw.Info.LastTokenUsage
	st.ContextLimit = raw.Info.ModelContextWindow
	if st.ContextLimit > 0 && st.LastTokenUsage.InputTokens > 0 {
		st.ContextTokens = min(100, int(math.Round(float64(st.LastTokenUsage.InputTokens)*100/float64(st.ContextLimit))))
		st.Capabilities.HasTokens = true
	}
	if raw.RateLimits.Primary != nil && raw.RateLimits.Secondary != nil {
		st.Quota = []model.QuotaCategory{{
			Name:           "openai",
			FiveH:          math.Max(0, math.Min(1, 1-raw.RateLimits.Primary.UsedPercent/100)),
			Weekly:         math.Max(0, math.Min(1, 1-raw.RateLimits.Secondary.UsedPercent/100)),
			FiveHResetsAt:  raw.RateLimits.Primary.ResetsAt,
			WeeklyResetsAt: raw.RateLimits.Secondary.ResetsAt,
		}}
		st.Capabilities.HasQuota = true
	}
	return st, nil
}
