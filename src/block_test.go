package main

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
		{Case: "prompt enabled", Expected: true, Type: Prompt, Segments: []*Segment{{active: true}}},
		{Case: "prompt disabled", Expected: false, Type: Prompt, Segments: []*Segment{{active: false}}},
		{Case: "prompt enabled multiple", Expected: true, Type: Prompt, Segments: []*Segment{{active: false}, {active: true}}},
		{Case: "rprompt enabled multiple", Expected: true, Type: RPrompt, Segments: []*Segment{{active: false}, {active: true}}},
	}
	for _, tc := range cases {
		block := &Block{
			Type:     tc.Type,
			Segments: tc.Segments,
		}
		assert.Equal(t, tc.Expected, block.enabled(), tc.Case)
	}
}

// func TestForegroundColor(t *testing.T) {
// 	cases := []struct {
// 		Case            string
// 		Expected        string
// 		PreviousSegment *Segment
// 		Segment         *Segment
// 	}{
// 		{Case: "Standard", Expected: "black", Segment: &Segment{Foreground: "black"}},
// 		{Case: "Segment color override", Expected: "red", Segment: &Segment{Foreground: "black", props: &properties{foreground: "red"}}},
// 		{Case: "Inherit", Expected: "yellow", Segment: &Segment{Foreground: Inherit}, PreviousSegment: &Segment{Foreground: "yellow"}},
// 	}

// 	for _, tc := range cases {
// 		block := &Block{}
// 		block.previousActiveSegment = tc.PreviousSegment
// 		block.activeSegment = tc.Segment
// 		assert.Equal(t, tc.Expected, block.foreground(), tc.Case)
// 	}
// }

// func TestBackgroundColor(t *testing.T) {
// 	cases := []struct {
// 		Case            string
// 		Expected        string
// 		PreviousSegment *Segment
// 		Segment         *Segment
// 	}{
// 		{Case: "Standard", Expected: "black", Segment: &Segment{Background: "black"}},
// 		{Case: "Segment color override", Expected: "red", Segment: &Segment{Background: "black", props: &properties{background: "red"}}},
// 		{Case: "Inherit", Expected: "yellow", Segment: &Segment{Background: Inherit}, PreviousSegment: &Segment{Background: "yellow"}},
// 	}

// 	for _, tc := range cases {
// 		block := &Block{}
// 		block.previousActiveSegment = tc.PreviousSegment
// 		block.activeSegment = tc.Segment
// 		assert.Equal(t, tc.Expected, block.background(), tc.Case)
// 	}
// }
