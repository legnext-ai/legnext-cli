package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "legnext",
	Short:   "Midjourney CLI for AI agents",
	Version: version,
	Long: `Legnext CLI — generate Midjourney images from the command line.

Designed for AI coding agents (Claude Code, Cursor, Cline) and shell scripts.
Outputs JSON by default. Blocks until tasks complete (auto-polls upstream).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
