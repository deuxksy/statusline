package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type PlatformSnapshot struct {
	Usage json.RawMessage `json:"usage,omitempty"`
	Costs json.RawMessage `json:"costs,omitempty"`
}

// Platform queries organization reporting endpoints only with an Admin API key.
func Platform(ctx context.Context, client *http.Client, baseURL, adminKey string, start time.Time) (PlatformSnapshot, error) {
	var snapshot PlatformSnapshot
	if adminKey == "" {
		return snapshot, nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	for _, endpoint := range []struct {
		path   string
		target *json.RawMessage
	}{
		{"organization/usage/completions", &snapshot.Usage},
		{"organization/costs", &snapshot.Costs},
	} {
		u, err := url.Parse(baseURL + "/" + endpoint.path)
		if err != nil {
			return snapshot, err
		}
		q := u.Query()
		q.Set("start_time", strconv.FormatInt(start.Unix(), 10))
		q.Set("bucket_width", "1d")
		u.RawQuery = q.Encode()
		var buckets []json.RawMessage
		for {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				return snapshot, err
			}
			req.Header.Set("Authorization", "Bearer "+adminKey)
			resp, err := client.Do(req)
			if err != nil {
				return snapshot, err
			}
			var page struct {
				Data     []json.RawMessage `json:"data"`
				HasMore  bool              `json:"has_more"`
				NextPage string            `json:"next_page"`
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return snapshot, fmt.Errorf("%s: HTTP %d", endpoint.path, resp.StatusCode)
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
			resp.Body.Close()
			if err != nil {
				return snapshot, err
			}
			buckets = append(buckets, page.Data...)
			if !page.HasMore {
				break
			}
			if page.NextPage == "" {
				return snapshot, fmt.Errorf("%s: missing next_page", endpoint.path)
			}
			q.Set("page", page.NextPage)
			u.RawQuery = q.Encode()
		}
		data, err := json.Marshal(map[string]any{"object": "page", "data": buckets, "has_more": false})
		if err != nil {
			return snapshot, err
		}
		*endpoint.target = data
	}
	return snapshot, nil
}
