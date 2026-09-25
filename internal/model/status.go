package model

// QuotaCategory — gemini 또는 3p 카테고리별 quota 잔여 비율
type QuotaCategory struct {
	Name    string  // "gemini" 또는 "3p"
	FiveH   float64 // 5시간 잔여 비율 (0.0~1.0)
	Weekly  float64 // 주간 잔여 비율 (0.0~1.0)
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

	ContextTokens int
	ContextLimit  int
	PromptTime    float64

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
