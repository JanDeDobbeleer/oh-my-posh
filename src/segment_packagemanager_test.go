package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	availablePackageUpdates = "PackageManager_AvailablePackageUpdates"
)

func TestPMSegment(t *testing.T) {
	cases := []struct {
		Case                          string
		CachedAvailablePackageUpdates string
		PackageManagerType            string
		AvailablePackageUpdates       string
		ExpectedString                string
		ExpectedEnabled               bool
		CacheFoundFail                bool
		ChocoOutdatedOutput           string
		ErrorMessage                  string
	}{
		{
			Case:                          "No updates available - not cached",
			CachedAvailablePackageUpdates: "",
			PackageManagerType:            "chocolatey",
			CacheFoundFail:                true,
			ChocoOutdatedOutput:           "Chocolatey v0.11.3\nOutdated Packages\nOutput is package name | current version | available version | pinned?\n\n\nChocolatey has determined 0 package(s) are outdated.",
			AvailablePackageUpdates:       "0",
			ExpectedEnabled:               false,
		},
		{
			Case:                          "No updates available - cached",
			CachedAvailablePackageUpdates: "0",
			CacheFoundFail:                false,
			AvailablePackageUpdates:       "0",
			ExpectedEnabled:               false,
		},
		{
			Case:                          "1 update available - not cached",
			PackageManagerType:            "chocolatey",
			CachedAvailablePackageUpdates: "",
			CacheFoundFail:                true,
			ChocoOutdatedOutput:           "Chocolatey v0.11.3\nOutdated Packages\nOutput is package name | current version | available version | pinned?\n\noh-my-posh|6.37.1|6.39.0|false\n\nChocolatey has determined 1 package(s) are outdated.",
			AvailablePackageUpdates:       "1",
			ExpectedString:                "\uf487 ↓1",
			ExpectedEnabled:               true,
		},
		{
			Case:                          "1 update available - cached",
			CachedAvailablePackageUpdates: "1",
			CacheFoundFail:                false,
			AvailablePackageUpdates:       "1",
			ExpectedString:                "\uf487 ↓1",
			ExpectedEnabled:               true,
		},
		{
			Case:                          "Bad choco outdated output",
			PackageManagerType:            "chocolatey",
			CachedAvailablePackageUpdates: "",
			CacheFoundFail:                true,
			ChocoOutdatedOutput:           "This is bad output",
			ExpectedEnabled:               false,
			ErrorMessage:                  "Unexpected output from choco outdated",
		},
		{
			Case:                          "Non-numeric available package updates - not cached",
			PackageManagerType:            "chocolatey",
			CachedAvailablePackageUpdates: "",
			CacheFoundFail:                true,
			ChocoOutdatedOutput:           "Chocolatey v0.11.3\nOutdated Packages\nOutput is package name | current version | available version | pinned?\n\noh-my-posh|6.37.1|6.39.0|false\n\nChocolatey has determined X package(s) are outdated.",
			ExpectedEnabled:               false,
			ErrorMessage:                  "Available package updates value is non-numeric",
		},
		{
			Case:                          "Unknown package manager type",
			PackageManagerType:            "nonExistentPackageManager",
			CachedAvailablePackageUpdates: "",
			CacheFoundFail:                true,
			ExpectedEnabled:               false,
			ErrorMessage:                  "Unknown package manager",
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		var props properties = map[Property]interface{}{
			"PackageManagerType": tc.PackageManagerType,
		}

		cache := &MockedCache{}
		cache.On("get", availablePackageUpdates).Return(tc.CachedAvailablePackageUpdates, !tc.CacheFoundFail)
		cache.On("set", availablePackageUpdates, tc.AvailablePackageUpdates, 5).Return()
		env.On("cache", nil).Return(cache)
		env.On("runShellCommand", "powershell", "choco outdated").Return(tc.ChocoOutdatedOutput)

		pm := &packagemanager{
			props: props,
			env:   env,
		}

		enabled := pm.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		if pm.Error != nil {
			assert.Equal(t, pm.Error.Error(), tc.ErrorMessage, pm.string(), tc.Case)
		} else {
			assert.Equal(t, tc.ErrorMessage, "", pm.string(), tc.Case)
		}

		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, pm.string(), tc.Case)
	}
}
