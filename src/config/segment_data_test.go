package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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

	// The reverse direction: every entry in the file should correspond to a real, currently
	// registered segment type, so a rename or removal is caught here instead of leaving a dead
	// entry that silently stops doing anything.
	var orphaned []string

	for segmentTypeKey := range data.Segments {
		if _, ok := Segments[SegmentType(segmentTypeKey)]; !ok {
			orphaned = append(orphaned, segmentTypeKey)
		}
	}

	assert.Empty(t, orphaned, "website/segment_data.json has entry/entries for unknown segment type(s) "+
		"(renamed or removed segment?): %v", orphaned)
}

// assertEntryKeysRoundTrip unmarshals entry into a fresh writer built by factory, marshals that
// writer back out, and asserts every key present anywhere in entry (at any nesting depth) is
// still present at the same path in the round-tripped JSON. A key that disappears was silently
// ignored by json.Unmarshal, meaning nothing in the real writer struct claims it.
func assertEntryKeysRoundTrip(t *testing.T, segmentType string, factory func() SegmentWriter, entry json.RawMessage) {
	t.Helper()

	writer := factory()

	require.NoError(t, json.Unmarshal(entry, &writer),
		"website/segment_data.json's %q entry does not unmarshal into segments.%T", segmentType, writer)

	roundTripped, err := json.Marshal(writer)
	require.NoError(t, err, "segments.%T failed to marshal back to JSON", writer)

	var original, back any

	require.NoError(t, json.Unmarshal(entry, &original))
	require.NoError(t, json.Unmarshal(roundTripped, &back))

	assertKeysSurvived(t, segmentType, original, back)
}

// assertKeysSurvived recursively compares original (decoded from segment_data.json) against
// back (decoded from the writer's own re-marshaled JSON), reporting every key or array element
// present in original but missing from back.
func assertKeysSurvived(t *testing.T, path string, original, back any) {
	t.Helper()

	switch o := original.(type) {
	case map[string]any:
		b, ok := back.(map[string]any)
		if !ok {
			t.Errorf("%s: not an object after round-tripping through the real struct (got %T) - "+
				"every key under here was silently dropped", path, back)
			return
		}

		for key, oVal := range o {
			childPath := fmt.Sprintf("%s.%s", path, key)

			bVal, present := b[key]
			if !assert.True(t, present, "%s: key %q is not a field the real struct recognizes "+
				"(typo, or the doc's template name differs from the struct's json tag - see this "+
				"segment's src/segments/*.go) - it silently renders as if absent", path, key) {
				continue
			}

			assertKeysSurvived(t, childPath, oVal, bVal)
		}
	case []any:
		b, ok := back.([]any)
		if !ok {
			t.Errorf("%s: not an array after round-tripping through the real struct (got %T)", path, back)
			return
		}

		for i, oVal := range o {
			if i >= len(b) {
				t.Errorf("%s[%d]: element dropped after round-tripping through the real struct", path, i)
				continue
			}

			assertKeysSurvived(t, fmt.Sprintf("%s[%d]", path, i), oVal, b[i])
		}
	default:
		// Scalar (string/number/bool) or null: nothing further to check.
	}
}
