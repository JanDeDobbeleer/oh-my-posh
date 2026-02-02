package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

const (
	renderTimeout = 10 * time.Second
)

var pid int

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render prompts via the daemon",
	Long: `Render all prompts via the daemon for faster display.

The daemon computes segments asynchronously and streams updates.
After a short timeout (100ms), partial results are returned with
cached values for slow segments. Updates stream as segments complete.

Output format (one per line):
  primary:<text>
  right:<text>
  secondary:<text>
  ...

Falls back to direct rendering if daemon is not available.`,
	Run: func(_ *cobra.Command, _ []string) {
		if shellName == "" {
			shellName = shell.GENERIC
		}

		if configFlag != "" {
			configFlag = path.ReplaceTildePrefixWithHomeDir(configFlag)
			if abs, err := filepath.Abs(configFlag); err == nil {
				configFlag = abs
			}
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
			Shell:         shellName,
			ShellVersion:  shellVersion,
			Plain:         plain,
			Cleared:       cleared,
			NoExitCode:    noStatus,
			Column:        column,
			JobCount:      jobCount,
			Escape:        escape,
			Force:         force,
		}

		if err := renderViaDaemon(flags, pid); err != nil {
			log.Debugf("daemon render failed: %v, falling back to direct render", err)
			renderDirect(flags)
		}
	},
}

func init() {
	renderCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	renderCmd.Flags().StringVar(&pswd, "pswd", "", "current working directory (according to pwsh)")
	renderCmd.Flags().StringVar(&shellName, "shell", "", "the shell to render for")
	renderCmd.Flags().StringVar(&shellVersion, "shell-version", "", "the shell version")
	renderCmd.Flags().IntVar(&status, "status", 0, "last known status code")
	renderCmd.Flags().BoolVar(&noStatus, "no-status", false, "no valid status code (cancelled or no command yet)")
	renderCmd.Flags().StringVar(&pipestatus, "pipestatus", "", "the PIPESTATUS array")
	renderCmd.Flags().Float64Var(&timing, "execution-time", 0, "timing of the last command")
	renderCmd.Flags().IntVarP(&stackCount, "stack-count", "s", 0, "number of locations on the stack")
	renderCmd.Flags().IntVarP(&terminalWidth, "terminal-width", "w", 0, "width of the terminal")
	renderCmd.Flags().BoolVar(&cleared, "cleared", false, "do we have a clear terminal or not")
	renderCmd.Flags().IntVar(&column, "column", 0, "the column position of the cursor")
	renderCmd.Flags().IntVar(&jobCount, "job-count", 0, "number of background jobs")
	renderCmd.Flags().BoolVar(&escape, "escape", true, "escape the ANSI sequences for the shell")
	renderCmd.Flags().BoolVarP(&force, "force", "f", false, "force rendering the segments")
	renderCmd.Flags().IntVar(&pid, "pid", 0, "shell process id")

	RootCmd.AddCommand(renderCmd)
}

func renderViaDaemon(flags *runtime.Flags, pid int) error {
	silent = true
	client, err := daemon.ConnectOrStart(startDetachedDaemon)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), renderTimeout)
	defer cancel()

	return client.RenderPrompt(ctx, flags, pid, "", nil, func(resp *ipc.PromptResponse) bool {
		outputPrompts(resp)
		return resp.Type != "complete"
	})
}

func outputPrompts(resp *ipc.PromptResponse) {
	if resp == nil || resp.Prompts == nil {
		return
	}

	// Output each prompt with a prefix for shell parsing
	// Format: type:text (text can contain newlines, shell handles it)
	//
	// IMPORTANT: Always output primary and right prompts even if empty.
	// The shell keeps previous values if a prompt type isn't sent,
	// so we must send empty values to clear stale prompts (e.g., git segment
	// persisting after leaving a repo).
	alwaysOutput := map[string]bool{"primary": true, "right": true}
	promptTypes := []string{"primary", "right", "secondary", "transient", "debug", "valid", "error"}

	for _, pt := range promptTypes {
		if p, ok := resp.Prompts[pt]; ok {
			// Always output primary/right; only output others if non-empty
			if alwaysOutput[pt] || p.Text != "" {
				fmt.Printf("%s:%s\n", pt, p.Text)
			}
		}
	}

	// Output status line so shell knows when a batch is complete
	// "update" = more updates may come, "complete" = all segments done
	fmt.Printf("status:%s\n", resp.Type)
}

func renderDirect(flags *runtime.Flags) {
	silent = true
	flags.Eval = false
	cache.Init(flags.Shell)

	eng := prompt.New(flags)

	defer func() {
		eng.SaveTemplateCache()
		cache.Close()
	}()

	// Always output primary and right prompts (even if empty) so shell can clear stale values
	fmt.Printf("primary:%s\n", eng.Primary())
	fmt.Printf("right:%s\n", eng.RPrompt())
	fmt.Println("status:complete")
}
