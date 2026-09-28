package collect

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type rpcResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type codexSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
}

func startCodex(ctx context.Context, binary string, args ...string) (*codexSession, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	session := &codexSession{cmd: cmd, stdin: stdin, reader: bufio.NewReader(stdout)}
	if _, err := session.request(1, "initialize", map[string]any{"clientInfo": map[string]string{"name": "statusline", "version": "0.4.0"}}); err != nil {
		session.close()
		return nil, err
	}
	if _, err := io.WriteString(stdin, "{\"method\":\"initialized\"}\n"); err != nil {
		session.close()
		return nil, err
	}
	return session, nil
}

func (s *codexSession) close() {
	_ = s.stdin.Close()
	_ = s.cmd.Wait()
}

func (s *codexSession) request(id int, method string, params any) (json.RawMessage, error) {
	body, err := json.Marshal(map[string]any{"id": id, "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintln(s.stdin, string(body)); err != nil {
		return nil, err
	}
	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		var response rpcResponse
		if json.Unmarshal(line, &response) != nil || response.ID != id {
			continue
		}
		if response.Error != nil {
			return nil, fmt.Errorf("%s: %s", method, response.Error.Message)
		}
		if len(response.Result) == 0 {
			return nil, fmt.Errorf("%s: empty result", method)
		}
		return response.Result, nil
	}
}

// Codex reads account snapshots from the locally authenticated app-server.
func Codex(ctx context.Context, binary string) (map[string]json.RawMessage, error) {
	session, err := startCodex(ctx, binary, "app-server", "--stdio")
	if err != nil {
		return nil, err
	}
	defer session.close()
	results := make(map[string]json.RawMessage, 3)
	for id, item := range []struct{ key, method string }{{"account", "account/read"}, {"rate_limits", "account/rateLimits/read"}, {"usage", "account/usage/read"}} {
		result, err := session.request(id+2, item.method, map[string]any{})
		if err != nil {
			return nil, err
		}
		results[item.key] = result
	}
	list, err := session.request(5, "thread/list", map[string]any{"limit": 1, "sortKey": "updated_at", "useStateDbOnly": true})
	if err != nil {
		return nil, err
	}
	var threads struct {
		Data []struct {
			ID              string          `json:"id"`
			Model           string          `json:"model"`
			ReasoningEffort string          `json:"reasoningEffort"`
			Cwd             string          `json:"cwd"`
			Status          json.RawMessage `json:"status"`
			UpdatedAt       int64           `json:"updatedAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list, &threads); err != nil {
		return nil, err
	}
	if len(threads.Data) > 0 {
		thread := threads.Data[0]
		metadata, err := json.Marshal(thread)
		if err != nil {
			return nil, err
		}
		results["latest_thread"] = metadata
		usage, err := session.request(6, "account/usage/read", map[string]any{"threadId": thread.ID})
		if err != nil {
			return nil, err
		}
		results["latest_thread_usage"] = usage
	}
	return results, nil
}
