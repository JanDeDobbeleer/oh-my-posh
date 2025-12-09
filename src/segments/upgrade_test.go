package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/alecthomas/assert"
)

func TestUpgrade(t *testing.T) {
	ugc := &upgrade.Config{}
	latest, _ := ugc.FetchLatest()

	cases := []struct {
		Case            string
		CurrentVersion  string
		LatestVersion   string
		CachedVersion   string
		ExpectedEnabled bool
		HasCache        bool
	}{
		{
			Case:            "Should upgrade",
			CurrentVersion:  "1.0.0",
			LatestVersion:   "1.0.1",
			ExpectedEnabled: true,
		},
		{
			Case:           "On latest",
			CurrentVersion: latest,
		},
		{
			Case:            "On previous, from cache",
			HasCache:        true,
			CurrentVersion:  "1.0.2",
			LatestVersion:   latest,
			CachedVersion:   "1.0.2",
			ExpectedEnabled: true,
		},
		{
			Case:           "On latest, version changed",
			HasCache:       true,
			CurrentVersion: latest,
			LatestVersion:  latest,
			CachedVersion:  "1.0.1",
		},
		{
			Case:            "On previous, version changed",
			HasCache:        true,
			CurrentVersion:  "1.0.2",
			LatestVersion:   latest,
			CachedVersion:   "1.0.1",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		if tc.CachedVersion == "" {
			tc.CachedVersion = tc.CurrentVersion
		}

		if tc.HasCache {
			data := &UpgradeCache{
				Latest:  tc.LatestVersion,
				Current: tc.CachedVersion,
			}
			cache.Set(cache.Device, UPGRADECACHEKEY, data, cache.INFINITE)
		}

		build.Version = tc.CurrentVersion

		ug := &Upgrade{}
		ug.Init(options.Map{}, env)

		enabled := ug.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		cache.DeleteAll(cache.Device)
	}
}
