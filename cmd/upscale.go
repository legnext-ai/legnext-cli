package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	upscaleType    string
	upscaleOut     string
	upscaleNoWait  bool
	upscaleTimeout time.Duration
	upscaleText    bool
)

var upscaleCmd = &cobra.Command{
	Use:   "upscale <task_id> <image_no>",
	Short: "Upscale one image from a completed grid (image_no: 0-3)",
	Args:  cobra.ExactArgs(2),
	RunE:  runUpscale,
}

func init() {
	rootCmd.AddCommand(upscaleCmd)
	upscaleCmd.Flags().StringVar(&upscaleType, "type", "subtle", "upscale variant: subtle or creative")
	upscaleCmd.Flags().StringVar(&upscaleOut, "out", "", "save upscaled image to this path")
	upscaleCmd.Flags().BoolVar(&upscaleNoWait, "no-wait", false, "return task_id immediately, don't poll")
	upscaleCmd.Flags().DurationVar(&upscaleTimeout, "timeout", 10*time.Minute, "max time to wait")
	upscaleCmd.Flags().BoolVar(&upscaleText, "text", false, "human-readable output")
}

func runUpscale(cmd *cobra.Command, args []string) error {
	jobID := args[0]
	imageNo, err := strconv.Atoi(args[1])
	if err != nil || imageNo < 0 || imageNo > 3 {
		return fmt.Errorf("image_no must be 0, 1, 2, or 3 (got %q)", args[1])
	}
	typeCode, err := upscaleTypeCode(upscaleType)
	if err != nil {
		return err
	}

	c, err := buildClient()
	if err != nil {
		return err
	}
	final, err := submitAndWait(c, "/api/v1/upscale", map[string]any{
		"jobId":   jobID,
		"imageNo": imageNo,
		"type":    typeCode,
	}, submitOpts{Out: upscaleOut, NoWait: upscaleNoWait, Timeout: upscaleTimeout})
	if err != nil {
		return err
	}
	return writeTaskResult(final, upscaleText)
}

func upscaleTypeCode(name string) (int, error) {
	switch name {
	case "subtle", "":
		return 0, nil
	case "creative":
		return 1, nil
	}
	return 0, fmt.Errorf("unknown --type %q (subtle or creative)", name)
}
