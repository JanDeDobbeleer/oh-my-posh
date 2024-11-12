package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCache = &cobra.Command{
	Use:   "cache [path|clear|edit]",
	Short: "Interact with the oh-my-posh cache",
	Long: `Interact with the oh-my-posh cache.

You can do the following:

- path: list cache path
- clear: remove all cache values
- edit: edit cache values`,
	ValidArgs: []string{
		"path",
		"clear",
		"edit",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		switch args[0] {
		case "path":
			fmt.Println(cache.Path())
		case "clear":
			deletedFiles, err := cache.Clear(cache.Path(), true)
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, file := range deletedFiles {
				fmt.Println("removed cache file:", file)
			}
		case "edit":
			cacheFilePath := filepath.Join(cache.Path(), cache.FileName)
			os.Exit(editFileWithEditor(cacheFilePath))
		}
	},
}

func init() {
	RootCmd.AddCommand(getCache)
}
