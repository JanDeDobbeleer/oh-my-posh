package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	cache_ "github.com/jandedobbeleer/oh-my-posh/src/cache/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"

	"github.com/alecthomas/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestUpgrade(t *testing.T) {
	ugc := &upgrade.Config{}
	latest, _ := ugc.Latest()

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
		cache := &cache_.Cache{}

		env.On("Cache").Return(cache)
		if len(tc.CachedVersion) == 0 {
			tc.CachedVersion = tc.CurrentVersion
		}

		cacheData := fmt.Sprintf(`{"latest":"%s", "current": "%s"}`, tc.LatestVersion, tc.CachedVersion)
		cache.On("Get", UPGRADECACHEKEY).Return(cacheData, tc.HasCache)
		cache.On("Set", testify_.Anything, testify_.Anything, testify_.Anything)

		build.Version = tc.CurrentVersion

		ug := &Upgrade{}
		ug.Init(properties.Map{}, env)

		enabled := ug.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
	}
}
