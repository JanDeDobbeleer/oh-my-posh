package cli

import (
	"fmt"
	"oh-my-posh/engine"
	"oh-my-posh/platform"
	"oh-my-posh/shell"

	"github.com/spf13/cobra"
)

var (
	print  bool
	strict bool
	manual bool

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
	initCmd.Flags().BoolVarP(&manual, "manual", "m", false, "enable/disable manual mode")
	_ = initCmd.MarkPersistentFlagRequired("config")
	RootCmd.AddCommand(initCmd)
}

func runInit(shellName string) {
	env := &platform.Shell{
		Version: cliVersion,
		CmdFlags: &platform.Flags{
			Shell:  shellName,
			Config: config,
			Strict: strict,
			Manual: manual,
		},
	}
	env.Init()
	defer env.Close()
	cfg := engine.LoadConfig(env)
	shell.Transient = cfg.TransientPrompt != nil
	shell.ErrorLine = cfg.ErrorLine != nil || cfg.ValidLine != nil
	shell.Tooltips = len(cfg.Tooltips) > 0
	for _, block := range cfg.Blocks {
		if block.Type == engine.RPrompt {
			shell.RPrompt = true
		}
	}
	if print {
		init := shell.PrintInit(env)
		fmt.Print(init)
		return
	}
	init := shell.Init(env)
	fmt.Print(init)
}
