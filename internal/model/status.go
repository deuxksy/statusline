package model

type HostCapabilities struct {
	HasTokens     bool
	HasSkills     bool
	HasTools      bool
	HasThinking   bool
	HasPermission bool
}

type UnifiedStatus struct {
	EngineName    string
	Model         string
	Cwd           string
	Hostname      string

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
	return &UnifiedStatus{
		EngineName:   engineName,
		Capabilities: caps,
	}
}
