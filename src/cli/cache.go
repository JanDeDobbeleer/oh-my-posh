package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"

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
- ttl: get/set cache TTL in days
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
			// Try daemon first
			if ipc.SocketExists() {
				if cleared := clearDaemonCache(); cleared {
					return
				}
			}
			// Fallback to CLI mode
			err := cache.Clear(true)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("cache cleared")
		case cache.TTL:
			if len(args) < 2 {
				// GET TTL
				if ipc.SocketExists() {
					if days, ok := getDaemonTTL(); ok {
						fmt.Printf("daemon TTL: %d days\n", days)
						return
					}
				}
				// Fallback to CLI mode - show default
				fmt.Printf("TTL: 7 days (default)\n")
				return
			}

			// SET TTL
			ttl, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("error parsing TTL:", err.Error())
				exitcode = 2
				return
			}

			// Try daemon first
			if ipc.SocketExists() {
				if set := setDaemonTTL(ttl); set {
					return
				}
			}
			// Fallback to CLI mode
			cache.Init(os.Getenv("POSH_SHELL"), cache.Persist)
			cache.Set(cache.Device, cache.TTL, ttl, cache.INFINITE)
			cache.Close()
			fmt.Printf("TTL set to %d days\n", ttl)
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

// clearDaemonCache attempts to clear the daemon cache.
// Returns true if successful, false if we should fall back to CLI mode.
func clearDaemonCache() bool {
	client, err := daemon.NewClient()
	if err != nil {
		return false
	}
	defer client.Close()

	if err := client.CacheClear(context.Background()); err != nil {
		return false
	}

	fmt.Println("daemon cache cleared")
	return true
}

// getDaemonTTL attempts to get the TTL from the daemon.
// Returns the days and true if successful, or 0 and false if we should fall back.
func getDaemonTTL() (int, bool) {
	client, err := daemon.NewClient()
	if err != nil {
		return 0, false
	}
	defer client.Close()

	days, err := client.CacheGetTTL(context.Background())
	if err != nil {
		return 0, false
	}

	return days, true
}

// setDaemonTTL attempts to set the TTL in the daemon.
// Returns true if successful, false if we should fall back to CLI mode.
func setDaemonTTL(days int) bool {
	client, err := daemon.NewClient()
	if err != nil {
		return false
	}
	defer client.Close()

	if err := client.CacheSetTTL(context.Background(), days); err != nil {
		return false
	}

	fmt.Printf("daemon TTL set to %d days\n", days)
	return true
}

func init() {
	cacheCmd.Flags().BoolVarP(&session, "session", "s", false, "show the session cache")
	RootCmd.AddCommand(cacheCmd)
}
