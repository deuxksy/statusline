package collect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func testClient(f roundTripFunc) *http.Client                             { return &http.Client{Transport: f} }
func response(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}
}

func TestPlatformSkipsWithoutAdminKey(t *testing.T) {
	called := false
	client := testClient(func(r *http.Request) (*http.Response, error) { called = true; return response(`{}`), nil })
	result, err := Platform(context.Background(), client, "https://example.test/v1", "", time.Now())
	if err != nil || called || result.Usage != nil || result.Costs != nil {
		t.Fatalf("result=%+v err=%v called=%v", result, err, called)
	}
}

func TestPlatformPaginatesAndKeepsSourcesSeparate(t *testing.T) {
	client := testClient(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer admin-test" {
			t.Error("missing admin authorization")
		}
		if r.URL.Query().Get("start_time") != "100" {
			t.Error("missing start_time")
		}
		if strings.HasSuffix(r.URL.Path, "/completions") && r.URL.Query().Get("page") == "" {
			return response(`{"data":[{"results":[{"input_tokens":10}]}],"has_more":true,"next_page":"next"}`), nil
		}
		return response(`{"data":[{"results":[]}],"has_more":false}`), nil
	})
	result, err := Platform(context.Background(), client, "https://example.test/v1", "admin-test", time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	var usage struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result.Usage, &usage); err != nil {
		t.Fatal(err)
	}
	if len(usage.Data) != 2 || len(result.Costs) == 0 {
		t.Fatalf("result=%+v", result)
	}
}
