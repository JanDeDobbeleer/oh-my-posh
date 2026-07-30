package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/gookit/goutil/jsonutil"
	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

// DataVersion marks a data file produced by `config export data`. Its presence
// signals a hermetic, machine-recorded file: every segment's raw message is a
// RecordedSegment envelope carrying the segment's recorded enabled state
// alongside its writer data, so a recorded-but-disabled segment can be
// suppressed without probing it live. Its absence signals a hand-written file:
// still supported, but not hermetic - see restoreData in segment.go for the
// derive-then-overlay path that takes instead.
const DataVersion = 1

// Data document top-level keys: used both when recording a data file
// (cli/config_export_data.go's buildDataDocument/buildMergedDataDocument) and
// when reading one back (LoadData below), so a typo in either place fails to
// compile instead of silently desynchronizing the recorder and the replayer.
const (
	DataVersionKey  = "version"
	DataEnvKey      = "env"
	DataSegmentsKey = "segments"
)

// ThemeFileExtensions is every extension a bundled theme file is recognized
// by: mirrors website/export_themes.mjs's THEME_EXTENSIONS, the third copy of
// this list living outside this repo's Go code.
var ThemeFileExtensions = []string{".omp.json", ".omp.toml", ".omp.yaml"}

// ThemeFiles enumerates the theme files in dir, sorted so a caller merging or
// comparing across all of them (cli's --themes merge mode, prompt's
// golden-fixture harness) gets a deterministic order - for the merge mode
// specifically, so mergeRichest's first-seen tie-break consistently favors
// the alphabetically first theme.
func ThemeFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read themes directory %s: %w", dir, err)
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		for _, ext := range ThemeFileExtensions {
			if strings.HasSuffix(name, ext) {
				files = append(files, filepath.Join(dir, name))
				break
			}
		}
	}

	sort.Strings(files)

	return files, nil
}

// RecordedSegment is the per-segment envelope a data file uses once it carries a
// version marker. A hand-written file skips this wrapper and stores a segment's
// raw writer fields directly, as it always has.
type RecordedSegment struct {
	Data json.RawMessage `json:"data"`
	// Methods holds the method results that have no room in Data, shaped like it. A recorded
	// value's methods normally sit alongside its fields, but a value that is not a struct has no
	// fields to sit alongside: battery.State is an int, and an entry whose "State" were the object
	// its String() belongs in would no longer unmarshal into the writer. Kept apart, Data stays
	// writer-shaped and a render with no writer merges this over it - see MergeRecordedMethods.
	Methods json.RawMessage `json:"methods,omitempty"`
	Enabled bool            `json:"enabled"`
}

// MergeRecordedMethods overlays a recorded segment's Methods tree on its Data tree. Two maps are
// merged key by key; anything else is replaced outright, which is what puts a scalar's method
// results in the place of the scalar.
func MergeRecordedMethods(data, methods any) any {
	dataMap, isDataMap := data.(map[string]any)

	methodsMap, isMethodsMap := methods.(map[string]any)
	if !isDataMap || !isMethodsMap {
		return methods
	}

	for key, value := range methodsMap {
		if existing, ok := dataMap[key]; ok {
			dataMap[key] = MergeRecordedMethods(existing, value)
			continue
		}

		dataMap[key] = value
	}

	return dataMap
}

// Data holds template data supplied via the --data flag, used to render a
// prompt deterministically without querying the real runtime.
type Data struct {
	Segments map[string]json.RawMessage
	Env      json.RawMessage
	Version  int
}

// EnvData holds the subset of the env section that maps directly onto
// runtime.Flags rather than the template cache. Pointer fields let callers
// detect whether a key was present in the data file.
type EnvData struct {
	PWD           *string
	Code          *int
	ExecutionTime *float64
	PipeStatus    *string
	Interrupted   *bool
	Executed      *bool
}

// The format is derived from the file extension: .json/.jsonc, .yaml/.yml, or .toml.
func LoadData(path string) (*Data, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

	root, err := decodeDataRoot(ext, raw)
	if err != nil {
		return nil, err
	}

	return dataFromRoot(root)
}

// ParseData decodes a data document already in memory as JSON - the one
// format a caller with no file path can name unambiguously (the js/wasm
// entrypoint's dataJSON argument is exactly this: the studio hands over the
// text of a file it already read on its own side, with no extension to
// derive a format from). LoadData's YAML/TOML branches stay file-only; they
// exist for a --data flag pointing at a hand-written file, not for the
// wasm entrypoint's use case, which always receives JSON (the same format
// `config export data` records).
func ParseData(raw []byte) (*Data, error) {
	root, err := decodeDataRoot(JSON, raw)
	if err != nil {
		return nil, err
	}

	return dataFromRoot(root)
}

