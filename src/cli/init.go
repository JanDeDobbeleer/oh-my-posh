package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	printOutput bool
	strict      bool
	debug       bool
	daemonMode  bool

	supportedShells = []string{
		"bash",
		"zsh",
		"fish",
		"powershell",
		"pwsh",
		"cmd",
		"nu",
		"elvish",
		"xonsh",
	}

	initCmd = createInitCmd()
)

func init() {
	RootCmd.AddCommand(initCmd)
}

func createInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [bash|zsh|fish|powershell|pwsh|cmd|nu|elvish|xonsh]",
		Short: "Initialize your shell and config",
		Long: `Initialize your shell and config.

See the documentation to initialize your shell: https://ohmyposh.dev/docs/installation/prompt.`,
		ValidArgs: supportedShells,
		Args:      NoArgsOrOneValidArg,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			runInit(args[0], getFullCommand(cmd, args))
		},
	}

	initCmd.Flags().BoolVarP(&printOutput, "print", "p", false, "print the init script")
	initCmd.Flags().BoolVarP(&strict, "strict", "s", false, "run in strict mode")
	initCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	initCmd.Flags().BoolVar(&eval, "eval", false, "output the full init script for eval")
	initCmd.Flags().BoolVar(&daemonMode, "daemon", false, "enable daemon mode")

	_ = initCmd.MarkPersistentFlagRequired("config")

	return initCmd
}

func runInit(sh, command string) {
	if os.Getenv("CURSOR_AGENT") == "1" {
		log.Errorf("oh-my-posh init is disabled when running inside Cursor agent mode")
		return
	}

	if debug {
		log.Enable(plain)
	}

	if sh == "powershell" {
		sh = shell.PWSH
	}

	cfg := config.Load(configFlag)

	// do not persist when in daemon mode
	persist := !daemonMode
	initCache(sh, persist)

	flags := &runtime.Flags{
		Shell:      sh,
		ConfigPath: cfg.Source,
		ConfigHash: cfg.Hash(),
		Strict:     strict,
		Debug:      debug,
		Init:       true,
		Eval:       eval,
		Plain:      plain,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	template.Init(env, cfg.Var, cfg.Maps)

	defer func() {
		if persist {
			cfg.Store()
		}
		template.SaveCache()
		if err := cache.Clear(false, shell.InitScriptName(env.Flags())); err != nil {
			log.Error(err)
		}
		cache.Close()
	}()

	feats := cfg.Features(env, daemonMode)

	var output string

	switch {
	case debug:
		output = shell.Debug(env, feats, &startTime)
	case printOutput:
		output = shell.Script(env, feats)
	default:
		output = shell.Init(env, feats)
	}

	shellDSC := shell.DSC()
	shellDSC.Load()
	shellDSC.Add(&shell.Shell{
		Command: command,
		Name:    sh,
	})
	shellDSC.Save()

	if silent {
		return
	}

	fmt.Print(output)
}

func getFullCommand(cmd *cobra.Command, args []string) string {
	// Start with the command path
	cmdPath := cmd.CommandPath()

	// Add arguments
	if len(args) > 0 {
		cmdPath += " " + strings.Join(args, " ")
	}

	// Add flags that were actually set
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}

		if flag.Value.Type() == "bool" && flag.Value.String() == "true" {
			cmdPath += fmt.Sprintf(" --%s", flag.Name)
			return
		}

		if flag.Name == "config" {
			configPath := filepath.Clean(flag.Value.String())
			configPath = strings.ReplaceAll(configPath, path.Home(), "~")
			cmdPath += fmt.Sprintf(" --%s=%s", flag.Name, configPath)
			return
		}

		cmdPath += fmt.Sprintf(" --%s=%s", flag.Name, flag.Value.String())
	})

	return cmdPath
}

func initCache(sh string, persist bool) {
	if !persist {
		cache.Init(sh, cache.NoSession)
		return
	}

	options := []cache.Option{cache.NewSession, cache.Persist}

	switch {
	case !printOutput:
		if (eval && sh == shell.PWSH) || sh == shell.ELVISH {
			cache.Init(sh)
			return
		}

		fallthrough
	default:
		cache.Init(sh, options...)
	}
}
