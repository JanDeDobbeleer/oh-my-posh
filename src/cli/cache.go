/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cli

import (
	"fmt"
	"oh-my-posh/environment"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCache = &cobra.Command{
	Use:   "cache [path|clear]",
	Short: "Interact with the oh-my-posh cache",
	Long: `Interact with the oh-my-posh cache.
You can do the following:

- path: list the cache path
- clear: remove all cache values`,
	ValidArgs: []string{
		"path",
		"clear",
	},
	Args: cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		env := &environment.ShellEnvironment{
			Version: cliVersion,
		}
		env.Init(false)
		defer env.Close()
		switch args[0] {
		case "path":
			fmt.Print(env.CachePath())
		case "clear":
			cacheFilePath := filepath.Join(env.CachePath(), environment.CacheFile)
			err := os.Remove(cacheFilePath)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("removed cache file at %s\n", cacheFilePath)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(getCache)
}
