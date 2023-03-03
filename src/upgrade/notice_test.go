package upgrade

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/stretchr/testify/assert"

	mock2 "github.com/stretchr/testify/mock"
)

func TestCanUpgrade(t *testing.T) {
	cases := []struct {
		Case           string
		Expected       bool
		CurrentVersion string
		LatestVersion  string
		Error          error
		Cache          bool
		GOOS           string
	}{
		{Case: "Up to date", CurrentVersion: "3.0.0", LatestVersion: "v3.0.0"},
		{Case: "Outdated Windows", Expected: true, CurrentVersion: "3.0.0", LatestVersion: "v3.0.1", GOOS: platform.WINDOWS},
		{Case: "Outdated Linux", Expected: true, CurrentVersion: "3.0.0", LatestVersion: "v3.0.1", GOOS: platform.LINUX},
		{Case: "Outdated Darwin", Expected: true, CurrentVersion: "3.0.0", LatestVersion: "v3.0.1", GOOS: platform.DARWIN},
		{Case: "Cached", Cache: true},
		{Case: "Error", Error: fmt.Errorf("error")},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Flags").Return(&platform.Flags{
			Version: tc.CurrentVersion,
		})
		cache := &mock.MockedCache{}
		cache.On("Get", UPGRADECACHEKEY).Return("", tc.Cache)
		cache.On("Set", mock2.Anything, mock2.Anything, mock2.Anything)
		env.On("Cache").Return(cache)
		env.On("GOOS").Return(tc.GOOS)

		json := fmt.Sprintf(`{"tag_name":"%s"}`, tc.LatestVersion)
		env.On("HTTPRequest", releaseURL).Return([]byte(json), tc.Error)
		// ignore the notice
		_, canUpgrade := Notice(env)
		assert.Equal(t, tc.Expected, canUpgrade, tc.Case)
	}
}
