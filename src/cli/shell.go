package cli

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
	basedsc "github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

var shellCmd = &cmdtree.Command{
	Use:   "shell get",
	Short: "Get the shell name",
	Long: `Get the shell name.

This command retrieves the name of the current shell being used.`,
	Example: `  oh-my-posh shell get`,
	ValidArgs: []string{
		"get",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cmdtree.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		flags := &runtime.Flags{
			Shell: os.Getenv("POSH_SHELL"),
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		switch args[0] {
		case "get":
			fmt.Print(env.Shell())
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	shellCmd.AddCommand(basedsc.Command(dsc.ShellDSC()))
	RootCmd.AddCommand(shellCmd)
}
