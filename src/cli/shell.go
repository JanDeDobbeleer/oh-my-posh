package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/spf13/cobra"
)

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell get",
	Short: "Get the shell name",
	Long: `Get the shell name.

This command retrieves the name of the current shell being used.`,
	Example: `  oh-my-posh shell get`,
	ValidArgs: []string{
		"get",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		flags := &runtime.Flags{
			Shell: shellName,
		}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		switch args[0] {
		case "get":
			fmt.Print(env.Shell())
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	shellCmd.AddCommand(dsc.Command(shell.DSC()))
	RootCmd.AddCommand(shellCmd)
}
