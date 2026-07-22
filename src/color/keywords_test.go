package color

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestResolveParentKeywordCollapsesGradientToLastStop(t *testing.T) {
	gradient := Ansi("linear-gradient(#FF0000, #00FF00, #0000FF)")
	parent := &Set{Background: gradient, Foreground: gradient}

	cases := []struct {
		Case     string
		Keyword  Ansi
		Expected Ansi
	}{
		{Case: "parentBackground", Keyword: ParentBackground, Expected: "#0000FF"},
		{Case: "parentForeground", Keyword: ParentForeground, Expected: "#0000FF"},
	}

	for _, tc := range cases {
		resolved := tc.Keyword.Resolve(nil, []*Set{parent})
		assert.Equal(t, tc.Expected, resolved, tc.Case)
	}
}

func TestResolveCurrentKeywordKeepsGradientIntact(t *testing.T) {
	gradient := Ansi("linear-gradient(#FF0000, #0000FF)")
	current := &Set{Background: gradient, Foreground: gradient}

	cases := []struct {
		Case     string
		Keyword  Ansi
		Expected Ansi
	}{
		{Case: "background", Keyword: Background, Expected: gradient},
		{Case: "foreground", Keyword: Foreground, Expected: gradient},
	}

	for _, tc := range cases {
		resolved := tc.Keyword.Resolve(current, nil)
		assert.Equal(t, tc.Expected, resolved, tc.Case)
	}
}

// TestResolveParentGradientKeywordStop pins the review fix: a parent gradient whose
// last stop is a keyword resolves against the PARENT's colors, never the child's,
// and unresolvable self-references degrade to transparent.
func TestResolveParentGradientKeywordStop(t *testing.T) {
	cases := []struct {
		Case     string
		Color    Ansi
		Expected Ansi
		Parents  []*Set
	}{
		{
			Case:     "foreground last stop resolves against the parent",
			Color:    ParentBackground,
			Parents:  []*Set{{Background: "linear-gradient(#FF0000, foreground)", Foreground: "#00FF00"}},
			Expected: "#00FF00",
		},
		{
			Case:     "foreground last stop collapses a parent foreground gradient",
			Color:    ParentBackground,
			Parents:  []*Set{{Background: "linear-gradient(#FF0000, foreground)", Foreground: "linear-gradient(#111111, #222222)"}},
			Expected: "#222222",
		},
		{
			Case:     "self-referential background stop degrades to transparent",
			Color:    ParentBackground,
			Parents:  []*Set{{Background: "linear-gradient(#FF0000, background)", Foreground: "#00FF00"}},
			Expected: Transparent,
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Color.Resolve(nil, tc.Parents), tc.Case)
	}
}
