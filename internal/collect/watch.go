package collect

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

func latestThreadPath(ctx context.Context, binary string) (string, error) {
	session, err := startCodex(ctx, binary, "app-server", "--stdio")
	if err != nil {
		return "", err
	}
	defer session.close()
	result, err := session.request(2, "thread/list", map[string]any{"limit": 1, "sortKey": "updated_at", "useStateDbOnly": true})
	if err != nil {
		return "", err
	}
	var list struct {
		Data []struct {
			Path string `json:"path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &list); err != nil {
		return "", err
	}
	if len(list.Data) == 0 || list.Data[0].Path == "" {
		return "", fmt.Errorf("no recent Codex thread")
	}
	return list.Data[0].Path, nil
}

func sessionEvent(line []byte) (any, bool) {
	var entry struct {
		Type    string `json:"type"`
		Payload struct {
			Type       string          `json:"type"`
			Model      string          `json:"model"`
			Cwd        string          `json:"cwd"`
			Effort     string          `json:"effort"`
			Info       json.RawMessage `json:"info"`
			RateLimits json.RawMessage `json:"rate_limits"`
		} `json:"payload"`
	}
	if json.Unmarshal(line, &entry) != nil {
		return nil, false
	}
	switch {
	case entry.Type == "event_msg" && entry.Payload.Type == "token_count":
		return map[string]any{"type": "token_count", "info": entry.Payload.Info, "rate_limits": entry.Payload.RateLimits}, true
	case entry.Type == "turn_context":
		return map[string]any{"type": "turn_context", "model": entry.Payload.Model, "cwd": entry.Payload.Cwd, "effort": entry.Payload.Effort}, true
	default:
		return nil, false
	}
}

// CodexWatch follows the latest local session's token and context events.
func CodexWatch(ctx context.Context, binary string, output io.Writer) error {
	path, err := latestThreadPath(ctx, binary)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	encoder := json.NewEncoder(output)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var pending []byte
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
		for {
			line, err := reader.ReadBytes('\n')
			pending = append(pending, line...)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if event, ok := sessionEvent(bytes.TrimSpace(pending)); ok {
				if err := encoder.Encode(event); err != nil {
					return err
				}
			}
			pending = pending[:0]
		}
	}
}
