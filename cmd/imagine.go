package cmd

import (
	"fmt"
	"strings"
	"time"

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
	imagineCmd.Flags().StringVar(&imagineAR, "ar", "", "aspect ratio, appended to prompt as --ar <value> (e.g. 16:9)")
	imagineCmd.Flags().StringVar(&imagineMode, "mode", "fast", "process mode: fast, turbo, or relax (non-fast is appended to prompt)")
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
		prompt += " --ar " + imagineAR
	}
	if imagineMode != "" && imagineMode != "fast" && !strings.Contains(prompt, "--"+imagineMode) {
		prompt += " --" + imagineMode
	}

	c, err := buildClient()
	if err != nil {
		return err
	}
	final, err := submitAndWait(c, "/api/v1/diffusion", map[string]any{
		"text": prompt,
	}, submitOpts{Out: imagineOut, NoWait: imagineNoWait, Timeout: imagineTimeout})
	if err != nil {
		return err
	}
	return writeTaskResult(final, imagineText)
}
