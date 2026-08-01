# How-To Guide: Add a New CLI Host Adapter

This guide explains how to implement a new CLI host adapter (e.g. for `codex`, `claude`, or a custom AI CLI).

---

## Step 1: Define Adapter Struct

Create `internal/adapter/<my_cli>.go`:

```go
package adapter

import (
	"encoding/json"
	"statusline/internal/model"
)

type MyCLIAdapter struct{}

func (a *MyCLIAdapter) Parse(input []byte, env map[string]string) (*model.UnifiedStatus, error) {
	st := model.NewUnifiedStatus("my_cli")
	if len(input) == 0 {
		return st, nil
	}

	var raw struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(input, &raw); err == nil {
		st.Model = raw.Model
	}

	return st, nil
}
```

## Step 2: Register in Discriminator (`internal/adapter/auto.go`)

Update `ParseInput`:

```go
func ParseInput(cliFlag string, rawInput []byte, env map[string]string) (*model.UnifiedStatus, error) {
	targetEngine := cliFlag
	if targetEngine == "auto" || targetEngine == "" {
		if bytes.Contains(rawInput, []byte("my_cli_marker")) {
			targetEngine = "my_cli"
		}
        ...
	}
    ...
}
```

## Step 3: Add Unit Tests

Add a test function in `internal/adapter/adapter_test.go`:

```go
func TestMyCLIAdapter(t *testing.T) {
    ...
}
```

Run `go test -v ./internal/adapter` to verify.
