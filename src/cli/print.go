package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

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
			prompt.DEBUG,
			prompt.PRIMARY,
			prompt.SECONDARY,
			prompt.TRANSIENT,
			prompt.RIGHT,
			prompt.TOOLTIP,
			prompt.VALID,
			prompt.ERROR,
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
				Type:          args[0],
				Cleared:       cleared,
				NoExitCode:    noStatus,
				Column:        column,
				JobCount:      jobCount,
				IsPrimary:     args[0] == prompt.PRIMARY,
				SaveCache:     saveCache,
			}

			eng := prompt.New(flags)

			defer func() {
				template.SaveCache()
				eng.Env.Close()
			}()

			switch args[0] {
			case prompt.DEBUG:
				fmt.Print(eng.ExtraPrompt(prompt.Debug))
			case prompt.PRIMARY:
				fmt.Print(eng.Primary())
			case prompt.SECONDARY:
				fmt.Print(eng.ExtraPrompt(prompt.Secondary))
			case prompt.TRANSIENT:
				fmt.Print(eng.ExtraPrompt(prompt.Transient))
			case prompt.RIGHT:
				fmt.Print(eng.RPrompt())
			case prompt.TOOLTIP:
				fmt.Print(eng.Tooltip(command))
			case prompt.VALID:
				fmt.Print(eng.ExtraPrompt(prompt.Valid))
			case prompt.ERROR:
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

	// Hide flags that are for internal use only.
	_ = printCmd.Flags().MarkHidden("save-cache")

	return printCmd
}
