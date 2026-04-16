package client

import (
	"bytes"
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
// They expose upstream implementation details, internal config, or noise
// that agents consuming the CLI don't need.
var internalFields = []string{
	// Supply-chain identity
	"provider",
	"provider_name",
	"worker",
	"worker_id",
	"upstream_task_id",
	"upstream_provider",
	"upstream_status",
	// Request echo / internal task config
	"config",
	"webhook_config",
	"input",
	"detail",
	"logs",
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

// unwrap detects the Legnext `{code,data,message}` envelope and strips it,
// falling back to passing the body through when the response is flat.
// Not all endpoints wrap — /api/account/* do, /api/v1/* return flat task JSON.
func unwrap(body []byte) (json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		// Not a JSON object (array, primitive, null). Pass through.
		return body, nil
	}
	_, hasCode := raw["code"]
	_, hasData := raw["data"]
	if hasCode && hasData {
		var code int
		_ = json.Unmarshal(raw["code"], &code)
		if code != 0 && code != 200 {
			var msg string
			_ = json.Unmarshal(raw["message"], &msg)
			return nil, fmt.Errorf("API error %d: %s", code, msg)
		}
		return sanitize(raw["data"]), nil
	}
	return sanitize(body), nil
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

// Post performs an authenticated POST with JSON body, then unwraps the envelope.
func (c *Client) Post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIBase+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("POST %s: HTTP %d: %s", path, resp.StatusCode, truncate(string(respBody), 500))
	}
	return unwrap(respBody)
}

func (c *Client) GetTask(ctx context.Context, taskID string) (json.RawMessage, error) {
	return c.Get(ctx, "/api/v1/job/"+taskID)
}

// WaitForTask polls GetTask at the given interval until the task reaches
// a terminal status (`completed` or `failed`) or ctx is cancelled.
// Non-terminal statuses (`pending`, `staged`, `processing`, `retry`) trigger
// another poll. The returned payload is the final task JSON on terminal status.
func (c *Client) WaitForTask(ctx context.Context, taskID string, interval time.Duration) (json.RawMessage, error) {
	for {
		data, err := c.GetTask(ctx, taskID)
		if err != nil {
			return nil, err
		}
		var t struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("parse task status: %w", err)
		}
		switch t.Status {
		case "completed", "failed":
			return data, nil
		case "pending", "staged", "processing", "retry", "":
			// keep polling
		default:
			return nil, fmt.Errorf("unknown task status: %s", t.Status)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for task %s: %w", taskID, ctx.Err())
		case <-time.After(interval):
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
