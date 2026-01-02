package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"

	"github.com/spf13/cobra"
)

var (
	format string
	output string
)

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
		if output == "" && format == "" {
			// usage error
			fmt.Println("neither output path nor export format is specified")
			exitcode = 2
			return
		}

		cache.Init(os.Getenv("POSH_SHELL"))

		err := setConfigFlag()
		if err != nil {
			exitcode = 666
			fmt.Println(err.Error())
			return
		}

		cfg := config.Load(configFlag)

		validateExportFormat := func() error {
			format = strings.ToLower(format)
			switch format {
			case config.JSON, config.JSONC:
				format = config.JSON
			case config.TOML, config.TML:
				format = config.TOML
			case config.YAML, config.YML:
				format = config.YAML
			default:
				formats := []string{config.JSON, config.JSONC, config.TOML, config.TML, config.YAML, config.YML}
				// usage error
				fmt.Printf("export format must be one of these: %s\n", strings.Join(formats, ", "))
				exitcode = 2
				return errors.New("invalid export format")
			}

			return nil
		}

		if len(format) != 0 {
			if err := validateExportFormat(); err != nil {
				return
			}
		}

		if output == "" {
			fmt.Print(cfg.Export(format))
			return
		}

		cfg.Source = cleanOutputPath(output)

		if format == "" {
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