// decodeDataRoot decodes raw into a generic map of raw JSON messages,
// dispatching on ext exactly as LoadData's own format switch used to do
// inline, before ParseData needed the same switch for its one fixed format.
func decodeDataRoot(ext string, raw []byte) (map[string]json.RawMessage, error) {
	var root map[string]json.RawMessage

	switch ext {
	case JSON, JSONC:
		raw = []byte(jsonutil.StripComments(string(raw)))

		if err := json.Unmarshal(raw, &root); err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}
	case YAML, YML:
		var generic map[string]any

		if err := yaml.Unmarshal(raw, &generic); err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}

		normalized, err := normalize(generic)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}

		root = normalized
	case TOML, TML:
		var generic map[string]any

		if err := toml.Unmarshal(raw, &generic); err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}

		normalized, err := normalize(generic)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}

		root = normalized
	default:
		return nil, fmt.Errorf("unsupported data file extension: %s", ext)
	}

	return root, nil
}

// dataFromRoot builds a Data from a generically-decoded document root,
// shared by LoadData (any of its three formats) and ParseData (JSON only)
// once each has the same map[string]json.RawMessage shape in hand.
func dataFromRoot(root map[string]json.RawMessage) (*Data, error) {
	data := &Data{}

	if versionRaw, OK := root[DataVersionKey]; OK {
		if err := json.Unmarshal(versionRaw, &data.Version); err != nil {
			return nil, fmt.Errorf("failed to parse version in data file: %w", err)
		}
	}

	if env, OK := root[DataEnvKey]; OK {
		data.Env = env
	}

	if segmentsRaw, OK := root[DataSegmentsKey]; OK {
		var segments map[string]json.RawMessage

		if err := json.Unmarshal(segmentsRaw, &segments); err != nil {
			return nil, fmt.Errorf("failed to parse segments in data file: %w", err)
		}

		data.Segments = segments
	}

	return data, nil
}

// normalize re-marshals a generically decoded section (as produced by the
// YAML/TOML unmarshalers) to a map of raw JSON so the rest of the pipeline
// can treat all formats uniformly.
func normalize(generic map[string]any) (map[string]json.RawMessage, error) {
	root := make(map[string]json.RawMessage, len(generic))

	for key, value := range generic {
		b, err := json.Marshal(sanitize(value))
		if err != nil {
			return nil, err
		}

		root[key] = b
	}

	return root, nil
}

// sanitize recursively walks a generically decoded value so nested
// maps/slices are sanitized too, not just the top level normalize hands it.
// It does not need a map[any]any case: both decoders normalize calls with
// (go.yaml.in/yaml/v3 and go-toml/v2) decode a nested map as map[string]any
// when unmarshaling into a map[string]any root, exactly as LoadData does -
// proved by round-tripping a nested-map YAML and TOML document through the
// real decoders and asserting the concrete type (see git history for this
// file around the removal of the map[any]any branch for the probe test that
// established this, if the reasoning ever needs re-verifying).
func sanitize(value any) any {
	switch v := value.(type) {
	case map[string]any:
		m := make(map[string]any, len(v))
		for key, val := range v {
			m[key] = sanitize(val)
		}

		return m
	case []any:
		s := make([]any, len(v))
		for i, val := range v {
			s[i] = sanitize(val)
		}

		return s
	default:
		return v
	}
}

// EnvFlags parses the env section for the properties that route into
// runtime.Flags instead of the template cache. Unknown keys are ignored.
func (d *Data) EnvFlags() (*EnvData, error) {
	env := &EnvData{}

	if len(d.Env) == 0 {
		return env, nil
	}

	if err := json.Unmarshal(d.Env, env); err != nil {
		return nil, fmt.Errorf("failed to parse env data: %w", err)
	}

	return env, nil
}

// ApplyFlags copies d onto flags: SegmentData/EnvData verbatim, plus the env
// keys that map directly onto runtime.Flags fields (see EnvFlags/EnvData),
// with precedence explicit CLI flag > data file > live environment. changed
// reports whether the corresponding CLI flag (by name) was set explicitly, in
// which case the data file's value is skipped; a nil changed means "never
// overridden," for a caller with no CLI flags of its own to check against
// (prompt/golden_test.go's fixture harness).
func (d *Data) ApplyFlags(flags *runtime.Flags, changed func(name string) bool) error {
	if changed == nil {
		changed = func(string) bool { return false }
	}

	flags.SegmentData = d.Segments
	flags.EnvData = d.Env

	envFlags, err := d.EnvFlags()
	if err != nil {
		return err
	}

	if envFlags.PWD != nil && !changed("pwd") {
		flags.PWD = *envFlags.PWD
	}

	if envFlags.Code != nil && !changed("status") {
		flags.ErrorCode = *envFlags.Code
	}

	if envFlags.ExecutionTime != nil && !changed("execution-time") {
		flags.ExecutionTime = *envFlags.ExecutionTime
	}

	if envFlags.PipeStatus != nil && !changed("pipestatus") {
		flags.PipeStatus = *envFlags.PipeStatus
	}

	if envFlags.Interrupted != nil && !changed("interrupted") {
		flags.Interrupted = *envFlags.Interrupted
	}

	if envFlags.Executed != nil && !changed("no-status") {
		flags.NoExitCode = !*envFlags.Executed
	}

	return nil
}
