package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockEnabled(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		Segments []*Segment
		Type     BlockType
	}{
		{Case: "line break block", Expected: true, Type: LineBreak},
		{Case: "prompt enabled", Expected: true, Type: Prompt, Segments: []*Segment{{Enabled: true}}},
		{Case: "prompt disabled", Expected: false, Type: Prompt, Segments: []*Segment{{Enabled: false}}},
		{Case: "prompt enabled multiple", Expected: true, Type: Prompt, Segments: []*Segment{{Enabled: false}, {Enabled: true}}},
		{Case: "rprompt enabled multiple", Expected: true, Type: RPrompt, Segments: []*Segment{{Enabled: false}, {Enabled: true}}},
	}
	for _, tc := range cases {
		block := &Block{
			Type:     tc.Type,
			Segments: tc.Segments,
		}
		assert.Equal(t, tc.Expected, block.Enabled(), tc.Case)
	}
}
