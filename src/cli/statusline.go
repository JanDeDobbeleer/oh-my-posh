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
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

// statuslineRun returns a cobra Run function for statusline commands.
// T is the type of the JSON data read from stdin.
// shellConst identifies the shell (used for cache init and terminal init).
// cacheKey is the session cache key under which data is stored.
// sessionID extracts the session ID from parsed data so it can be set as POSH_SESSION_ID.
// defaultCfg is called when no --config flag is provided or parsing fails.
func statuslineRun[T any](shellConst, cacheKey string, sessionID func(*T) string, defaultCfg func() *config.Config) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, _ []string) {
		log.Debugf("%s command started", shellConst)

		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Error(err)
			return
		}

		log.Debugf("received data from stdin: %s", string(stdinData))

		processStatuslineData(stdinData, shellConst, cacheKey, sessionID)

		// Only use a config file when --config was explicitly passed on the command line.
		// POSH_CONFIG is intentionally ignored here: statusline commands always render their
		// own dedicated layout and must not inherit the user's regular shell prompt config.
		explicitConfig := ""
		if cmd.Root().PersistentFlags().Changed("config") {
			explicitConfig = configFlag
			log.Debugf("using explicit config: %s", explicitConfig)
		}

		flags := &runtime.Flags{
			ConfigPath: explicitConfig,
			Shell:      shellConst,
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		cfg, err := config.Parse(explicitConfig)
		if err != nil {
			log.Debug("no config found, using default")
			cfg = defaultCfg()
		}

		template.Init(env, cfg.Var, cfg.Maps)
		terminal.Init(shellConst)
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

		fmt.Print(eng.Status())
	}
}

// processStatuslineData parses stdin JSON into T and stores it in the session cache.
func processStatuslineData[T any](stdinData []byte, shellConst, cacheKey string, sessionID func(*T) string) {
	if len(stdinData) == 0 {
		cache.Init(shellConst, cache.Persist, cache.NoSession)
		return
	}

	var data T
	if err := json.Unmarshal(stdinData, &data); err != nil {
		log.Error(err)
		cache.Init(shellConst, cache.Persist, cache.NoSession)
		return
	}

	if id := sessionID(&data); id != "" {
		os.Setenv("POSH_SESSION_ID", id)
		log.Debugf("set POSH_SESSION_ID to: %s", id)
	}

	cache.Init(shellConst, cache.Persist)
	cache.Set(cache.Session, cacheKey, data, cache.INFINITE)
	log.Debugf("stored %s data in session cache", shellConst)
}
