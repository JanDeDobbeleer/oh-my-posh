/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"oh-my-posh/engine"
	"oh-my-posh/environment"

	"github.com/spf13/cobra"
)

var (
	print bool

	initCmd = &cobra.Command{
		Use:   "init [bash|zsh|fish|powershell|pwsh|cmd] --config ~/.mytheme.omp.json",
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
		},
		Args: cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runInit(args[0])
		},
	}
)

func init() { // nolint:gochecknoinits
	initCmd.Flags().BoolVarP(&print, "print", "p", false, "print the init script")
	_ = initCmd.MarkPersistentFlagRequired("config")
	promptCmd.AddCommand(initCmd)
}

func runInit(shell string) {
	env := &environment.ShellEnvironment{
		CmdFlags: &environment.Flags{
			Shell:  shell,
			Config: config,
		},
	}
	env.Init(false)
	defer env.Close()
	if print {
		init := engine.PrintShellInit(env)
		fmt.Print(init)
		return
	}
	init := engine.InitShell(env)
	fmt.Print(init)
}
