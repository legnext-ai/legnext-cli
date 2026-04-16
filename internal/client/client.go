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

func New(apiKey, apiBase string) *Client {
	return &Client{
		APIKey:  apiKey,
		APIBase: apiBase,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Get performs an authenticated GET and returns the raw JSON body on 2xx.
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
	return body, nil
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
