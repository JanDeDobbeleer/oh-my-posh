package cli

import (
	"fmt"
	"oh-my-posh/engine"
	"oh-my-posh/environment"
	"oh-my-posh/shell"

	"github.com/spf13/cobra"
)

var (
	print  bool
	strict bool

	initCmd = &cobra.Command{
		Use:   "init [bash|zsh|fish|powershell|pwsh|cmd|nu] --config ~/.mytheme.omp.json",
		Short: "Initialize your shell and config",
		Long: `Initialize your shell and config.

See the documentation to initialize your shell: https://ohmyposh.dev/docs/installation/prompt.`,
		ValidArgs: []string{
			"bash",
			"zsh",
			"fish",
			"powershell",
			"pwsh",
			"cmd",
			"nu",
		},
		Args: NoArgsOrOneValidArg,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			runInit(args[0])
		},
	}
)

func init() { //nolint:gochecknoinits
	initCmd.Flags().BoolVarP(&print, "print", "p", false, "print the init script")
	initCmd.Flags().BoolVarP(&strict, "strict", "s", false, "run in strict mode")
	_ = initCmd.MarkPersistentFlagRequired("config")
	rootCmd.AddCommand(initCmd)
}

func runInit(shellName string) {
	env := &environment.ShellEnvironment{
		Version: cliVersion,
		CmdFlags: &environment.Flags{
			Shell:  shellName,
			Config: config,
			Strict: strict,
		},
	}
	env.Init()
	defer env.Close()
	cfg := engine.LoadConfig(env)
	shell.Transient = cfg.TransientPrompt != nil
	shell.ErrorLine = cfg.ErrorLine != nil || cfg.ValidLine != nil
	shell.Tooltips = len(cfg.Tooltips) > 0
	if print {
		init := shell.PrintInit(env)
		fmt.Print(init)
		return
	}
	init := shell.Init(env)
	fmt.Print(init)
}
