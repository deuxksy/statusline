package model

// QuotaCategory — 공급자별 quota 잔여 비율과 reset 시각
type QuotaCategory struct {
	Name           string
	FiveH          float64 // 5시간 잔여 비율 (0.0~1.0)
	Weekly         float64 // 주간 잔여 비율 (0.0~1.0)
	FiveHResetsAt  int64
	WeeklyResetsAt int64
}

type TokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	CacheWriteInputTokens int `json:"cache_write_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
	TotalTokens           int `json:"total_tokens"`
}

type HostCapabilities struct {
	HasTokens     bool
	HasSkills     bool
	HasTools      bool
	HasThinking   bool
	HasPermission bool
	HasQuota      bool
}

type UnifiedStatus struct {
	EngineName string
	Model      string
	Cwd        string
	Hostname   string

	GitRepo   string
	GitBranch string
	GitStatus string

	ContextTokens   int
	ContextLimit    int
	PromptTime      float64
	TotalTokenUsage TokenUsage
	LastTokenUsage  TokenUsage
	ReasoningEffort string

	ActiveSkills []string
	LastSkill    string
	LastTool     string

	Permission    string
	ThinkingState string

	Quota []QuotaCategory

	TerminalWidth int

	Capabilities HostCapabilities
}

func NewUnifiedStatus(engineName string) *UnifiedStatus {
	caps := HostCapabilities{}
	if engineName == "claude" || engineName == "antigravity" {
		caps.HasTokens = true
		caps.HasSkills = true
		caps.HasTools = true
		caps.HasThinking = true
		caps.HasPermission = true
	}
	if engineName == "antigravity" {
		caps.HasQuota = true
	}
	return &UnifiedStatus{
		EngineName:   engineName,
		Capabilities: caps,
	}
}
