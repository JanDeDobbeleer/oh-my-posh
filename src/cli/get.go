package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	color2 "github.com/gookit/color"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [shell|millis|accent|toggles|width]",
	Short: "Get a value from oh-my-posh",
	Long: `Get a value from oh-my-posh.

This command is used to get the value of the following variables:

- shell
- millis
- accent
- toggles
- width`,
	ValidArgs: []string{
		"millis",
		"shell",
		"accent",
		"toggles",
		"width",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		if args[0] == "millis" {
			fmt.Print(time.Now().UnixNano() / 1000000)
			return
		}

		flags := &runtime.Flags{
			Shell: shellName,
		}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		switch args[0] {
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
		case "toggles":
			togglesCache, _ := env.Session().Get(cache.TOGGLECACHE)
			var toggles []string
			if len(togglesCache) != 0 {
				toggles = strings.Split(togglesCache, ",")
			}
			if len(toggles) == 0 {
				fmt.Println("No segments are toggled off")
				return
			}
			fmt.Println("Toggled off segments:")
			for _, toggle := range toggles {
				fmt.Println("- " + toggle)
			}
		case "width":
			width, err := env.TerminalWidth()
			if err != nil {
				fmt.Println("error getting terminal width:", err.Error())
				return
			}
			fmt.Println(width)
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
}
