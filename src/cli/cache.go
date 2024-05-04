package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"

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
		env := &platform.Shell{
			CmdFlags: &platform.Flags{},
		}
		env.Init()
		defer env.Close()
		switch args[0] {
		case "path":
			fmt.Print(env.CachePath())
		case "clear":
			cacheFilePath := filepath.Join(env.CachePath(), platform.CacheFile)
			err := os.Remove(cacheFilePath)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("removed cache file at %s\n", cacheFilePath)
		case "edit":
			cacheFilePath := filepath.Join(env.CachePath(), platform.CacheFile)
			editFileWithEditor(cacheFilePath)
		}
	},
}

func init() {
	RootCmd.AddCommand(getCache)
}

func editFileWithEditor(file string) {
	editor := os.Getenv("EDITOR")
	var args []string
	if strings.Contains(editor, " ") {
		splitted := strings.Split(editor, " ")
		editor = splitted[0]
		args = splitted[1:]
	}
	args = append(args, file)
	cmd := exec.Command(editor, args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
