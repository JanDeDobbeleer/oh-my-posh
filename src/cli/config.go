/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cli

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [export|migrate|get]",
	Short: "Interact with the configuration",
	Long: `Interact with the configuration
It allows to export, migrate or get a configuration value.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(configCmd)
}
