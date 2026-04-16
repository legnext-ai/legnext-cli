package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/legnext-ai/legnext-cli/internal/client"
	"github.com/legnext-ai/legnext-cli/internal/config"
)

func buildClient() (*client.Client, error) {
	apiKey, apiBase, err := config.Resolve()
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no API key; run `legnext login` or set %s", config.EnvAPIKey)
	}
	return client.New(apiKey, apiBase), nil
}

type submitOpts struct {
	Out     string
	NoWait  bool
	Timeout time.Duration
}

// submitAndWait handles the full task lifecycle used by imagine/upscale/variation/describe:
// POST the body, optionally wait for the task to terminate, optionally download
// the first output image, and augment with duration_seconds / cost_points / path.
func submitAndWait(c *client.Client, apiPath string, body any, opts submitOpts) (json.RawMessage, error) {
	ctx := context.Background()
	started := time.Now()

	submitData, err := c.Post(ctx, apiPath, body)
	if err != nil {
		return nil, fmt.Errorf("submit: %w", err)
	}
	if opts.NoWait {
		return submitData, nil
	}

	var submitted struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(submitData, &submitted); err != nil {
		return nil, fmt.Errorf("parse submit response: %w", err)
	}
	if submitted.JobID == "" {
		return nil, fmt.Errorf("submit response missing job_id: %s", string(submitData))
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	final, err := c.WaitForTask(pollCtx, submitted.JobID, 3*time.Second)
	if err != nil {
		return nil, err
	}

	var savedPath string
	if opts.Out != "" {
		url := extractFirstImageURL(final)
		if url == "" {
			return nil, fmt.Errorf("task completed but no image URL in response")
		}
		savedPath, err = downloadImage(ctx, url, opts.Out)
		if err != nil {
			return nil, fmt.Errorf("download: %w", err)
		}
	}

	return augmentResult(final, time.Since(started), savedPath), nil
}

func augmentResult(data json.RawMessage, elapsed time.Duration, savedPath string) json.RawMessage {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	if elapsed > 0 {
		v["duration_seconds"] = elapsed.Round(100 * time.Millisecond).Seconds()
	}
	if savedPath != "" {
		v["path"] = savedPath
	}
	if pts := extractCostPoints(v); pts > 0 {
		v["cost_points"] = pts
	}
	out, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return out
}

// extractCostPoints pulls meta.usage.consume from a task response.
func extractCostPoints(v map[string]any) int {
	meta, ok := v["meta"].(map[string]any)
	if !ok {
		return 0
	}
	usage, ok := meta["usage"].(map[string]any)
	if !ok {
		return 0
	}
	consume, ok := usage["consume"].(float64)
	if !ok {
		return 0
	}
	return int(consume)
}

func extractFirstImageURL(data json.RawMessage) string {
	var r struct {
		Output struct {
			ImageURL  string   `json:"image_url"`
			ImageURLs []string `json:"image_urls"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return ""
	}
	if r.Output.ImageURL != "" {
		return r.Output.ImageURL
	}
	if len(r.Output.ImageURLs) > 0 {
		return r.Output.ImageURLs[0]
	}
	return ""
}

func downloadImage(ctx context.Context, url, dest string) (string, error) {
	target := dest
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		target = filepath.Join(dest, path.Base(url))
	}
	if dir := filepath.Dir(target); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}

	f, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}
	return target, nil
}

// writeTaskResult writes a task payload as JSON (default) or as human-readable
// text when asText is true. Used by image-producing commands.
func writeTaskResult(data json.RawMessage, asText bool) error {
	if !asText {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	var r struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
		Path   string `json:"path"`
		Output struct {
			ImageURL  string   `json:"image_url"`
			ImageURLs []string `json:"image_urls"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	fmt.Fprintf(os.Stdout, "Task %s — %s\n", r.JobID, r.Status)
	if r.Output.ImageURL != "" {
		fmt.Fprintf(os.Stdout, "  %s\n", r.Output.ImageURL)
	}
	for _, u := range r.Output.ImageURLs {
		fmt.Fprintf(os.Stdout, "  %s\n", u)
	}
	if r.Path != "" {
		fmt.Fprintf(os.Stdout, "Saved: %s\n", r.Path)
	}
	return nil
}
