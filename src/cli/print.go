package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/spf13/cobra"
)

var (
	pwd           string
	pswd          string
	exitCode      int
	timing        float64
	stackCount    int
	terminalWidth int
	eval          bool
	cleared       bool

	command      string
	shellVersion string
	plain        bool
)

// printCmd represents the prompt command
var printCmd = &cobra.Command{
	Use:   "print [debug|primary|secondary|transient|right|tooltip|valid|error]",
	Short: "Print the prompt/context",
	Long:  "Print one of the prompts based on the location/use-case.",
	ValidArgs: []string{
		"debug",
		"primary",
		"secondary",
		"transient",
		"right",
		"tooltip",
		"valid",
		"error",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		flags := &platform.Flags{
			Config:        config,
			PWD:           pwd,
			PSWD:          pswd,
			ErrorCode:     exitCode,
			ExecutionTime: timing,
			StackCount:    stackCount,
			TerminalWidth: terminalWidth,
			Eval:          eval,
			Shell:         shellName,
			ShellVersion:  shellVersion,
			Plain:         plain,
			Primary:       args[0] == "primary",
			Cleared:       cleared,
		}

		eng := engine.New(flags)
		defer eng.Env.Close()

		switch args[0] {
		case "debug":
			fmt.Print(eng.PrintExtraPrompt(engine.Debug))
		case "primary":
			fmt.Print(eng.PrintPrimary())
		case "secondary":
			fmt.Print(eng.PrintExtraPrompt(engine.Secondary))
		case "transient":
			fmt.Print(eng.PrintExtraPrompt(engine.Transient))
		case "right":
			fmt.Print(eng.PrintRPrompt())
		case "tooltip":
			fmt.Print(eng.PrintTooltip(command))
		case "valid":
			fmt.Print(eng.PrintExtraPrompt(engine.Valid))
		case "error":
			fmt.Print(eng.PrintExtraPrompt(engine.Error))
		default:
			_ = cmd.Help()
		}
	},
}

func init() { //nolint:gochecknoinits
	printCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	printCmd.Flags().StringVar(&pswd, "pswd", "", "current working directory (according to pwsh)")
	printCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
	printCmd.Flags().StringVar(&shellVersion, "shell-version", "", "the shell version")
	printCmd.Flags().IntVarP(&exitCode, "error", "e", 0, "last exit code")
	printCmd.Flags().Float64Var(&timing, "execution-time", 0, "timing of the last command")
	printCmd.Flags().IntVarP(&stackCount, "stack-count", "s", 0, "number of locations on the stack")
	printCmd.Flags().IntVarP(&terminalWidth, "terminal-width", "w", 0, "width of the terminal")
	printCmd.Flags().StringVar(&command, "command", "", "tooltip command")
	printCmd.Flags().BoolVarP(&plain, "plain", "p", false, "plain text output (no ANSI)")
	printCmd.Flags().BoolVar(&cleared, "cleared", false, "do we have a clear terminal or not")
	printCmd.Flags().BoolVar(&eval, "eval", false, "output the prompt for eval")
	RootCmd.AddCommand(printCmd)
}
