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
		StatuslineSnapshot bool   `json:"statusline_snapshot"`
		Model              string `json:"model"`
		Cwd                string `json:"cwd"`
		Effort             string `json:"effort"`
		Speed              string `json:"speed"`
		Permission         string `json:"permission"`
		ContextTokenCount  *int   `json:"context_token_count"`
		ContextUsed        *int   `json:"context_used"`
		ResetTickets       int    `json:"reset_tickets"`
		ResetTicketsExpire string `json:"reset_tickets_expire"`
		Info               struct {
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
	st.CycleMode = raw.Speed
	st.Permission = raw.Permission
	st.Capabilities.HasPermission = raw.Permission != ""
	st.ResetTickets = raw.ResetTickets
	st.ResetTicketsExpire = raw.ResetTicketsExpire
	st.ReasoningEffort = raw.Effort
	st.TotalTokenUsage = raw.Info.TotalTokenUsage
	st.LastTokenUsage = raw.Info.LastTokenUsage
	st.ContextLimit = raw.Info.ModelContextWindow
	if raw.StatuslineSnapshot {
		switch {
		case st.ContextLimit > 0 && raw.ContextTokenCount != nil && *raw.ContextTokenCount >= 0:
			st.ContextTokens = *raw.ContextTokenCount
			st.Capabilities.HasTokens = true
		case raw.ContextUsed != nil:
			st.ContextLimit = 0
			st.ContextTokens = max(0, min(100, *raw.ContextUsed))
			st.Capabilities.HasTokens = true
		}
	} else if st.ContextLimit > 0 && st.LastTokenUsage.InputTokens > 0 {
		st.ContextTokens = min(100, int(math.Round(float64(st.LastTokenUsage.InputTokens)*100/float64(st.ContextLimit))))
		st.Capabilities.HasTokens = true
	}
	hasPrimary, hasSecondary := raw.RateLimits.Primary != nil, raw.RateLimits.Secondary != nil
	if hasPrimary && hasSecondary || raw.StatuslineSnapshot && (hasPrimary || hasSecondary) {
		quota := model.QuotaCategory{Name: "openai", OmitFiveH: raw.RateLimits.Primary == nil, OmitWeekly: raw.RateLimits.Secondary == nil}
		if primary := raw.RateLimits.Primary; primary != nil {
			quota.FiveH = math.Max(0, math.Min(1, 1-primary.UsedPercent/100))
			quota.FiveHResetsAt = primary.ResetsAt
		}
		if secondary := raw.RateLimits.Secondary; secondary != nil {
			quota.Weekly = math.Max(0, math.Min(1, 1-secondary.UsedPercent/100))
			quota.WeeklyResetsAt = secondary.ResetsAt
		}
		st.Quota = []model.QuotaCategory{quota}
		st.Capabilities.HasQuota = true
	}
	return st, nil
}
