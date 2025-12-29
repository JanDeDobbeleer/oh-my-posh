package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
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

		// Process Claude data and initialize cache
		processClaudeData(stdinData)

		flags := &runtime.Flags{
			ConfigPath: configFlag,
			Shell:      shell.CLAUDE,
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		var cfg *config.Config

		cfg, err = config.Parse(configFlag)
		if err != nil {
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

		defer func() {
			template.SaveCache()
			cache.Close()
		}()

		result := eng.Status()
		fmt.Print(result)
	},
}

// processClaudeData handles parsing and caching of Claude JSON data
func processClaudeData(stdinData []byte) {
	if len(stdinData) == 0 {
		cache.Init(shell.CLAUDE, cache.Persist, cache.NoSession)
		return
	}

	var claudeData segments.ClaudeData
	if err := json.Unmarshal(stdinData, &claudeData); err != nil {
		log.Error(err)
		cache.Init(shell.CLAUDE, cache.Persist, cache.NoSession)
		return
	}

	log.Debugf("parsed Claude data: session_id=%s, model=%s", claudeData.SessionID, claudeData.Model.DisplayName)

	// Set the session ID from Claude data if available
	if claudeData.SessionID != "" {
		os.Setenv("POSH_SESSION_ID", claudeData.SessionID)
		log.Debugf("set POSH_SESSION_ID to: %s", claudeData.SessionID)
	}

	// Initialize cache first so we can store the data
	cache.Init(shell.CLAUDE, cache.Persist)

	// Store the parsed data in session cache
	cache.Set(cache.Session, cache.CLAUDECACHE, claudeData, cache.INFINITE)
	log.Debug("stored Claude data in session cache")
}

func init() {
	RootCmd.AddCommand(claudeCmd)
}
