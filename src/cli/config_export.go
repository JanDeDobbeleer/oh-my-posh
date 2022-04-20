package cli

import (
	"fmt"
	"oh-my-posh/engine"
	"oh-my-posh/environment"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	output string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your configuration",
	Long: `Export your configuration
You can choose to print the output to stdout, or export your configuration in the format of your choice.

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
			Version: cliVersion,
			CmdFlags: &environment.Flags{
				Config: config,
			},
		}
		env.Init(false)
		defer env.Close()
		cfg := engine.LoadConfig(env)
		if len(output) == 0 {
			fmt.Print(cfg.Export(format))
			return
		}
		cfg.Output = cleanOutputPath(output, env)
		format := strings.TrimPrefix(filepath.Ext(output), ".")
		if format == "yml" {
			format = engine.YAML
		}
		cfg.Write(format)
	},
}

func cleanOutputPath(path string, env environment.Environment) string {
	if strings.HasPrefix(path, "~") {
		path = strings.TrimPrefix(path, "~")
		path = filepath.Join(env.Home(), path)
	}
	if !filepath.IsAbs(path) {
		if absConfigFile, err := filepath.Abs(path); err == nil {
			path = absConfigFile
		}
	}
	return filepath.Clean(path)
}

func init() { // nolint:gochecknoinits
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "configuration format to migrate to")
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "config file to export to")
	configCmd.AddCommand(exportCmd)
}
