package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestOSInfo(t *testing.T) {
	cases := []struct {
		Case              string
		ExpectedString    string
		GOOS              string
		Platform          string
		Icon              string
		IsWSL             bool
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
		{
			Case:              "show distro name, mapped",
			ExpectedString:    "<3",
			DisplayDistroName: true,
			GOOS:              "linux",
			Icon:              "<3",
			Platform:          "love",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return(tc.GOOS)
		env.On("Platform").Return(tc.Platform)
		env.On("TemplateCache").Return(&cache.Template{
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
