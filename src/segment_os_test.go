package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSInfo(t *testing.T) {
	cases := []struct {
		Case              string
		ExpectedString    string
		GOOS              string
		WSLDistro         string
		Platform          string
		DisplayDistroName bool
	}{
		{
			Case:           "WSL debian - icon",
			ExpectedString: "WSL at \uf306",
			GOOS:           "linux",
			WSLDistro:      "debian",
			Platform:       "debian",
		},
		{
			Case:              "WSL debian - name",
			ExpectedString:    "WSL at burps",
			GOOS:              "linux",
			WSLDistro:         "burps",
			Platform:          "debian",
			DisplayDistroName: true,
		},
		{
			Case:           "plain linux - icon",
			ExpectedString: "\uf306",
			GOOS:           "linux",
			Platform:       "debian",
		},
		{
			Case:              "plain linux - name",
			ExpectedString:    "debian",
			GOOS:              "linux",
			Platform:          "debian",
			DisplayDistroName: true,
		},
		{
			Case:           "windows",
			ExpectedString: "windows",
			GOOS:           "windows",
		},
		{
			Case:           "darwin",
			ExpectedString: "darwin",
			GOOS:           "darwin",
		},
		{
			Case:           "unknown",
			ExpectedString: "unknown",
			GOOS:           "unknown",
		},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getRuntimeGOOS").Return(tc.GOOS)
		env.On("getenv", "WSL_DISTRO_NAME").Return(tc.WSLDistro)
		env.On("getPlatform").Return(tc.Platform)
		osInfo := &osInfo{
			env: env,
			props: properties{
				WSL:               "WSL",
				WSLSeparator:      " at ",
				DisplayDistroName: tc.DisplayDistroName,
				Windows:           "windows",
				MacOS:             "darwin",
			},
		}
		assert.Equal(t, tc.ExpectedString, osInfo.string(), tc.Case)
		if tc.WSLDistro != "" {
			assert.Equal(t, tc.WSLDistro, osInfo.os, tc.Case)
		} else if tc.Platform != "" {
			assert.Equal(t, tc.Platform, osInfo.os, tc.Case)
		} else {
			assert.Equal(t, tc.GOOS, osInfo.os, tc.Case)
		}
	}
}
