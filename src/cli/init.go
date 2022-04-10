/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cli

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/shell"

	"github.com/spf13/cobra"
)

var (
	print bool

	initCmd = &cobra.Command{
		Use:   "init [bash|zsh|fish|powershell|pwsh|cmd|nu] --config ~/.mytheme.omp.json",
		Short: "Initialize your shell and configuration",
		Long: `Allows to initialize your shell and configuration.
See the documentation to initialize your shell: https://ohmyposh.dev/docs/prompt.`,
		ValidArgs: []string{
			"bash",
			"zsh",
			"fish",
			"powershell",
			"pwsh",
			"cmd",
			"nu",
		},
		Args: cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			runInit(args[0])
		},
	}
)

func init() { // nolint:gochecknoinits
	initCmd.Flags().BoolVarP(&print, "print", "p", false, "print the init script")
	_ = initCmd.MarkPersistentFlagRequired("config")
	rootCmd.AddCommand(initCmd)
}

func runInit(shellName string) {
	env := &environment.ShellEnvironment{
		Version: cliVersion,
		CmdFlags: &environment.Flags{
			Shell:  shellName,
			Config: config,
		},
	}
	env.Init(false)
	defer env.Close()
	if print {
		init := shell.PrintInit(env)
		fmt.Print(init)
		return
	}
	init := shell.Init(env)
	fmt.Print(init)
}
