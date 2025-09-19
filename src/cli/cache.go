package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCache = &cobra.Command{
	Use:   "cache [path|clear|ttl]",
	Short: "Interact with the oh-my-posh cache",
	Long: `Interact with the oh-my-posh cache.

You can do the following:

- path: list cache path
- clear: remove all cache values
- ttl: get cache TTL in days`,
	ValidArgs: []string{
		"path",
		"clear",
		cache.TTL,
	},
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		switch args[0] {
		case "path":
			fmt.Println(cache.Path())
		case "clear":
			err := cache.Clear(true)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("cache cleared")
		case cache.TTL:
			// get the second argument as int
			if len(args) < 2 {
				fmt.Println("please provide a TTL value in days")
				exitcode = 2
				return
			}

			ttl, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("error parsing TTL:", err.Error())
				exitcode = 2
				return
			}

			cache.Init(os.Getenv("POSH_SHELL"), cache.Persist)
			cache.Set(cache.Device, cache.TTL, ttl, cache.INFINITE)
			cache.Close()
		}
	},
}

func init() {
	RootCmd.AddCommand(getCache)
}
