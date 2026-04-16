package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillMarkdown []byte

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install the Claude Code skill so your agent can discover legnext",
	Long: `Write ~/.claude/skills/legnext/SKILL.md so Claude Code (and any other
SKILL.md-aware agent) knows when to invoke the Legnext CLI and how each
command works.

Idempotent by default — refuses to overwrite an existing skill unless
--force is given.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite an existing SKILL.md")
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".claude", "skills", "legnext")
	path := filepath.Join(dir, "SKILL.md")

	if !initForce {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("skill already exists at %s; pass --force to overwrite", path)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, skillMarkdown, 0o644); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"status":    "ok",
		"skill":     "legnext",
		"path":      path,
		"overwrote": initForce,
	})
}
