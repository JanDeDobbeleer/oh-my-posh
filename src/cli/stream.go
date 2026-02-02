package cli

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/spf13/cobra"
)

// streamCmd represents the stream command
var streamCmd = createStreamCmd()

func init() {
	RootCmd.AddCommand(streamCmd)
}

func createStreamCmd() *cobra.Command {
	streamCmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream the prompt with incremental updates",
		Long: `Stream the primary prompt with incremental updates as segments complete.
Output format: null-byte delimited prompt strings (each complete prompt separated by \0).
This allows multi-line prompts to be handled correctly.
The shell can read records incrementally and update the display.
Command exits when all segments are resolved.`,
		Args: cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			if shellName == "" {
				shellName = shell.GENERIC
			}

			flags := &runtime.Flags{
				ConfigPath:    configFlag,
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
				Type:          prompt.PRIMARY,
				Cleared:       cleared,
				NoExitCode:    noStatus,
				Column:        column,
				JobCount:      jobCount,
				IsPrimary:     true,
				Escape:        escape,
				Force:         force,
				Streaming:     true,
			}

			options := []cache.Option{}
			if saveCache {
				options = append(options, cache.Persist)
			}

			cache.Init(shellName, options...)

			eng := prompt.New(flags)

			defer func() {
				template.SaveCache()
				cache.Close()
			}()

			// Stream prompt updates
			for promptString := range eng.StreamPrimary() {
				fmt.Print(promptString)
				fmt.Print("\x00") // Null byte delimiter for multi-line prompts
				os.Stdout.Sync()  // Flush stdout to ensure PowerShell can read immediately
			}
		},
	}

	streamCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	streamCmd.Flags().StringVar(&pswd, "pswd", "", "current working directory (according to pwsh)")
	streamCmd.Flags().StringVar(&shellName, "shell", "", "the shell to stream for")
	streamCmd.Flags().StringVar(&shellVersion, "shell-version", "", "the shell version")
	streamCmd.Flags().IntVar(&status, "status", 0, "last known status code")
	streamCmd.Flags().BoolVar(&noStatus, "no-status", false, "no valid status code (cancelled or no command yet)")
	streamCmd.Flags().StringVar(&pipestatus, "pipestatus", "", "the PIPESTATUS array")
	streamCmd.Flags().Float64Var(&timing, "execution-time", 0, "timing of the last command")
	streamCmd.Flags().IntVarP(&stackCount, "stack-count", "s", 0, "number of locations on the stack")
	streamCmd.Flags().IntVarP(&terminalWidth, "terminal-width", "w", 0, "width of the terminal")
	streamCmd.Flags().BoolVar(&cleared, "cleared", false, "do we have a clear terminal or not")
	streamCmd.Flags().BoolVar(&eval, "eval", false, "output the prompt for eval")
	streamCmd.Flags().IntVar(&column, "column", 0, "the column position of the cursor")
	streamCmd.Flags().IntVar(&jobCount, "job-count", 0, "number of background jobs")
	streamCmd.Flags().BoolVar(&saveCache, "save-cache", false, "save updated cache to file")
	streamCmd.Flags().BoolVar(&escape, "escape", true, "escape the ANSI sequences for the shell")
	streamCmd.Flags().BoolVarP(&force, "force", "f", false, "force rendering the segments")

	// Hide flags that are for internal use only.
	_ = streamCmd.Flags().MarkHidden("save-cache")

	return streamCmd
}
