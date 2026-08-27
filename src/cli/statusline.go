package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

func statuslineRun[T any](shellConst, cacheKey string, sessionID, pwd func(*T) string, defaultCfg func() *config.Config) func(*cmdtree.Command, []string) {
	return func(cmd *cmdtree.Command, _ []string) {
		// Only use a config file when --config was explicitly passed on the command line.
		// POSH_CONFIG is intentionally ignored here: statusline commands always render their
		// own dedicated layout and must not inherit the user's regular shell prompt config.
		explicitConfig := ""
		if cmd.Root().PersistentFlags().Changed("config") {
			explicitConfig = configFlag
			log.Debugf("using explicit config: %s", explicitConfig)
		}

		runStatusline(os.Stdin, os.Stdout, explicitConfig, shellConst, cacheKey, sessionID, pwd, defaultCfg)
	}
}

func runStatusline[T any](in io.Reader, out io.Writer, explicitConfig, shellConst, cacheKey string,
	sessionID, pwd func(*T) string, defaultCfg func() *config.Config,
) {
	log.Debugf("%s command started", shellConst)

	stdinData, err := io.ReadAll(in)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("received data from stdin: %s", string(stdinData))

	data := processStatuslineData(stdinData, shellConst, cacheKey, sessionID)

	flags := statuslineFlags(explicitConfig, shellConst, data, pwd)

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

	fmt.Fprint(out, eng.Status())
}

func statuslinePWD(candidate string) string {
	if candidate == "" {
		return ""
	}

	if !filepath.IsAbs(candidate) {
		log.Debugf("ignoring relative payload directory: %s", candidate)
		return ""
	}

	info, err := os.Stat(candidate)
	if err != nil || !info.IsDir() {
		log.Debugf("ignoring unusable payload directory: %s", candidate)
		return ""
	}

	return candidate
}

func workingDirectory(currentDir, cwd string) string {
	if currentDir != "" {
		return currentDir
	}

	return cwd
}

func statuslineFlags[T any](explicitConfig, shellConst string, data *T, pwd func(*T) string) *runtime.Flags {
	flags := &runtime.Flags{
		ConfigPath: explicitConfig,
		Shell:      shellConst,
	}

	if data == nil {
		return flags
	}

	flags.PWD = statuslinePWD(pwd(data))

	return flags
}

func processStatuslineData[T any](stdinData []byte, shellConst, cacheKey string, sessionID func(*T) string) *T {
	if len(stdinData) == 0 {
		cache.Init(shellConst, cache.Persist, cache.NoSession)
		return nil
	}

	var data T
	if err := json.Unmarshal(stdinData, &data); err != nil {
		log.Error(err)
		cache.Init(shellConst, cache.Persist, cache.NoSession)
		return nil
	}

	if id := sessionID(&data); id != "" {
		os.Setenv("POSH_SESSION_ID", id)
		log.Debugf("set POSH_SESSION_ID to: %s", id)
	}

	cache.Init(shellConst, cache.Persist)
	cache.Session.Set(cacheKey, data, cache.INFINITE)
	log.Debugf("stored %s data in session cache", shellConst)

	return &data
}
