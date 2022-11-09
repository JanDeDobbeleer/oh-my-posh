package cli

import (
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/platform"
	"time"

	color2 "github.com/gookit/color"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [shell|millis|accent]",
	Short: "Get a value from oh-my-posh",
	Long: `Get a value from oh-my-posh.

This command is used to get the value of the following variables:

- shell
- millis
- accent`,
	ValidArgs: []string{
		"millis",
		"shell",
		"accent",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		env := &platform.Shell{
			Version: cliVersion,
			CmdFlags: &platform.Flags{
				Shell: shellName,
			},
		}
		env.Init()
		defer env.Close()
		switch args[0] {
		case "millis":
			fmt.Print(time.Now().UnixNano() / 1000000)
		case "shell":
			fmt.Println(env.Shell())
		case "accent":
			rgb, err := color.GetAccentColor(env)
			if err != nil {
				fmt.Println("error getting accent color:", err.Error())
				return
			}
			accent := color2.RGB(rgb.R, rgb.G, rgb.B)
			fmt.Println("#" + accent.Hex())
		default:
			_ = cmd.Help()
		}
	},
}

func init() { //nolint:gochecknoinits
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
}
