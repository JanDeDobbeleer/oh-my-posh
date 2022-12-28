package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/mock"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"

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
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: make(map[string]string),
			WSL: tc.IsWSL,
		})
		osInfo := &Os{
			env: env,
			props: properties.Map{
				DisplayDistroName: tc.DisplayDistroName,
				Windows:           "windows",
				MacOS:             "darwin",
			},
		}
		_ = osInfo.Enabled()
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, osInfo.Template(), osInfo), tc.Case)
	}
}
