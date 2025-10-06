package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"

	"github.com/spf13/cobra"
)

var (
	session bool
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache [path|clear|ttl|show]",
	Short: "Interact with the oh-my-posh cache",
	Long: `Interact with the oh-my-posh cache.

You can do the following:

- path: list cache path
- clear: remove all cache values
- ttl: get cache TTL in days
- show: print a detailed list of all cached values`,
	ValidArgs: []string{
		"path",
		"clear",
		cache.TTL,
		"show",
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
		case "show":
			cache.Init(os.Getenv("POSH_SHELL"))
			store := cache.Device
			if session {
				store = cache.Session
			}

			fmt.Println(cache.Print(store))
		}
	},
}

func init() {
	cacheCmd.Flags().BoolVarP(&session, "session", "s", false, "show the session cache")
	RootCmd.AddCommand(cacheCmd)
}
