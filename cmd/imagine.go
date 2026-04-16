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
	"strings"
	"time"

	"github.com/legnext-ai/legnext-cli/internal/client"
	"github.com/legnext-ai/legnext-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	imagineAR      string
	imagineMode    string
	imagineOut     string
	imagineNoWait  bool
	imagineTimeout time.Duration
	imagineText    bool
)

var imagineCmd = &cobra.Command{
	Use:   "imagine <prompt>",
	Short: "Generate an image from a prompt (blocks until done)",
	Args:  cobra.ExactArgs(1),
	RunE:  runImagine,
}

func init() {
	rootCmd.AddCommand(imagineCmd)
	imagineCmd.Flags().StringVar(&imagineAR, "ar", "", "aspect ratio appended to prompt as --ar <value> (e.g. 16:9)")
	imagineCmd.Flags().StringVar(&imagineMode, "mode", "fast", "process mode: fast, turbo, or relax")
	imagineCmd.Flags().StringVar(&imagineOut, "out", "", "save first image to this path (file or directory)")
	imagineCmd.Flags().BoolVar(&imagineNoWait, "no-wait", false, "return task_id immediately, don't poll")
	imagineCmd.Flags().DurationVar(&imagineTimeout, "timeout", 10*time.Minute, "max time to wait for task to complete")
	imagineCmd.Flags().BoolVar(&imagineText, "text", false, "human-readable output instead of JSON")
}

func runImagine(cmd *cobra.Command, args []string) error {
	prompt := strings.TrimSpace(args[0])
	if prompt == "" {
		return fmt.Errorf("empty prompt")
	}
	if imagineAR != "" && !strings.Contains(prompt, "--ar ") {
		prompt = prompt + " --ar " + imagineAR
	}
	if imagineMode != "" && imagineMode != "fast" && !strings.Contains(prompt, "--"+imagineMode) {
		prompt = prompt + " --" + imagineMode
	}

	apiKey, apiBase, err := config.Resolve()
	if err != nil {
		return err
	}
	if apiKey == "" {
		return fmt.Errorf("no API key; run `legnext login` or set %s", config.EnvAPIKey)
	}
	c := client.New(apiKey, apiBase)

	ctx := context.Background()
	started := time.Now()

	submitData, err := c.Post(ctx, "/api/v1/diffusion", map[string]any{
		"text": prompt,
	})
	if err != nil {
		return fmt.Errorf("submit task: %w", err)
	}

	var submitted struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(submitData, &submitted); err != nil {
		return fmt.Errorf("parse submit response: %w", err)
	}
	if submitted.JobID == "" {
		return fmt.Errorf("submit response missing job_id: %s", string(submitData))
	}

	if imagineNoWait {
		return writeImagineOutput(submitData, "", 0)
	}

	pollCtx, cancel := context.WithTimeout(ctx, imagineTimeout)
	defer cancel()
	final, err := c.WaitForTask(pollCtx, submitted.JobID, 3*time.Second)
	if err != nil {
		return err
	}

	var savedPath string
	if imagineOut != "" {
		url := extractFirstImageURL(final)
		if url == "" {
			return fmt.Errorf("task completed but no image URL in response")
		}
		savedPath, err = downloadImage(ctx, url, imagineOut)
		if err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	return writeImagineOutput(final, savedPath, time.Since(started))
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

func writeImagineOutput(data json.RawMessage, savedPath string, elapsed time.Duration) error {
	augmented := data
	var v map[string]any
	if err := json.Unmarshal(data, &v); err == nil {
		if elapsed > 0 {
			v["duration_seconds"] = elapsed.Round(100 * time.Millisecond).Seconds()
		}
		if savedPath != "" {
			v["path"] = savedPath
		}
		if pts := extractCostPoints(v); pts > 0 {
			v["cost_points"] = pts
		}
		if out, err := json.Marshal(v); err == nil {
			augmented = out
		}
	}

	if !imagineText {
		_, err := fmt.Fprintln(os.Stdout, string(augmented))
		return err
	}

	var r struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
		Output struct {
			ImageURL  string   `json:"image_url"`
			ImageURLs []string `json:"image_urls"`
		} `json:"output"`
	}
	if err := json.Unmarshal(augmented, &r); err != nil {
		_, err := fmt.Fprintln(os.Stdout, string(augmented))
		return err
	}
	fmt.Fprintf(os.Stdout, "Task %s — %s\n", r.JobID, r.Status)
	if r.Output.ImageURL != "" {
		fmt.Fprintf(os.Stdout, "  %s\n", r.Output.ImageURL)
	}
	for _, u := range r.Output.ImageURLs {
		fmt.Fprintf(os.Stdout, "  %s\n", u)
	}
	if savedPath != "" {
		fmt.Fprintf(os.Stdout, "Saved: %s\n", savedPath)
	}
	return nil
}
