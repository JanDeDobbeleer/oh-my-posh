package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// segmentDataPath is the hand-written sample data the website's theme gallery and studio
// render with (see website/export_themes.mjs's --data flag). It is committed, not generated -
// unlike registry.json below there is nothing to regenerate here, only something to keep honest.
var segmentDataPath = filepath.Join("..", "..", "website", "segment_data.json")

// TestSegmentDataCoversRegistry asserts that every documented segment type (the live registry
// derived the same way TestSegmentCatalogMatchesRegistry does) has an entry in
// website/segment_data.json. A type with no entry there renders as absent in the website's
// gallery/studio - see segment_data.json's own comment on the "env" section, and the "55 types
// have none" gap this test exists to close and keep closed.
//
// root and text are not special-cased here: both have a fixed-literal template with no
// segment-specific fields, so they get a trivial {} entry instead of an exemption. That is
// simpler than threading an exclusion list through this test, and it is not a no-op: an empty
// JSON object is still non-empty *bytes* ("{}"), so overlayData (segment.go) still applies it and
// flips the segment enabled, exactly like every other pinned entry.
func TestSegmentDataCoversRegistry(t *testing.T) {
	data, err := LoadData(segmentDataPath)
	require.NoError(t, err, "failed to load %s", segmentDataPath)

	var missing []string

	for segmentType := range Segments {
		if segmentType == EXIT {
			// EXIT is a legacy alias for STATUS (see buildSegmentCatalog's own comment on this);
			// STATUS itself is checked below, so EXIT would only ever be a duplicate here.
			continue
		}

		if _, ok := data.Segments[string(segmentType)]; !ok {
			missing = append(missing, string(segmentType))
		}
	}

	assert.Empty(t, missing, "website/segment_data.json has no entry for segment type(s): %v", missing)
}

// TestSegmentDataKeysAreRecognized is the guard against a mistyped or stale key silently
// dropping out of a segment_data.json entry - a hand-written data file with a key that does not
// match the real writer struct is not a parse error, it is a *silent no-op*: json.Unmarshal
// ignores fields it does not recognize, so the segment renders with that one field simply
// missing, or - if every key in the entry is unrecognized - renders completely empty. That
// failure mode has already happened twice in this project (see the task that added this test).
//
// This deliberately checks against the real segments.<Type> struct via a round trip
// (json.Unmarshal into a fresh writer, json.Marshal it back out, then diff key-for-key) rather
// than against website/docs/segments/*.mdx's "### Properties" tables. The two are not the same
// contract: a documented property name is always the Go template accessor (PascalCase, e.g.
// `.Model`), but the JSON key segment_data.json actually needs is whatever json tag the struct
// field carries, which for a segment whose writer embeds an upstream API payload (claude,
// copilot, copilot_cli, nightscout, orthodoxcal, carbonintensity, upgrade, pulumi, quasar) is a
// different, often snake_case or lowercase, string. Parsing the docs in Go as well would mean
// maintaining two schemas that could drift from each other; unmarshaling into the actual struct
// this repo ships is both simpler and exactly the contract that matters - if a key survives this
// round trip, the real code path (config/segment.go's overlayData) accepts it too, and if it
// does not, overlayData would silently have dropped it exactly the same way.
//
// website/scripts/extract-segment-properties.mjs parses the docs' Properties tables instead;
// that script is for authoring/auditing segment_data.json against what users see documented, not
// for enforcement - the docs are the source of truth for what a property is *called* to a user,
// this test is the source of truth for what segment_data.json must *write*.
func TestSegmentDataKeysAreRecognized(t *testing.T) {
	data, err := LoadData(segmentDataPath)
	require.NoError(t, err, "failed to load %s", segmentDataPath)

	for segmentType, factory := range Segments {
		if segmentType == EXIT {
			continue
		}

		entry, ok := data.Segments[string(segmentType)]
		if !ok {
			// TestSegmentDataCoversRegistry already reports this; do not double-report it here.
			continue
		}

		t.Run(string(segmentType), func(t *testing.T) {
			assertEntryKeysRoundTrip(t, string(segmentType), factory, entry)
		})
	}

	// The reverse direction: every entry should correspond to something that can actually ask for
	// it, so a rename or removal is caught here instead of leaving a dead entry that silently
	// stops doing anything.
	//
	// An entry is keyed by Segment.DataKey, which is the segment's alias where it has one - that
	// is how a config with two segments of the same type gives each its own data. So a key that
	// is not a segment type is not automatically dead: it is dead only if no bundled theme claims
	// it as an alias either. cloud-context and pixelrobots both alias an az segment to "azpwsh".
	aliases := themeAliases(t)

	var orphaned []string

	for segmentTypeKey := range data.Segments {
		if _, ok := Segments[SegmentType(segmentTypeKey)]; ok {
			continue
		}

		if aliases[segmentTypeKey] {
			continue
		}

		orphaned = append(orphaned, segmentTypeKey)
	}

	assert.Empty(t, orphaned, "website/segment_data.json has entry/entries matching neither a registered "+
		"segment type nor an alias any bundled theme uses (renamed or removed?): %v", orphaned)
}

