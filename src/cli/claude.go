package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

// claudeCmd represents the claude command
var claudeCmd = &cobra.Command{
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
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		log.Debug("claude command started")

		// Read JSON from stdin
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Error(err)
			return
		}

		log.Debugf("received data from stdin: %s", string(stdinData))

		// Store the JSON in an environment variable for the Claude segment to read
		if len(stdinData) > 0 {
			os.Setenv("POSH_CLAUDE_STATUS", string(stdinData))
		}

		flags := &runtime.Flags{
			ConfigPath: configFlag,
			Shell:      shell.CLAUDE,
		}

		cache.Init(shell.CLAUDE, cache.Persist, cache.NoSession)

		env := &runtime.Terminal{}
		env.Init(flags)

		var cfg *config.Config

		cfg, errCode := config.Parse(configFlag)
		if errCode != 0 {
			cfg = config.Claude()
		}

		template.Init(env, cfg.Var, cfg.Maps)
		terminal.Init(shell.CLAUDE)
		terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
		terminal.Colors = cfg.MakeColors(env)

		eng := &prompt.Engine{
			Config: cfg,
			Env:    env,
		}

		result := eng.Status()
		fmt.Print(result)
	},
}

func init() {
	RootCmd.AddCommand(claudeCmd)
}
