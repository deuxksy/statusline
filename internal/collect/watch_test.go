package collect

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSessionEventFiltersContent(t *testing.T) {
	if _, ok := sessionEvent([]byte(`{"type":"event_msg","payload":{"type":"user_message","message":"private"}}`)); ok {
		t.Fatal("user content must not be emitted")
	}
	event, ok := sessionEvent([]byte(`{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":3}},"rate_limits":{"primary":{"used_percent":7}},"message":"private"}}`))
	if !ok {
		t.Fatal("token event omitted")
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || !json.Valid(data) || bytes.Contains(data, []byte("private")) {
		t.Fatal("invalid token event")
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 {
		t.Fatalf("unexpected fields: %v", fields)
	}
	if _, ok := sessionEvent([]byte(`{"type":"turn_context","payload":{"model":"gpt","cwd":"/tmp","effort":"high","developer_instructions":"private"}}`)); !ok {
		t.Fatal("context event omitted")
	}
}
