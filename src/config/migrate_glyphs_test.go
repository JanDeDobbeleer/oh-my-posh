package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCodePoints(t *testing.T) {
	codepoints, err := getGlyphCodePoints()
	if connectionError, ok := err.(*ConnectionError); ok {
		t.Log(connectionError.Error())
		return
	}
	assert.Equal(t, 1939, len(codepoints))
}

func TestEscapeGlyphs(t *testing.T) {
	cases := []struct {
		Input    string
		Expected string
	}{
		{Input: "󰉋", Expected: "\\udb80\\ude4b"},
		{Input: "a", Expected: "a"},
		{Input: "\ue0b4", Expected: "\\ue0b4"},
		{Input: "\ufd03", Expected: "\\ufd03"},
		{Input: "}", Expected: "}"},
		{Input: "🏚", Expected: "🏚"},
		{Input: "\U000f0bc9", Expected: "\\udb82\\udfc9"},
		{Input: "󰯉", Expected: "\\udb82\\udfc9"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.Expected, EscapeGlyphs(tc.Input, false), tc.Input)
	}
}
