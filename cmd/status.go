package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/legnext-ai/legnext-cli/internal/client"
	"github.com/legnext-ai/legnext-cli/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <task_id>",
	Short: "Fetch a task's current state without polling",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	apiKey, apiBase, err := config.Resolve()
	if err != nil {
		return err
	}
	if apiKey == "" {
		return fmt.Errorf("no API key; run `legnext login` or set %s", config.EnvAPIKey)
	}

	c := client.New(apiKey, apiBase)
	data, err := c.GetTask(context.Background(), args[0])
	if err != nil {
		return err
	}

	var v map[string]any
	if err := json.Unmarshal(data, &v); err == nil {
		if pts := extractCostPoints(v); pts > 0 {
			v["cost_points"] = pts
		}
		if out, err := json.Marshal(v); err == nil {
			data = out
		}
	}

	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}
