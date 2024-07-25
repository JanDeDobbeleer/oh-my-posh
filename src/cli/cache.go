package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

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

		env := &runtime.Terminal{
			CmdFlags: &runtime.Flags{},
		}

		env.Init()
		defer env.Close()

		switch args[0] {
		case "path":
			fmt.Println(env.CachePath())
		case "clear":
			clear(env.CachePath())
		case "edit":
			cacheFilePath := filepath.Join(env.CachePath(), cache.FileName)
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

func clear(cachePath string) {
	// get all files in the cache directory that start with omp.cache and delete them
	files, err := os.ReadDir(cachePath)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), cache.FileName) {
			continue
		}

		path := filepath.Join(cachePath, file.Name())
		if err := os.Remove(path); err == nil {
			fmt.Println("removed cache file:", path)
		}
	}
}
