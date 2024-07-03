package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestGetPalette(t *testing.T) {
	palette := color.Palette{
		"red":  "#ff0000",
		"blue": "#0000ff",
	}
	cases := []struct {
		Case            string
		Palettes        *color.Palettes
		Palette         color.Palette
		ExpectedPalette color.Palette
	}{
		{
			Case: "match",
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
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
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
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
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
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
		env := &mock.Environment{}
		env.On("TemplateCache").Return(&cache.Template{
			Env:   map[string]string{},
			Shell: "bash",
		})
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		cfg := &Config{
			env:      env,
			Palette:  tc.Palette,
			Palettes: tc.Palettes,
		}
		got := cfg.getPalette()
		assert.Equal(t, tc.ExpectedPalette, got, tc.Case)
	}
}
