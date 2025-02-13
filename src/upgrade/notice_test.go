package upgrade

import (
	"os"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	cache_ "github.com/jandedobbeleer/oh-my-posh/src/cache/mock"
	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestCanUpgrade(t *testing.T) {
	ugc := &Config{}
	latest, _ := ugc.Latest()

	cases := []struct {
		Case           string
		CurrentVersion string
		Installer      string
		Expected       bool
		Cache          bool
	}{
		{Case: "Up to date", CurrentVersion: latest},
		{Case: "Outdated Linux", Expected: true, CurrentVersion: "3.0.0"},
		{Case: "Outdated Darwin", Expected: true, CurrentVersion: "3.0.0"},
		{Case: "Cached", Cache: true, CurrentVersion: latest},
		{Case: "Windows Store", Installer: "ws"},
	}

	for _, tc := range cases {
		build.Version = tc.CurrentVersion
		c := &cache_.Cache{}
		c.On("Get", CACHEKEY).Return("", tc.Cache)
		c.On("Set", testify_.Anything, testify_.Anything, testify_.Anything)
		ugc.Cache = c

		if len(tc.Installer) > 0 {
			os.Setenv("POSH_INSTALLER", tc.Installer)
		}

		_, canUpgrade := ugc.Notice()
		assert.Equal(t, tc.Expected, canUpgrade, tc.Case)

		os.Setenv("POSH_INSTALLER", "")
	}
}
