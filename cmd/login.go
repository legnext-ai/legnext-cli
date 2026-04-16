package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/legnext-ai/legnext-cli/internal/client"
	"github.com/legnext-ai/legnext-cli/internal/config"
	"github.com/spf13/cobra"
)

var loginAPIKey string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Validate and save your Legnext API key",
	Long: `Validate an API key against the Legnext API, then save it to
~/.legnext/config.json (mode 0600). Get your key at
https://legnext.ai/app/api-keys.

Non-interactive:
  legnext login --api-key af3e...

Skip login entirely (agent / CI):
  export LEGNEXT_API_KEY=af3e...
  legnext balance`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key (skip interactive prompt)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	key := strings.TrimSpace(loginAPIKey)
	if key == "" {
		key = strings.TrimSpace(os.Getenv(config.EnvAPIKey))
	}
	if key == "" {
		fmt.Fprint(os.Stderr, "Paste your API key (from https://legnext.ai/app/api-keys): ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return fmt.Errorf("read key: %w", err)
		}
		key = strings.TrimSpace(line)
	}
	if key == "" {
		return fmt.Errorf("empty API key")
	}

	apiBase := strings.TrimSpace(os.Getenv(config.EnvAPIBase))
	if apiBase == "" {
		apiBase = config.DefaultAPIBase
	}

	c := client.New(key, apiBase)
	if _, err := c.Balance(context.Background()); err != nil {
		return fmt.Errorf("validate key against %s: %w", apiBase, err)
	}

	saved := &config.Config{APIKey: key}
	if apiBase != config.DefaultAPIBase {
		saved.APIBase = apiBase
	}
	if err := saved.Save(); err != nil {
		return err
	}

	path, _ := config.Path()
	return json.NewEncoder(os.Stdout).Encode(map[string]string{
		"status":  "ok",
		"api_key": config.MaskKey(key),
		"saved":   path,
	})
}
