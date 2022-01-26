package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSInfo(t *testing.T) {
	cases := []struct {
		Case              string
		ExpectedString    string
		GOOS              string
		IsWSL             bool
		Platform          string
		DisplayDistroName bool
	}{
		{
			Case:           "WSL debian - icon",
			ExpectedString: "WSL at \uf306",
			GOOS:           "linux",
			IsWSL:          true,
			Platform:       "debian",
		},
		{
			Case:              "WSL debian - name",
			ExpectedString:    "WSL at debian",
			GOOS:              "linux",
			IsWSL:             true,
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
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return(tc.GOOS)
		env.On("Platform").Return(tc.Platform)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
			WSL: tc.IsWSL,
		})
		osInfo := &osInfo{
			env: env,
			props: properties{
				DisplayDistroName: tc.DisplayDistroName,
				Windows:           "windows",
				MacOS:             "darwin",
			},
		}
		_ = osInfo.enabled()
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, osInfo.template(), osInfo), tc.Case)
	}
}
