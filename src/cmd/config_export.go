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

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your configuration",
	Long: `Export your configuration
You can choose to print the output to stdout, or export your configuration in the format of you choice.

Example usage

> oh-my-posh config export --config ~/myconfig.omp.json

Exports the ~/myconfig.omp.json config file and prints the result to stdout.

> oh-my-posh config export --config ~/myconfig.omp.json --format toml

Exports the ~/myconfig.omp.json config file to toml and prints the result to stdout.

> oh-my-posh config export --config ~/myconfig.omp.json --format toml --write

Exports the  ~/myconfig.omp.json config file to toml and writes the result to your config file.
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
			cfg.Write()
			return
		}
		fmt.Print(cfg.Export(format))
	},
}

func init() { // nolint:gochecknoinits
	exportCmd.Flags().BoolVarP(&write, "write", "w", false, "write the migrated configuration back to the config file")
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "configuration format to migrate to")
	configCmd.AddCommand(exportCmd)
}