// assertEntryKeysRoundTrip checks that every key in an entry names something the writer actually
// has: a field, or a method whose result the recorder captured.
//
// A direct check rather than a JSON round trip. The recorder keys by Go field name (see
// cli/segment_data.go) while encoding/json marshals by tag, so round-tripping would compare
// "CumulativeTotal" against "cumulative_total" and call every tagged field a typo.
func assertEntryKeysRoundTrip(t *testing.T, segmentType string, factory func() SegmentWriter, entry json.RawMessage) {
	t.Helper()

	// A recorded file wraps each segment as { enabled, data }; the fields being checked are the
	// ones inside. A hand-written file stores them flat, which is what the unwrap falls back to.
	if recorded, isRecorded := decodeRecordedSegment(entry); isRecorded {
		entry = recorded.Data
	}

	writer := factory()

	require.NoError(t, json.Unmarshal(entry, &writer),
		"website/segment_data.json's %q entry does not unmarshal into segments.%T", segmentType, writer)

	var original map[string]any
	require.NoError(t, json.Unmarshal(entry, &original))

	assertNamesExist(t, segmentType, original, reflect.ValueOf(writer))
}

// assertNamesExist walks an entry alongside the writer it describes, reporting any key that names
// neither a field nor a method.
func assertNamesExist(t *testing.T, path string, entry map[string]any, value reflect.Value) {
	t.Helper()

	names, fields := reachableNames(value)

	for key, nested := range entry {
		childPath := fmt.Sprintf("%s.%s", path, key)

		if !assert.True(t, names[key], "%s: key %q names neither a field nor a recorded method on "+
			"the real struct (typo, or a renamed field - see this segment's src/segments/*.go) - "+
			"it silently renders as if absent", path, key) {
			continue
		}

		// Only a nested object can be walked further, and only where the struct has a field to
		// walk into: a method result has no struct behind it to check against.
		nestedEntry, isObject := nested.(map[string]any)
		field, hasField := fields[key]

		if isObject && hasField && reflect.Indirect(field).Kind() == reflect.Struct {
			assertNamesExist(t, childPath, nestedEntry, field)
		}
	}
}

// reachableNames returns every name a recorded entry may legitimately use for this value: its
// exported fields (which the recorder keys by Go name) and its exported methods (whose results
// the recorder captures alongside them), plus the fields themselves for walking into.
func reachableNames(value reflect.Value) (map[string]bool, map[string]reflect.Value) {
	names := make(map[string]bool)
	fields := make(map[string]reflect.Value)

	if !value.IsValid() {
		return names, fields
	}

	for i := range value.NumMethod() {
		names[value.Type().Method(i).Name] = true
	}

	structValue := reflect.Indirect(value)
	if structValue.Kind() != reflect.Struct {
		return names, fields
	}

	var collect func(reflect.Value)

	collect = func(sv reflect.Value) {
		for i := range sv.NumField() {
			field := sv.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			// Embedded fields are flattened by the recorder, so their names belong to the outer
			// struct as far as an entry is concerned.
			if field.Anonymous && reflect.Indirect(sv.Field(i)).Kind() == reflect.Struct {
				collect(reflect.Indirect(sv.Field(i)))
				continue
			}

			names[field.Name] = true
			fields[field.Name] = sv.Field(i)
		}
	}

	collect(structValue)

	return names, fields
}

// themeAliases collects every alias the bundled themes declare, so an entry keyed by one is
// recognised as the deliberate thing it is rather than reported as a dead segment type.
func themeAliases(t *testing.T) map[string]bool {
	t.Helper()

	aliases := make(map[string]bool)

	entries, err := os.ReadDir(filepath.Join("..", "..", "themes"))
	require.NoError(t, err)

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".omp.json") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join("..", "..", "themes", entry.Name()))
		require.NoError(t, err)

		var theme struct {
			Blocks []struct {
				Segments []struct {
					Alias string `json:"alias"`
				} `json:"segments"`
			} `json:"blocks"`
		}

		if err := json.Unmarshal(raw, &theme); err != nil {
			continue
		}

		for _, block := range theme.Blocks {
			for _, segment := range block.Segments {
				if segment.Alias != "" {
					aliases[segment.Alias] = true
				}
			}
		}
	}

	return aliases
}
