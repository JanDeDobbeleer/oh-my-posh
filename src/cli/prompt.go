/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cli

import (
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Set up the prompt for your shell",
	Long: `Set up the prompt for your shell
Allows to initialize one of the supported shells, or to set the prompt manually for a custom shell.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(promptCmd)
}
