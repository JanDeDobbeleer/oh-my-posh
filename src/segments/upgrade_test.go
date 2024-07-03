package segments

import (
	"errors"
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
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		HasCache        bool
		CurrentVersion  string
		LatestVersion   string
		CachedVersion   string
		Error           error
	}{
		{
			Case:            "Should upgrade",
			CurrentVersion:  "1.0.0",
			LatestVersion:   "1.0.1",
			ExpectedEnabled: true,
		},
		{
			Case:           "On latest",
			CurrentVersion: "1.0.1",
			LatestVersion:  "1.0.1",
		},
		{
			Case:  "Error on update check",
			Error: errors.New("error"),
		},
		{
			Case:            "On previous, from cache",
			HasCache:        true,
			CurrentVersion:  "1.0.2",
			LatestVersion:   "1.0.3",
			CachedVersion:   "1.0.2",
			ExpectedEnabled: true,
		},
		{
			Case:           "On latest, version changed",
			HasCache:       true,
			CurrentVersion: "1.0.2",
			LatestVersion:  "1.0.2",
			CachedVersion:  "1.0.1",
		},
		{
			Case:            "On previous, version changed",
			HasCache:        true,
			CurrentVersion:  "1.0.2",
			LatestVersion:   "1.0.3",
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

		json := fmt.Sprintf(`{"tag_name":"v%s"}`, tc.LatestVersion)
		env.On("HTTPRequest", upgrade.RELEASEURL).Return([]byte(json), tc.Error)

		ug := &Upgrade{
			env:   env,
			props: properties.Map{},
		}

		enabled := ug.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
	}
}
