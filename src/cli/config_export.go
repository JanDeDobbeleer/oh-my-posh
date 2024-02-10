package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/spf13/cobra"
)

var (
	output string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your config",
	Long: `Export your config.

You can choose to print the output to stdout, or export your config in the format of your choice.

Example usage:

> oh-my-posh config export --config ~/myconfig.omp.json

Exports the ~/myconfig.omp.json config file and prints the result to stdout.

> oh-my-posh config export --config ~/myconfig.omp.json --format toml

Exports the ~/myconfig.omp.json config file to toml and prints the result to stdout.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		env := &platform.Shell{
			CmdFlags: &platform.Flags{
				Config: config,
			},
		}
		env.Init()
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

func cleanOutputPath(path string, env platform.Environment) string {
	if strings.HasPrefix(path, "~") {
		path = strings.TrimPrefix(path, "~")
		path = filepath.Join(env.Home(), path)
	}
	if !filepath.IsAbs(path) {
		if absPath, err := filepath.Abs(path); err == nil {
			path = absPath
		}
	}
	return filepath.Clean(path)
}

func init() { //nolint:gochecknoinits
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "config format to migrate to")
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "config file to export to")
	configCmd.AddCommand(exportCmd)
}
