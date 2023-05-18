package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"

	"github.com/alecthomas/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestUpgrade(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		HasCache        bool
		CurrentVersion  string
		LatestVersion   string
		CacheVersion    string
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
			Case:  "Version error",
			Error: errors.New("error"),
		},
		{
			Case:            "On previous, from cache",
			HasCache:        true,
			CurrentVersion:  "1.0.2",
			LatestVersion:   "1.0.3",
			CacheVersion:    "1.0.2",
			ExpectedEnabled: true,
		},
		{
			Case:           "On latest, version changed",
			HasCache:       true,
			CurrentVersion: "1.0.2",
			LatestVersion:  "1.0.2",
			CacheVersion:   "1.0.1",
		},
		{
			Case:            "On previous, version changed",
			HasCache:        true,
			CurrentVersion:  "1.0.2",
			LatestVersion:   "1.0.3",
			CacheVersion:    "1.0.1",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		cache := &mock.MockedCache{}

		env.On("Cache").Return(cache)
		if len(tc.CacheVersion) == 0 {
			tc.CacheVersion = tc.CurrentVersion
		}
		cacheData := fmt.Sprintf(`{"latest":"v%s", "current": "v%s"}`, tc.LatestVersion, tc.CacheVersion)
		cache.On("Get", UPGRADECACHEKEY).Return(cacheData, tc.HasCache)
		cache.On("Set", mock2.Anything, mock2.Anything, mock2.Anything)

		env.On("Flags").Return(&platform.Flags{Version: tc.CurrentVersion})

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
