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

This command renders the latest OpenAI Codex token usage data from local Codex
session transcripts. It reads the newest token_count event from
$CODEX_HOME/sessions, falling back to ~/.codex/sessions when CODEX_HOME is not set.

When JSON is provided on stdin, stdin takes precedence over local session discovery.

Example usage:
  oh-my-posh codex --config ~/.config/ohmyposh/codex.omp.json
  oh-my-posh codex --session 019e9eac-83ec-7393-ae18-cc2e566394d5
  cat codex-status.json | oh-my-posh codex --config ~/.config/ohmyposh/codex.omp.json`,
	Args: cobra.NoArgs,
	Run: statuslineRunWithDataSource[segments.CodexData](
		shell.CODEX,
		cache.CODEXCACHE,
		func(d *segments.CodexData) string { return d.ThreadID },
		config.Codex,
		codexStatusDataSource,
	),
}

func init() {
	codexCmd.Flags().String("session", "", "Codex session/thread ID to render; defaults to the newest session with token usage")
	codexCmd.Flags().String("codex-home", "", "Codex home directory; defaults to CODEX_HOME or ~/.codex")
	codexCmd.Flags().String("session-root", "", "Codex sessions directory; defaults to <codex-home>/sessions")
	RootCmd.AddCommand(codexCmd)
}
