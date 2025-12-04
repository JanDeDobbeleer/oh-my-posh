package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
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
			SimpleTemplate: cache.SimpleTemplate{
				Shell: "bash",
			},
		}
		template.Init(env, nil, nil)

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
			ExpectedFeats:         0,
		},
		{
			Case:          "auto upgrade enabled",
			AutoUpgrade:   true,
			ExpectedFeats: shell.Upgrade,
		},
		{
			Case:           "auto upgrade via cache",
			AutoUpgradeKey: true,
			ExpectedFeats:  shell.Upgrade,
		},
		{
			Case:          "notice enabled, no auto upgrade",
			DisplayNotice: true,
			ExpectedFeats: shell.Notice,
		},
		{
			Case:          "notice via cache, no auto upgrade",
			NoticeKey:     true,
			ExpectedFeats: shell.Notice,
		},
		{
			Case:                  "force upgrade ignores cache",
			UpgradeCacheKeyExists: true,
			Force:                 true,
			AutoUpgrade:           true,
			ExpectedFeats:         shell.Upgrade,
		},
	}

	for _, tc := range cases {
		if tc.UpgradeCacheKeyExists {
			cache.Set(cache.Device, upgrade.CACHEKEY, "", cache.INFINITE)
		}

		if tc.AutoUpgradeKey {
			cache.Set(cache.Device, AUTOUPGRADE, true, cache.INFINITE)
		}

		if tc.NoticeKey {
			cache.Set(cache.Device, UPGRADENOTICE, true, cache.INFINITE)
		}

		cfg := &Config{
			Upgrade: &upgrade.Config{
				Auto:          tc.AutoUpgrade,
				Force:         tc.Force,
				DisplayNotice: tc.DisplayNotice,
			},
		}

		got := cfg.upgradeFeatures()
		assert.Equal(t, tc.ExpectedFeats, got, tc.Case)

		cache.DeleteAll(cache.Device)
	}
}

func TestRefreshIntervalFeature(t *testing.T) {
	cases := []struct {
		Case             string
		Shell            string
		RefreshInterval  int
		ExpectedEnabled  bool
	}{
		{
			Case:            "pwsh with refresh interval",
			Shell:           shell.PWSH,
			RefreshInterval: 5000,
			ExpectedEnabled: true,
		},
		{
			Case:            "zsh with refresh interval",
			Shell:           shell.ZSH,
			RefreshInterval: 3000,
			ExpectedEnabled: true,
		},
		{
			Case:            "bash with refresh interval",
			Shell:           shell.BASH,
			RefreshInterval: 3000,
			ExpectedEnabled: true,
		},
		{
			Case:            "fish with refresh interval",
			Shell:           shell.FISH,
			RefreshInterval: 3000,
			ExpectedEnabled: true,
		},
		{
			Case:            "pwsh without refresh interval",
			Shell:           shell.PWSH,
			RefreshInterval: 0,
			ExpectedEnabled: false,
		},
		{
			Case:            "elvish with refresh interval (unsupported)",
			Shell:           shell.ELVISH,
			RefreshInterval: 5000,
			ExpectedEnabled: false,
		},
		{
			Case:            "cmd with refresh interval (unsupported)",
			Shell:           shell.CMD,
			RefreshInterval: 5000,
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		env.On("Shell").Return(tc.Shell)
		env.On("Getenv", "OMP_CACHE_DISABLED").Return("")

		cfg := &Config{
			RefreshInterval: tc.RefreshInterval,
			Blocks:          []*Block{},
			Upgrade: &upgrade.Config{
				Auto:          false,
				DisplayNotice: false,
			},
		}

		feats := cfg.Features(env)
		hasRefreshInterval := feats&shell.RefreshInterval != 0

		assert.Equal(t, tc.ExpectedEnabled, hasRefreshInterval, tc.Case)
	}
}
