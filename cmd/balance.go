package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/legnext-ai/legnext-cli/internal/client"
	"github.com/legnext-ai/legnext-cli/internal/config"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show account balance",
	RunE:  runBalance,
}

func init() {
	rootCmd.AddCommand(balanceCmd)
}

func runBalance(cmd *cobra.Command, args []string) error {
	apiKey, apiBase, err := config.Resolve()
	if err != nil {
		return err
	}
	if apiKey == "" {
		return fmt.Errorf("no API key; run `legnext login` or set %s", config.EnvAPIKey)
	}

	c := client.New(apiKey, apiBase)
	body, err := c.Balance(context.Background())
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(body))
	return err
}
