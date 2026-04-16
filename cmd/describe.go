package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	describeNoWait  bool
	describeTimeout time.Duration
	describeText    bool
)

var describeCmd = &cobra.Command{
	Use:   "describe <image_url>",
	Short: "Describe an image, returning suggested Midjourney prompts",
	Long: `Describe an image by its public URL. The image must be reachable
over HTTP(S); local file paths are not yet supported.

Returns up to 4 suggested Midjourney-style prompts from the model.`,
	Args: cobra.ExactArgs(1),
	RunE: runDescribe,
}

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.Flags().BoolVar(&describeNoWait, "no-wait", false, "return task_id immediately, don't poll")
	describeCmd.Flags().DurationVar(&describeTimeout, "timeout", 5*time.Minute, "max time to wait")
	describeCmd.Flags().BoolVar(&describeText, "text", false, "human-readable output (lists the 4 prompts)")
}

func runDescribe(cmd *cobra.Command, args []string) error {
	c, err := buildClient()
	if err != nil {
		return err
	}
	final, err := submitAndWait(c, "/api/v1/describe", map[string]any{
		"imgUrl": args[0],
	}, submitOpts{NoWait: describeNoWait, Timeout: describeTimeout})
	if err != nil {
		return err
	}

	if !describeText {
		_, err := fmt.Fprintln(os.Stdout, string(final))
		return err
	}
	return writeDescribeText(final)
}

func writeDescribeText(data json.RawMessage) error {
	var r struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
		Output struct {
			Prompts []string `json:"prompts"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	fmt.Fprintf(os.Stdout, "Task %s — %s\n", r.JobID, r.Status)
	for i, p := range r.Output.Prompts {
		fmt.Fprintf(os.Stdout, "\n[%d] %s\n", i+1, p)
	}
	return nil
}
