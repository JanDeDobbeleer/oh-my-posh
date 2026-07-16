package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/jsonutil"
	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

// Data holds template data supplied via the --data flag, used to render a
// prompt deterministically without querying the real runtime.
type Data struct {
	Segments map[string]json.RawMessage
	Env      json.RawMessage
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

// LoadData reads and parses a template data file. The format is derived
// from the file extension: .json/.jsonc, .yaml/.yml, or .toml.
func LoadData(path string) (*Data, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

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

		if root, err = normalize(generic); err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}
	case TOML, TML:
		var generic map[string]any

		if err := toml.Unmarshal(raw, &generic); err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}

		if root, err = normalize(generic); err != nil {
			return nil, fmt.Errorf("failed to parse data file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported data file extension: %s", ext)
	}

	data := &Data{}

	if env, OK := root["env"]; OK {
		data.Env = env
	}

	if segmentsRaw, OK := root["segments"]; OK {
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

// sanitize recursively converts map[any]any (which the YAML/TOML decoders
// may produce for nested maps) into map[string]any so encoding/json can
// marshal it.
func sanitize(value any) any {
	switch v := value.(type) {
	case map[any]any:
		m := make(map[string]any, len(v))
		for key, val := range v {
			m[fmt.Sprintf("%v", key)] = sanitize(val)
		}

		return m
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
