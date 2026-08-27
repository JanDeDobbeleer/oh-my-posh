package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var antigravityCmd = &cmdtree.Command{
	Use:   "antigravity",
	Short: "Render a prompt for Antigravity CLI statusline",
	Long: `Render a prompt for Antigravity CLI statusline integration.

This command reads Antigravity CLI's contextual JSON data from stdin and renders
a prompt that can include an Antigravity segment with session information like
model name, token usage, and more.

Example usage in Antigravity CLI settings (~/.gemini/antigravity-cli/settings.json):
  {
    "statusLine": {
      "type": "command",
      "command": "oh-my-posh antigravity --config ~/.config/ohmyposh/antigravity.toml"
    }
  }`,
	Args: cmdtree.NoArgs,
	Run: statuslineRun(
		shell.ANTIGRAVITY,
		cache.ANTIGRAVITYCACHE,
		func(d *segments.AntigravityData) string { return d.SessionID },
		antigravityPWD,
		config.Antigravity,
	),
}

func init() {
	RootCmd.AddCommand(antigravityCmd)
}

func antigravityPWD(d *segments.AntigravityData) string {
	return workingDirectory(d.Workspace.CurrentDir, d.CWD)
}
