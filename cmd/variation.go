package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	variationType    string
	variationRemix   string
	variationOut     string
	variationNoWait  bool
	variationTimeout time.Duration
	variationText    bool
)

var variationCmd = &cobra.Command{
	Use:   "variation <task_id> <image_no>",
	Short: "Create a variation from one image of a completed grid (image_no: 0-3)",
	Args:  cobra.ExactArgs(2),
	RunE:  runVariation,
}

func init() {
	rootCmd.AddCommand(variationCmd)
	variationCmd.Flags().StringVar(&variationType, "type", "subtle", "variation strength: subtle or strong")
	variationCmd.Flags().StringVar(&variationRemix, "remix", "", "optional remix prompt (max 8192 chars)")
	variationCmd.Flags().StringVar(&variationOut, "out", "", "save first image to this path")
	variationCmd.Flags().BoolVar(&variationNoWait, "no-wait", false, "return task_id immediately, don't poll")
	variationCmd.Flags().DurationVar(&variationTimeout, "timeout", 10*time.Minute, "max time to wait")
	variationCmd.Flags().BoolVar(&variationText, "text", false, "human-readable output")
}

func runVariation(cmd *cobra.Command, args []string) error {
	jobID := args[0]
	imageNo, err := strconv.Atoi(args[1])
	if err != nil || imageNo < 0 || imageNo > 3 {
		return fmt.Errorf("image_no must be 0, 1, 2, or 3 (got %q)", args[1])
	}
	typeCode, err := variationTypeCode(variationType)
	if err != nil {
		return err
	}

	c, err := buildClient()
	if err != nil {
		return err
	}
	body := map[string]any{
		"jobId":   jobID,
		"imageNo": imageNo,
		"type":    typeCode,
	}
	if variationRemix != "" {
		body["remixPrompt"] = variationRemix
	}
	final, err := submitAndWait(c, "/api/v1/variation", body, submitOpts{
		Out: variationOut, NoWait: variationNoWait, Timeout: variationTimeout,
	})
	if err != nil {
		return err
	}
	return writeTaskResult(final, variationText)
}

func variationTypeCode(name string) (int, error) {
	switch name {
	case "subtle", "":
		return 0, nil
	case "strong":
		return 1, nil
	}
	return 0, fmt.Errorf("unknown --type %q (subtle or strong)", name)
}
