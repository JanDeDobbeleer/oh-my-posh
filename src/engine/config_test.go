package engine

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

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
		{Input: "\U000F011B", Expected: "\\udb80\\udd1b"},
		{Input: "󰄛", Expected: "\\udb80\\udd1b"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.Expected, escapeGlyphs(tc.Input, false), tc.Input)
	}
}

func TestGetPalette(t *testing.T) {
	palette := ansi.Palette{
		"red":  "#ff0000",
		"blue": "#0000ff",
	}
	cases := []struct {
		Case            string
		Palettes        *ansi.Palettes
		Palette         ansi.Palette
		ExpectedPalette ansi.Palette
	}{
		{
			Case: "match",
			Palettes: &ansi.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]ansi.Palette{
					"bash": palette,
					"zsh": {
						"red":  "#ff0001",
						"blue": "#0000fb",
					},
				},
			},
			ExpectedPalette: palette,
		},
		{
			Case: "no match, no fallback",
			Palettes: &ansi.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]ansi.Palette{
					"fish": palette,
					"zsh": {
						"red":  "#ff0001",
						"blue": "#0000fb",
					},
				},
			},
			ExpectedPalette: nil,
		},
		{
			Case: "no match, default",
			Palettes: &ansi.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]ansi.Palette{
					"zsh": {
						"red":  "#ff0001",
						"blue": "#0000fb",
					},
				},
			},
			Palette:         palette,
			ExpectedPalette: palette,
		},
		{
			Case:            "no palettes",
			ExpectedPalette: nil,
		},
	}
	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env:   map[string]string{},
			Shell: "bash",
		})
		env.On("DebugF", mock2.Anything, mock2.Anything).Return(nil)
		cfg := &Config{
			env:      env,
			Palette:  tc.Palette,
			Palettes: tc.Palettes,
		}
		got := cfg.getPalette()
		assert.Equal(t, tc.ExpectedPalette, got, tc.Case)
	}
}
