package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

var output string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your config",
	Long: `Export your config.

You can choose to print the output to stdout, or export your config in the format of your choice.

Example usage:

> oh-my-posh config export --config ~/myconfig.omp.json --format toml

Exports the config file "~/myconfig.omp.json" in TOML format and prints the result to stdout.

> oh-my-posh config export --output ~/new_config.omp.json

Exports the current config to "~/new_config.omp.json" (in JSON format).`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if len(output) == 0 && len(format) == 0 {
			// usage error
			fmt.Println("neither output path nor export format is specified")
			exitcode = 2
			return
		}

		configFile := config.Path(configFlag)
		cfg := config.Load(configFile, shell.GENERIC, false)

		validateExportFormat := func() error {
			format = strings.ToLower(format)
			switch format {
			case "json", "jsonc":
				format = config.JSON
			case "toml", "tml":
				format = config.TOML
			case "yaml", "yml":
				format = config.YAML
			default:
				formats := []string{"json", "jsonc", "toml", "tml", "yaml", "yml"}
				// usage error
				fmt.Printf("export format must be one of these: %s\n", strings.Join(formats, ", "))
				exitcode = 2
				return errors.New("invalide export format")
			}

			return nil
		}

		if len(format) != 0 {
			if err := validateExportFormat(); err != nil {
				return
			}
		}

		if len(output) == 0 {
			fmt.Print(cfg.Export(format))
			return
		}

		cfg.Output = cleanOutputPath(output)

		if len(format) == 0 {
			format = strings.TrimPrefix(filepath.Ext(output), ".")
			if err := validateExportFormat(); err != nil {
				return
			}
		}

		cfg.Write(format)
	},
}

func cleanOutputPath(output string) string {
	output = path.ReplaceTildePrefixWithHomeDir(output)

	if !filepath.IsAbs(output) {
		if absPath, err := filepath.Abs(output); err == nil {
			output = absPath
		}
	}

	return filepath.Clean(output)
}

func init() {
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "config format to migrate to")
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "config file to export to")
	configCmd.AddCommand(exportCmd)
}
