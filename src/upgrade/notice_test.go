package upgrade

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	cache "github.com/jandedobbeleer/oh-my-posh/src/cache/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"

	testify "github.com/stretchr/testify/mock"
)

func TestCanUpgrade(t *testing.T) {
	cases := []struct {
		Error          error
		Case           string
		CurrentVersion string
		LatestVersion  string
		GOOS           string
		Installer      string
		Expected       bool
		Cache          bool
	}{
		{Case: "Up to date", CurrentVersion: "3.0.0", LatestVersion: "v3.0.0"},
		{Case: "Outdated Windows", Expected: true, CurrentVersion: "3.0.0", LatestVersion: "v3.0.1", GOOS: runtime.WINDOWS},
		{Case: "Outdated Linux", Expected: true, CurrentVersion: "3.0.0", LatestVersion: "v3.0.1", GOOS: runtime.LINUX},
		{Case: "Outdated Darwin", Expected: true, CurrentVersion: "3.0.0", LatestVersion: "v3.0.1", GOOS: runtime.DARWIN},
		{Case: "Cached", Cache: true},
		{Case: "Error", Error: fmt.Errorf("error")},
		{Case: "Windows Store", Installer: "ws"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		build.Version = tc.CurrentVersion
		c := &cache.Cache{}
		c.On("Get", CACHEKEY).Return("", tc.Cache)
		c.On("Set", testify.Anything, testify.Anything, testify.Anything)
		env.On("Cache").Return(c)
		env.On("GOOS").Return(tc.GOOS)
		env.On("Getenv", "POSH_INSTALLER").Return(tc.Installer)

		json := fmt.Sprintf(`{"tag_name":"%s"}`, tc.LatestVersion)
		env.On("HTTPRequest", RELEASEURL).Return([]byte(json), tc.Error)
		// ignore the notice
		_, canUpgrade := Notice(env, false)
		assert.Equal(t, tc.Expected, canUpgrade, tc.Case)
	}
}
