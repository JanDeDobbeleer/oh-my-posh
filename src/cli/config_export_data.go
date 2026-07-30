package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/render"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var (
	outputData string
	sanitize   bool
	themesDir  string
)

var dataCmd = &cmdtree.Command{
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

Prints the recorded data to stdout.

> oh-my-posh config export data --sanitize --themes ../themes --output prompt/testdata/fixtures/prompt.data.json

Records every theme found in ../themes against the real environment, merges
them into a single sanitized fixture (the most common recorded value per
segment key wins), and writes it to the given output path. --config is
ignored in this mode. This is the single command that regenerates
src/prompt/testdata/fixtures - run from src/.`,
	Args: cmdtree.NoArgs,
	Run: func(cmd *cmdtree.Command, _ []string) {
		cache.Init(os.Getenv("POSH_SHELL"))

		if themesDir != "" {
			if !sanitize {
				exitcode = 666
				fmt.Println("--themes requires --sanitize: a merged multi-theme fixture is only meant to be committed sanitized")
				return
			}

			defer func() {
				template.SaveCache()
				cache.Close()
			}()

			doc, err := buildMergedDataDocument(themesDir)
			if err != nil {
				exitcode = 666
				fmt.Println(err.Error())
				return
			}

			writeDataOutput(doc)

			return
		}

		setConfigFlag()

		cfg := config.Load(configFlag)

		// render.Config's own eng.Primary() call runs every segment against the
		// real environment and populates both the template cache and each segment's
		// writer, which is what we record below.
		if _, err := render.Config(cfg, 120, false, func(flags *runtime.Flags) error {
			return applyDataFile(flags, cmd.Flags().Changed)
		}); err != nil {
			exitcode = 666
			fmt.Println(err.Error())
			return
		}

		defer func() {
			template.SaveCache()
			cache.Close()
		}()

		doc, err := buildDataDocument(cfg)
		if err != nil {
			exitcode = 666
			fmt.Println(err.Error())
			return
		}

		if sanitize {
			doc, err = sanitizeDataDocument(doc, cfg)
			if err != nil {
				exitcode = 666
				fmt.Println(err.Error())
				return
			}
		}

		writeDataOutput(doc)
	},
}

// writeDataOutput prints doc to stdout, or writes it to --output when set.
// Shared by the single-config path and the --themes merge path.
func writeDataOutput(doc []byte) {
	if outputData == "" {
		fmt.Println(string(doc))
		return
	}

	if err := os.WriteFile(cleanOutputPath(outputData), doc, 0o644); err != nil {
		exitcode = 666
		fmt.Println(err.Error())
	}
}

// Extracted from dataCmd's Run so it can be unit tested without a real environment.
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
			writer := segment.Writer()
			if writer == nil {
				continue
			}

			raw, methods, err := recordSegmentData(writer)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal segment %s: %w", segment.DataKey(), err)
			}

			recorded := config.RecordedSegment{Data: raw, Methods: methods, Enabled: segment.Enabled}

			recordedRaw, err := json.Marshal(recorded)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal segment %s: %w", segment.DataKey(), err)
			}

			key := segment.DataKey()
			if _, exists := segments[key]; exists {
				fmt.Fprintf(os.Stderr, "warning: multiple segments share the data key %q; the last one wins - add an alias to disambiguate\n", key)
			}

			segments[key] = recordedRaw
		}
	}

	doc := map[string]any{
		config.DataVersionKey:  config.DataVersion,
		config.DataEnvKey:      envFields,
		config.DataSegmentsKey: segments,
	}

	return json.MarshalIndent(doc, "", "  ")
}

func init() {
	dataCmd.Flags().StringVarP(&outputData, "output", "o", "", "data file to export to")
	dataCmd.Flags().BoolVar(&sanitize, "sanitize", false,
		"scrub identity (username, hostname, paths, git/cloud identity, sysinfo, battery) from the recorded data, for sharing or committing fixtures")
	// The same flag the image and print commands carry, for the same reason plus one: a fixture
	// recorded once can be re-recorded through the writers to pick up whatever the data format
	// has since gained - method results, say - without giving up the values it was curated with.
	dataCmd.Flags().StringVar(&dataPath, "data", "",
		"path to a template data file to seed the recording with, instead of the live environment")
	dataCmd.Flags().StringVar(&themesDir, "themes", "",
		"record every theme in this directory and merge them into one sanitized fixture, ignoring --config (requires --sanitize)")

	exportCmd.AddCommand(dataCmd)
}

// recordThemeSanitized runs one theme's segments against the real environment,
// exactly as the single-config path above does, then returns its sanitized env
// and segment maps for merging. Each call gets its own fresh template cache
// (resetTemplateCache=true, below, drives render.Config's own
// template.ResetCache before template.Init) so one theme's Var/Maps never leak
// into the next theme's render, the same isolation prompt/golden_test.go's
// renderTheme relies on between themes.
func recordThemeSanitized(themePath string) (env, segments map[string]json.RawMessage, err error) {
	cfg := config.Load(themePath)
	if cfg.Source == "" {
		return nil, nil, fmt.Errorf("failed to parse theme %s", themePath)
	}

	// --data seeds the writers with a fixture's own values before they render, so re-recording an
	// existing file keeps what it was curated with and only adds what the format has since gained.
	// Without it every theme would record whatever this machine happens to look like.
	if _, err := render.Config(cfg, 120, true, func(flags *runtime.Flags) error {
		return applyDataFile(flags, func(string) bool { return false })
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to record theme %s: %w", themePath, err)
	}

	doc, err := buildDataDocument(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to record theme %s: %w", themePath, err)
	}

	sanitized, err := sanitizeDataDocument(doc, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sanitize theme %s: %w", themePath, err)
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(sanitized, &root); err != nil {
		return nil, nil, fmt.Errorf("failed to parse recorded theme %s: %w", themePath, err)
	}

	if raw, ok := root[config.DataEnvKey]; ok {
		if err := json.Unmarshal(raw, &env); err != nil {
			return nil, nil, fmt.Errorf("failed to parse env for theme %s: %w", themePath, err)
		}
	}

	if raw, ok := root[config.DataSegmentsKey]; ok {
		if err := json.Unmarshal(raw, &segments); err != nil {
			return nil, nil, fmt.Errorf("failed to parse segments for theme %s: %w", themePath, err)
		}
	}

	return env, segments, nil
}

// countPopulatedLeaves parses raw and counts its non-zero-value leaves: a
// non-empty string, a non-zero number, true, or a non-empty array/map each
// count as 1; a leaf holding the zero value for its type (including null)
// counts as 0. Containers themselves are not counted, only what they hold -
// an empty array contributes 0, a 3-element array contributes the sum of
// its elements' own counts.
func countPopulatedLeaves(raw json.RawMessage) int {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0
	}

	return countPopulated(v)
}

func countPopulated(v any) int {
	switch t := v.(type) {
	case bool:
		if t {
			return 1
		}

		return 0
	case float64:
		if t != 0 {
			return 1
		}

		return 0
	case string:
		if t != "" {
			return 1
		}

		return 0
	case []any:
		n := 0
		for _, e := range t {
			n += countPopulated(e)
		}

		return n
	case map[string]any:
		n := 0

		for k, e := range t {
			// Every segment writer embeds segments.Base, which contributes an
			// exported "Segment" field (Index, Text) purely for the engine's
			// own bookkeeping. segment.Render (config/segment.go) calls
			// writer.SetIndex and writer.SetText on every render, unconditionally
			// overwriting whatever a fixture recorded there before the template
			// ever runs - so counting it here would let a theme's incidental
			// block position (a bigger recorded Index, from sorting earlier in
			// this run) outweigh an actual difference in real segment data when
			// picking the richest variant.
			if k == "Segment" {
				continue
			}

			n += countPopulated(e)
		}

		return n
	default:
		// nil (JSON null) and any other unrecognized shape carry no data.
		return 0
	}
}

// mergeRichest picks, among values (which must be supplied in a stable,
// meaningful order - here: alphabetical by theme), the one whose decoded JSON
// has the most populated leaves (countPopulatedLeaves). A test fixture is not
// trying to model "what a typical machine looks like" the way mergeMostCommon
// did - it exists to exercise template logic, and a segment that never
// fetches optional data (git without fetch_status, sysinfo without extra
// fields, ...) simply never references the fields it leaves empty, so richer
// data is harmless to every theme and gives more of them something real to
// render. Among equally rich values, prefer whichever is most common (most
// representative of a real recording); a remaining tie goes to whichever
// distinct value was seen first, so the result is fully deterministic and
// reproducible across runs and machines.
func mergeRichest(values []json.RawMessage) json.RawMessage {
	counts := make(map[string]int, len(values))
	firstSeen := make(map[string]int, len(values))
	richness := make(map[string]int, len(values))

	for i, v := range values {
		s := string(v)
		if _, ok := firstSeen[s]; !ok {
			firstSeen[s] = i
			richness[s] = countPopulatedLeaves(v)
		}

		counts[s]++
	}

	var best string

	bestRichness, bestCount, bestOrder := -1, -1, len(values)+1

	for s, c := range counts {
		r := richness[s]
		o := firstSeen[s]

		switch {
		case r > bestRichness,
			r == bestRichness && c > bestCount,
			r == bestRichness && c == bestCount && o < bestOrder:
			best, bestRichness, bestCount, bestOrder = s, r, c, o
		}
	}

	return json.RawMessage(best)
}

// buildMergedDataDocument records every theme in themesDir against the real
// environment, sanitizes each recording independently (so every candidate
// value going into the merge is already scrubbed of identity), then merges
// them into one fixture: for every env and segment key that appears in any
// theme, the richest recorded value wins (mergeRichest) - not the most common
// one. A fixture exists to exercise template logic, not to model a typical
// machine: a theme that never enables an optional field (git fetch_status,
// extra sysinfo fields, ...) never references it either way, so picking the
// most populated variant costs those themes nothing and gives every other
// theme's template more real data - ahead/behind counts, working-tree status,
// upstream icons - to render instead of zero values. This is what lets a
// single ~23KB fixture stand in for the 124 per-theme recordings it replaces -
// every segment key any bundled theme uses gets a plausible, richly populated
// value, without carrying 124 near-duplicate copies of the sparsest ones.
func buildMergedDataDocument(themesDir string) ([]byte, error) {
	themePaths, err := config.ThemeFiles(themesDir)
	if err != nil {
		return nil, err
	}

	if len(themePaths) == 0 {
		return nil, fmt.Errorf("no theme files found in %s", themesDir)
	}

	envValues := make(map[string][]json.RawMessage)
	segmentValues := make(map[string][]json.RawMessage)

	for _, themePath := range themePaths {
		env, segments, err := recordThemeSanitized(themePath)
		if err != nil {
			return nil, err
		}

		for key, value := range env {
			envValues[key] = append(envValues[key], value)
		}

		for key, value := range segments {
			segmentValues[key] = append(segmentValues[key], value)
		}
	}

	mergedEnv := make(map[string]json.RawMessage, len(envValues))
	for key, values := range envValues {
		mergedEnv[key] = mergeRichest(values)
	}

	mergedSegments := make(map[string]json.RawMessage, len(segmentValues))
	for key, values := range segmentValues {
		mergedSegments[key] = mergeRichest(values)
	}

	doc := map[string]any{
		config.DataVersionKey:  config.DataVersion,
		config.DataEnvKey:      mergedEnv,
		config.DataSegmentsKey: mergedSegments,
	}

	return json.MarshalIndent(doc, "", "  ")
}
