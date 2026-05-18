package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

const copilotServiceName = "copilot"

var copilotCmd = &cobra.Command{
	Use:   copilotServiceName,
	Short: "Render a prompt for GitHub Copilot CLI statusline",
	Long: `Render a prompt for GitHub Copilot CLI statusline integration.

This command reads GitHub Copilot CLI's contextual JSON data from stdin and renders
a prompt that can include a Copilot CLI segment with session information like
model name, token usage, costs, and more.

Example usage in GitHub Copilot CLI settings (%USERPROFILE%\.copilot\statusline.cmd):
  @echo off
  oh-my-posh copilot --config %USERPROFILE%\.config\ohmyposh\copilot.toml`,
	Args: cobra.NoArgs,
	Run: statuslineRun(
		shell.COPILOTCLI,
		cache.COPILOTCLICACHE,
		func(d *segments.CopilotCLIData) string { return d.SessionID },
		config.CopilotCLI,
	),
}

func init() {
	RootCmd.AddCommand(copilotCmd)
}
