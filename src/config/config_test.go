package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	cache_ "github.com/jandedobbeleer/oh-my-posh/src/cache/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"

	"github.com/stretchr/testify/assert"
	mock_ "github.com/stretchr/testify/mock"
)

func TestGetPalette(t *testing.T) {
	palette := color.Palette{
		"red":  "#ff0000",
		"blue": "#0000ff",
	}

	cases := []struct {
		Palettes        *color.Palettes
		Palette         color.Palette
		ExpectedPalette color.Palette
		Case            string
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
		{
			Case: "match, with override",
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
					"bash": {
						"red":    "#ff0001",
						"yellow": "#ffff00",
					},
				},
			},
			Palette: palette,
			ExpectedPalette: color.Palette{
				"red":    "#ff0001",
				"blue":   "#0000ff",
				"yellow": "#ffff00",
			},
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		env.On("Shell").Return("bash")

		template.Cache = &cache.Template{
			Shell: "bash",
		}
		template.Init(env, nil)

		cfg := &Config{
			Palette:  tc.Palette,
			Palettes: tc.Palettes,
		}

		got := cfg.getPalette()
		assert.Equal(t, tc.ExpectedPalette, got, tc.Case)
	}
}
func TestUpgradeFeatures(t *testing.T) {
	cases := []struct {
		Case                  string
		ExpectedFeats         shell.Features
		UpgradeCacheKeyExists bool
		AutoUpgrade           bool
		Force                 bool
		DisplayNotice         bool
		AutoUpgradeKey        bool
		NoticeKey             bool
	}{
		{
			Case:                  "cache exists, no force",
			UpgradeCacheKeyExists: true,
			ExpectedFeats:         shell.Features{},
		},
		{
			Case:          "auto upgrade enabled",
			AutoUpgrade:   true,
			ExpectedFeats: shell.Features{shell.Upgrade},
		},
		{
			Case:           "auto upgrade via cache",
			AutoUpgradeKey: true,
			ExpectedFeats:  shell.Features{shell.Upgrade},
		},
		{
			Case:          "notice enabled, no auto upgrade",
			DisplayNotice: true,
			ExpectedFeats: shell.Features{shell.Notice},
		},
		{
			Case:          "notice via cache, no auto upgrade",
			NoticeKey:     true,
			ExpectedFeats: shell.Features{shell.Notice},
		},
		{
			Case:                  "force upgrade ignores cache",
			UpgradeCacheKeyExists: true,
			Force:                 true,
			AutoUpgrade:           true,
			ExpectedFeats:         shell.Features{shell.Upgrade},
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		c := &cache_.Cache{}
		env.On("Cache").Return(c)

		if tc.UpgradeCacheKeyExists {
			c.On("Get", upgrade.CACHEKEY).Return("", true)
		} else {
			c.On("Get", upgrade.CACHEKEY).Return("", false)
		}

		c.On("Set", upgrade.CACHEKEY, "", mock_.Anything).Return()

		if tc.AutoUpgradeKey {
			c.On("Get", AUTOUPGRADE).Return("", true)
		} else {
			c.On("Get", AUTOUPGRADE).Return("", false)
		}

		if tc.NoticeKey {
			c.On("Get", UPGRADENOTICE).Return("", true)
		} else {
			c.On("Get", UPGRADENOTICE).Return("", false)
		}

		cfg := &Config{
			Upgrade: &upgrade.Config{
				Auto:          tc.AutoUpgrade,
				Force:         tc.Force,
				DisplayNotice: tc.DisplayNotice,
			},
		}

		got := cfg.UpgradeFeatures(env)
		assert.Equal(t, tc.ExpectedFeats, got, tc.Case)
	}
}
