package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	APIKey  string
	APIBase string
	HTTP    *http.Client
}

// internalFields are stripped from response data before returning to callers.
// They expose upstream implementation details that leak Legnext's supply chain
// and aren't useful to agents consuming the CLI.
var internalFields = []string{
	"provider",
	"provider_name",
	"worker",
	"worker_id",
	"upstream_task_id",
	"upstream_provider",
	"upstream_status",
}

func New(apiKey, apiBase string) *Client {
	return &Client{
		APIKey:  apiKey,
		APIBase: apiBase,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Get performs an authenticated GET, unwraps the `{code,data,message}`
// envelope, and strips internal fields from the data payload.
func (c *Client) Get(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIBase+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s %s: HTTP %d: %s", req.Method, path, resp.StatusCode, truncate(string(body), 500))
	}
	return unwrap(body)
}

// unwrap parses the Legnext API envelope and strips internal fields.
func unwrap(body []byte) (json.RawMessage, error) {
	var env struct {
		Code    int             `json:"code"`
		Data    json.RawMessage `json:"data"`
		Message string          `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parse envelope: %w (body=%s)", err, truncate(string(body), 200))
	}
	if env.Code != 0 && env.Code != 200 {
		return nil, fmt.Errorf("API error %d: %s", env.Code, env.Message)
	}
	return sanitize(env.Data), nil
}

// sanitize recursively removes internal-only fields from a JSON value.
// On any parse error the input is returned untouched (fail-open for output).
func sanitize(data json.RawMessage) json.RawMessage {
	if len(data) == 0 {
		return data
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	stripInternal(v)
	out, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return out
}

func stripInternal(v any) {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range internalFields {
			delete(x, k)
		}
		for _, val := range x {
			stripInternal(val)
		}
	case []any:
		for _, item := range x {
			stripInternal(item)
		}
	}
}

func (c *Client) Balance(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/api/account/balance")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
