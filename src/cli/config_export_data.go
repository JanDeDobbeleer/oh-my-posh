package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

var outputData string

// dataCmd represents the "config export data" command
var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Export a template data file for your config",
	Long: `Export a template data file for your config.

Runs your config's segments against the real environment and records the
resulting template context and segment data to a file. Feed the recorded
file back in with --data on print/image to render deterministically,
without querying the real environment.

Example usage:

> oh-my-posh config export data --config ~/myconfig.omp.json --output ~/myconfig.data.json

Exports the recorded data to ~/myconfig.data.json.

> oh-my-posh config export data --config ~/myconfig.omp.json

Prints the recorded data to stdout.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		cache.Init(os.Getenv("POSH_SHELL"))

		err := setConfigFlag()
		if err != nil {
			exitcode = 666
			fmt.Println(err.Error())
			return
		}

		cfg := config.Load(configFlag)

		flags := &runtime.Flags{
			ConfigPath:    cfg.Source,
			Shell:         shell.GENERIC,
			TerminalWidth: 120,
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		template.Init(env, cfg.Var, cfg.Maps)

		defer func() {
			template.SaveCache()
			cache.Close()
		}()

		// set sane defaults for things we don't need while recording
		cfg.ConsoleTitleTemplate = ""
		cfg.PWD = ""
		cfg.ShellIntegration = false

		terminal.Init(shell.GENERIC)
		terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
		terminal.Colors = cfg.MakeColors(env)

		eng := &prompt.Engine{
			Config: cfg,
			Env:    env,
		}

		// Executing the primary prompt runs every segment against the real
		// environment and populates both the template cache and each
		// segment's writer, which is what we record below.
		eng.Primary()

		doc, err := buildDataDocument(cfg)
		if err != nil {
			exitcode = 666
			fmt.Println(err.Error())
			return
		}

		if outputData == "" {
			fmt.Println(string(doc))
			return
		}

		if err := os.WriteFile(cleanOutputPath(outputData), doc, 0o644); err != nil {
			exitcode = 666
			fmt.Println(err.Error())
		}
	},
}

// buildDataDocument builds the recorder's output document from an
// already-rendered config: the template cache's simple fields (env) plus
// every enabled segment's writer, keyed by DataKey (segments). Extracted
// from dataCmd's Run so it can be unit tested without a real environment.
func buildDataDocument(cfg *config.Config) ([]byte, error) {
	envRaw, err := json.Marshal(template.Cache.SimpleTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template cache: %w", err)
	}

	var envFields map[string]json.RawMessage
	if err := json.Unmarshal(envRaw, &envFields); err != nil {
		return nil, fmt.Errorf("failed to marshal template cache: %w", err)
	}

	// SegmentsCache is internal cache plumbing, and Var is already covered
	// by the config's own "var" section - neither belongs in a recorded
	// data file.
	delete(envFields, "SegmentsCache")
	delete(envFields, "Var")

	segments := make(map[string]json.RawMessage)

	for _, block := range cfg.Blocks {
		for _, segment := range block.Segments {
			if !segment.Enabled {
				continue
			}

			writer := segment.Writer()
			if writer == nil {
				continue
			}

			raw, err := json.Marshal(writer)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal segment %s: %w", segment.DataKey(), err)
			}

			key := segment.DataKey()
			if _, exists := segments[key]; exists {
				fmt.Fprintf(os.Stderr, "warning: multiple segments share the data key %q; the last one wins - add an alias to disambiguate\n", key)
			}

			segments[key] = raw
		}
	}

	doc := map[string]any{
		"env":      envFields,
		"segments": segments,
	}

	return json.MarshalIndent(doc, "", "  ")
}

func init() {
	dataCmd.Flags().StringVarP(&outputData, "output", "o", "", "data file to export to")

	exportCmd.AddCommand(dataCmd)
}
