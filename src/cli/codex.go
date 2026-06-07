package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

var codexCmd = &cobra.Command{
	Use:   "codex",
	Short: "Render a prompt for OpenAI Codex status data",
	Long: `Render a prompt for OpenAI Codex status data.

This command reads Codex status JSON data from stdin and renders a prompt
that can include a Codex segment with model, token, and rate limit information.

Example usage:
  cat codex-status.json | oh-my-posh codex --config ~/.config/ohmyposh/codex.omp.json`,
	Args: cobra.NoArgs,
	Run: statuslineRun[segments.CodexData](
		shell.CODEX,
		cache.CODEXCACHE,
		func(d *segments.CodexData) string { return d.ThreadID },
		config.Codex,
	),
}

func init() {
	RootCmd.AddCommand(codexCmd)
}
