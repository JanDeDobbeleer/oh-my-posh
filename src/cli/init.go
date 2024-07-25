package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"

	"github.com/spf13/cobra"
)

var (
	printOutput bool
	strict      bool
	manual      bool
	debug       bool

	supportedShells = []string{
		"bash",
		"zsh",
		"fish",
		"powershell",
		"pwsh",
		"cmd",
		"nu",
		"tcsh",
		"elvish",
		"xonsh",
	}

	initCmd = &cobra.Command{
		Use:   "init [bash|zsh|fish|powershell|pwsh|cmd|nu|tcsh|elvish|xonsh]",
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
			runInit(args[0])
		},
	}
)

func init() {
	initCmd.Flags().BoolVarP(&printOutput, "print", "p", false, "print the init script")
	initCmd.Flags().BoolVarP(&strict, "strict", "s", false, "run in strict mode")
	initCmd.Flags().BoolVarP(&manual, "manual", "m", false, "enable/disable manual mode")
	_ = initCmd.MarkPersistentFlagRequired("config")
	RootCmd.AddCommand(initCmd)
}

func runInit(shellName string) {
	env := &runtime.Terminal{
		CmdFlags: &runtime.Flags{
			Shell:  shellName,
			Config: configFlag,
			Strict: strict,
			Manual: manual,
		},
	}
	env.Init()
	defer env.Close()

	cfg := config.Load(env)

	// TODO: this can be removed I think
	// allow overriding the upgrade notice from the config
	if cfg.DisableNotice || cfg.AutoUpgrade {
		env.Cache().Set(upgrade.CACHEKEY, "disabled", -1)
	}

	feats := cfg.Features()

	if printOutput {
		init := shell.PrintInit(env, feats)
		fmt.Print(init)
		return
	}

	init := shell.Init(env, feats)
	fmt.Print(init)
}
