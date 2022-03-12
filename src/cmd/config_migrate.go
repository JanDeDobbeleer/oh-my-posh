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
	write  bool
	format string
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate your configuration",
	Long: `Migrate your configuration
You can choose to print the output to stdout, or migrate your configuration in the format of you choice.

Example usage

> oh-my-posh config migrate --config ~/myconfig.omp.json

Migrates the ~/myconfig.omp.json config file and prints the result to stdout.

> oh-my-posh config migrate --config ~/myconfig.omp.json --format toml

Migrates the ~/myconfig.omp.json config file to toml and prints the result to stdout.

> oh-my-posh config migrate --config ~/myconfig.omp.json --format toml --write

Migrates the  ~/myconfig.omp.json config file to toml and writes the result to your config file.
A backup of the current config can be found at ~/myconfig.omp.json.bak.`,
	Run: func(cmd *cobra.Command, args []string) {
		env := &environment.ShellEnvironment{
			CmdFlags: &environment.Flags{
				Config: config,
			},
		}
		env.Init(false)
		defer env.Close()
		cfg := engine.LoadConfig(env)
		if write {
			cfg.BackupAndMigrate(env)
			return
		}
		cfg.Migrate(env)
		fmt.Print(cfg.Export(format))
	},
}

func init() { //nolint:gochecknoinits
	migrateCmd.Flags().BoolVarP(&write, "write", "w", false, "write the migrated configuration back to the config file")
	migrateCmd.Flags().StringVarP(&format, "format", "f", "json", "the configuration format to migrate to")
	configCmd.AddCommand(migrateCmd)
}
