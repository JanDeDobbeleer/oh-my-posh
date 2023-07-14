package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

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
		Icon              string
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
		{
			Case:           "crazy distro, specific icon",
			ExpectedString: "crazy distro",
			GOOS:           "linux",
			Platform:       "crazy",
			Icon:           "crazy distro",
		},
		{
			Case:           "crazy distro, not mapped",
			ExpectedString: "\uf17c",
			GOOS:           "linux",
			Platform:       "crazy",
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

		props := properties.Map{
			DisplayDistroName: tc.DisplayDistroName,
			Windows:           "windows",
			MacOS:             "darwin",
		}

		if len(tc.Icon) != 0 {
			props[properties.Property(tc.Platform)] = tc.Icon
		}

		osInfo := &Os{
			env:   env,
			props: props,
		}

		_ = osInfo.Enabled()
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, osInfo.Template(), osInfo), tc.Case)
	}
}
