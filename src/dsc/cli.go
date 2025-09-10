package dsc

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/spf13/cobra"
)

var (
	state string
)

type resource interface {
	Load()
	Save()
	Resolve()
	ToJSON() string
	Schema() string
	Apply(schema string) error
	Test(input string) error
}

func Command(r resource) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "dsc",
		Short:     "Manage Oh My Posh DSC (Desired State Configuration)",
		Long:      "Manage Oh My Posh DSC (Desired State Configuration).",
		ValidArgs: []string{"get", "set", "test", "schema", "export"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			env := &runtime.Terminal{}
			env.Init(&runtime.Flags{})

			cache.Init(os.Getenv("POSH_SHELL"), false)

			defer func() {
				cache.Close()
			}()

			var err error

			switch args[0] {
			case "get", "export":
				r.Load()
				r.Resolve()
				fmt.Print(r.ToJSON())
			case "set":
				if state == "" {
					err = newError("please provide a state configuration to set")
					break
				}

				r.Load()
				err = r.Apply(state)
			case "schema":
				fmt.Print(r.Schema())
			case "test":
				if state == "" {
					err = newError("please provide a state configuration to test")
					break
				}

				r.Load()
				err = r.Test(state)
			default:
				_ = cmd.Help()
				return
			}

			if err != nil {
				fmt.Println(err.Error())
				return
			}
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "State configuration to set")
	return cmd
}
