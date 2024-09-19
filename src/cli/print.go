package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/spf13/cobra"
)

var (
	pwd           string
	pswd          string
	status        int
	pipestatus    string
	timing        float64
	stackCount    int
	terminalWidth int
	eval          bool
	cleared       bool
	cached        bool
	jobCount      int
	saveCache     bool

	command      string
	shellVersion string
	plain        bool
	noStatus     bool
	column       int
)

// printCmd represents the prompt command
var printCmd = createPrintCmd()

func init() {
	RootCmd.AddCommand(printCmd)
}

func createPrintCmd() *cobra.Command {
	printCmd := &cobra.Command{
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

			flags := &runtime.Flags{
				Config:        configFlag,
				PWD:           pwd,
				PSWD:          pswd,
				ErrorCode:     status,
				PipeStatus:    pipestatus,
				ExecutionTime: timing,
				StackCount:    stackCount,
				TerminalWidth: terminalWidth,
				Eval:          eval,
				Shell:         shellName,
				ShellVersion:  shellVersion,
				Plain:         plain,
				Primary:       args[0] == "primary",
				Cleared:       cleared,
				NoExitCode:    noStatus,
				Column:        column,
				JobCount:      jobCount,
				SaveCache:     saveCache,
			}

			eng := prompt.New(flags)
			defer eng.Env.Close()

			switch args[0] {
			case "debug":
				fmt.Print(eng.ExtraPrompt(prompt.Debug))
			case "primary":
				fmt.Print(eng.Primary())
			case "secondary":
				fmt.Print(eng.ExtraPrompt(prompt.Secondary))
			case "transient":
				fmt.Print(eng.ExtraPrompt(prompt.Transient))
			case "right":
				fmt.Print(eng.RPrompt())
			case "tooltip":
				fmt.Print(eng.Tooltip(command))
			case "valid":
				fmt.Print(eng.ExtraPrompt(prompt.Valid))
			case "error":
				fmt.Print(eng.ExtraPrompt(prompt.Error))
			default:
				_ = cmd.Help()
			}
		},
	}

	printCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	printCmd.Flags().StringVar(&pswd, "pswd", "", "current working directory (according to pwsh)")
	printCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
	printCmd.Flags().StringVar(&shellVersion, "shell-version", "", "the shell version")
	printCmd.Flags().IntVar(&status, "status", 0, "last known status code")
	printCmd.Flags().BoolVar(&noStatus, "no-status", false, "no valid status code (cancelled or no command yet)")
	printCmd.Flags().StringVar(&pipestatus, "pipestatus", "", "the PIPESTATUS array")
	printCmd.Flags().Float64Var(&timing, "execution-time", 0, "timing of the last command")
	printCmd.Flags().IntVarP(&stackCount, "stack-count", "s", 0, "number of locations on the stack")
	printCmd.Flags().IntVarP(&terminalWidth, "terminal-width", "w", 0, "width of the terminal")
	printCmd.Flags().StringVar(&command, "command", "", "tooltip command")
	printCmd.Flags().BoolVarP(&plain, "plain", "p", false, "plain text output (no ANSI)")
	printCmd.Flags().BoolVar(&cleared, "cleared", false, "do we have a clear terminal or not")
	printCmd.Flags().BoolVar(&eval, "eval", false, "output the prompt for eval")
	printCmd.Flags().IntVar(&column, "column", 0, "the column position of the cursor")
	printCmd.Flags().IntVar(&jobCount, "job-count", 0, "number of background jobs")
	printCmd.Flags().BoolVar(&saveCache, "save-cache", false, "save updated cache to file")

	// Deprecated flags, should be kept to avoid breaking CLI integration.
	printCmd.Flags().IntVarP(&status, "error", "e", 0, "last exit code")
	printCmd.Flags().BoolVar(&noStatus, "no-exit-code", false, "no valid exit code (cancelled or no command yet)")
	printCmd.Flags().BoolVar(&cached, "cached", false, "use a cached prompt")

	// Hide flags that are deprecated or for internal use only.
	_ = printCmd.Flags().MarkHidden("error")
	_ = printCmd.Flags().MarkHidden("no-exit-code")
	_ = printCmd.Flags().MarkHidden("cached")
	_ = printCmd.Flags().MarkHidden("save-cache")

	return printCmd
}
