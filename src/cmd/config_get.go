/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"oh-my-posh/environment"
	"time"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [shell|cache-path|millies]",
	Short: "Get a value from the oh-my-posh configuration",
	Long: `Get a value from the oh-my-posh configuration.
This command is used to get the value of the following variables:

- shell
- cache-path
- millis`,
	ValidArgs: []string{
		"millis",
		"shell",
		"cache-path",
	},
	Args: cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		env := &environment.ShellEnvironment{}
		env.Init(false)
		defer env.Close()
		switch args[0] {
		case "millis":
			fmt.Print(time.Now().UnixNano() / 1000000)
		case "shell":
			fmt.Println(env.Shell())
		case "cache-path":
			fmt.Print(env.CachePath())
		}
	},
}

func init() { // nolint:gochecknoinits
	configCmd.AddCommand(getCmd)
}
