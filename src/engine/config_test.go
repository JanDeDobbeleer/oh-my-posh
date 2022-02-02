package engine

import (
	"oh-my-posh/segments"
	"testing"

	"github.com/gookit/config/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func testClearDefaultConfig() {
	config.Default().ClearAll()
}

func TestParseMappedLocations(t *testing.T) {
	defer testClearDefaultConfig()
	cases := []struct {
		Case string
		JSON string
	}{
		{Case: "new format", JSON: `{ "properties": { "mapped_locations": {"folder1": "one","folder2": "two"} } }`},
		{Case: "old format", JSON: `{ "properties": { "mapped_locations": [["folder1", "one"], ["folder2", "two"]] } }`},
	}
	for _, tc := range cases {
		config.ClearAll()
		config.WithOptions(func(opt *config.Options) {
			opt.DecoderConfig = &mapstructure.DecoderConfig{
				TagName: "config",
			}
		})
		err := config.LoadStrings(config.JSON, tc.JSON)
		assert.NoError(t, err)
		var segment Segment
		err = config.BindStruct("", &segment)
		assert.NoError(t, err)
		mappedLocations := segment.Properties.GetKeyValueMap(segments.MappedLocations, make(map[string]string))
		assert.Equal(t, "two", mappedLocations["folder2"])
	}
}

func TestEscapeGlyphs(t *testing.T) {
	defer testClearDefaultConfig()
	cases := []struct {
		Input    string
		Expected string
	}{
		{Input: "a", Expected: "a"},
		{Input: "\ue0b4", Expected: "\\ue0b4"},
		{Input: "\ufd03", Expected: "\\ufd03"},
		{Input: "}", Expected: "}"},
		{Input: "🏚", Expected: "🏚"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.Expected, escapeGlyphs(tc.Input), tc.Input)
	}
}
