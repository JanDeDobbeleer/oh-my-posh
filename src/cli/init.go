package cli

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/spf13/cobra"
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
		"tcsh",
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

			var startTime time.Time
			if debug {
				startTime = time.Now()
			}

			env := &runtime.Terminal{
				CmdFlags: &runtime.Flags{
					Shell:  shellName,
					Config: configFlag,
					Strict: strict,
					Debug:  debug,
				},
			}

			env.Init()
			defer env.Close()

			cfg := config.Load(env)

			feats := cfg.Features()

			var output string

			switch {
			case printOutput, debug:
				output = shell.PrintInit(env, feats, &startTime)
			default:
				output = shell.Init(env, feats)
			}

			if silent {
				return
			}

			fmt.Print(output)
		},
	}

	initCmd.Flags().BoolVarP(&printOutput, "print", "p", false, "print the init script")
	initCmd.Flags().BoolVarP(&strict, "strict", "s", false, "run in strict mode")
	initCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")

	_ = initCmd.MarkPersistentFlagRequired("config")

	return initCmd
}
