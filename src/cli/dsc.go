package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/spf13/cobra"
)

var (
	state string
)

var dscCmd = &cobra.Command{
	Use:   "dsc [set|get|schema|export]",
	Short: "Desired State Configuration",
	Long: `Desired State Configuration.

Available commands:

- set: Set the desired state configuration
- get: Get the desired state configuration
- schema: Get the desired state configuration schema
- export: Export the desired state configuration`,
	ValidArgs: []string{
		"set",
		"get",
		"schema",
		"export",
		"test",
	},
	Args: cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		flags := &runtime.Flags{}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		var err error

		switch args[0] {
		case "set":
			if state == "" {
				err = fmt.Errorf("please provide a state configuration to set")
				break
			}

			err = dsc.Set(env.Cache(), state)
		case "schema":
			fmt.Print(dsc.Schema)
		case "export", "get":
			fmt.Print(dsc.Export(env.Cache()))
		case "test":
			if state == "" {
				err = fmt.Errorf("please provide a state configuration to test")
				break
			}

			err = dsc.Test(env.Cache(), state)
		default:
			_ = cmd.Help()
		}

		if err == nil {
			return
		}

		fmt.Println(dsc.Error(err))
		exitcode = 70
	},
}

func init() {
	dscCmd.Flags().StringVar(&state, "state", "", "State configuration to set")
	RootCmd.AddCommand(dscCmd)
}
