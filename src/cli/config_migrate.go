package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

var (
	write  bool
	format string
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate your config",
	Long: `Migrate your config.

You can choose to print the output to stdout, or migrate your config in the format of your choice.

Example usage

> oh-my-posh config migrate --config ~/myconfig.omp.json

Migrates the ~/myconfig.omp.json config file and prints the result to stdout.

> oh-my-posh config migrate --config ~/myconfig.omp.json --format toml

Migrates the ~/myconfig.omp.json config file to TOML and prints the result to stdout.

> oh-my-posh config migrate --config ~/myconfig.omp.json --format toml --write

Migrates the ~/myconfig.omp.json config file to TOML and writes the result to your config file.

A backup of the current config can be found at ~/myconfig.omp.json.bak.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		configFile := config.Path(configFlag)
		cfg := config.Load(configFile, shell.GENERIC, true)

		flags := &runtime.Flags{
			Config:  configFile,
			Migrate: true,
		}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		if write {
			cfg.BackupAndMigrate()
			return
		}
		cfg.Migrate()
		fmt.Print(cfg.Export(format))
	},
}

func init() {
	migrateCmd.Flags().BoolVarP(&write, "write", "w", false, "write the migrated config back to the config file")
	migrateCmd.Flags().StringVarP(&format, "format", "f", "json", "the config format to migrate to")
	configCmd.AddCommand(migrateCmd)
}
