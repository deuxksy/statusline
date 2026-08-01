package adapter

import "statusline/internal/model"

type EngineAdapter interface {
	Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error)
}
