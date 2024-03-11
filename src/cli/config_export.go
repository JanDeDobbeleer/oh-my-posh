package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

		if len(output) == 0 && len(format) == 0 {
			// usage error
			os.Exit(2)
		}

		formats := []string{"json", "jsonc", "toml", "tml", "yaml", "yml"}
		if len(format) != 0 && !slices.Contains(formats, format) {
			// usage error
			os.Exit(2)
		}

		if len(output) == 0 {
			fmt.Print(cfg.Export(format))
			return
		}

		cfg.Output = cleanOutputPath(output, env)

		if len(format) == 0 {
			format = strings.TrimPrefix(filepath.Ext(output), ".")
		}

		switch format {
		case "json", "jsonc":
			format = engine.JSON
		case "toml", "tml":
			format = engine.TOML
		case "yaml", "yml":
			format = engine.YAML
		default:
			// data error
			os.Exit(65)
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
