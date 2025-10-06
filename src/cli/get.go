package cli

import (
	"fmt"
	"os"
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
		cache.TTL,
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
			Shell: os.Getenv("POSH_SHELL"),
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		switch args[0] {
		case "shell":
			fmt.Print(env.Shell())
			return
		case "accent":
			rgb, err := color.GetAccentColor(env)
			if err != nil {
				fmt.Println("error getting accent color:", err.Error())
				return
			}
			accent := color2.RGB(rgb.R, rgb.G, rgb.B)
			fmt.Print("#" + accent.Hex())
			return
		case "width":
			width, err := env.TerminalWidth()
			if err != nil {
				fmt.Println("error getting terminal width:", err.Error())
				return
			}

			fmt.Print(width)
			return
		}

		cache.Init(env.Shell(), cache.Persist)

		defer func() {
			cache.Close()
		}()

		switch args[0] {
		case "toggles":
			togglesMap, _ := cache.Get[map[string]bool](cache.Session, cache.TOGGLECACHE)
			if len(togglesMap) == 0 {
				fmt.Println("No segments are toggled off")
				return
			}

			fmt.Println("Toggled off segments:")
			for toggle := range togglesMap {
				fmt.Println("- " + toggle)
			}
		case cache.TTL:
			fmt.Print(cache.GetTTL())
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}
