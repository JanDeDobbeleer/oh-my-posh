package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
	initCmd.Flags().BoolVar(&eval, "eval", false, "output the prompt for eval")

	_ = initCmd.MarkPersistentFlagRequired("config")

	return initCmd
}

func runInit(sh, command string) {
	var startTime time.Time

	if debug {
		startTime = time.Now()
		log.Enable(plain)
	}

	cfg, hash := config.Load(configFlag, sh, false)

	flags := &runtime.Flags{
		Shell:     sh,
		Config:    cfg.Source,
		Strict:    strict,
		Debug:     debug,
		SaveCache: true,
		Init:      true,
		Eval:      eval,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	template.Init(env, cfg.Var, cfg.Maps)

	defer func() {
		template.SaveCache()
		env.Close()
	}()

	feats := cfg.Features(env)
	flags.ConfigHash = fmt.Sprintf("%s.%s", hash, feats.Hash())

	var output string

	switch {
	case printOutput, debug:
		output = shell.PrintInit(env, feats, &startTime)
	default:
		output = shell.Init(env, feats)
	}

	if !debug {
		configDSC := config.DSC()
		configDSC.Load(env.Cache())
		configDSC.Add(configFlag)
		configDSC.Save()

		shellDSC := shell.DSC()
		shellDSC.Load(env.Cache())
		shellDSC.Add(&shell.Shell{
			Command: command,
			Name:    sh,
		})
		shellDSC.Save()
	}

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
