package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var claudeCmd = &cmdtree.Command{
	Use:   "claude",
	Short: "Render a prompt for Claude Code statusline",
	Long: `Render a prompt for Claude Code statusline integration.

This command reads Claude Code's contextual JSON data from stdin and renders
a prompt that can include a Claude segment with session information like
model name, costs, tokens, and more.

Example usage in Claude Code settings:
  "statusLine": {
    "command": "oh-my-posh claude --config ~/.config/ohmyposh/claude.toml"
  }`,
	Args: cmdtree.NoArgs,
	Run: statuslineRun[segments.ClaudeData](
		shell.CLAUDE,
		cache.CLAUDECACHE,
		func(d *segments.ClaudeData) string { return d.SessionID },
		claudePWD,
		config.Claude,
	),
}

func init() {
	RootCmd.AddCommand(claudeCmd)
}

func claudePWD(d *segments.ClaudeData) string {
	return workingDirectory(d.Workspace.CurrentDir, d.CWD)
}
