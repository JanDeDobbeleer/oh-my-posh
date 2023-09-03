package segments

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	testify_mock "github.com/stretchr/testify/mock"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

func TestHelmSegment(t *testing.T) {
	cases := []struct {
		Case            string
		HelmExists      bool
		ExpectedEnabled bool
		ExpectedString  string
		Template        string
		DisplayMode     string
		ChartFile       string
	}{
		{
			Case:            "Helm not installed",
			HelmExists:      false,
			ExpectedEnabled: false,
		},
		{
			Case:            "DisplayMode always inside chart",
			HelmExists:      true,
			ExpectedEnabled: true,
			ExpectedString:  "Helm 3.12.3",
			DisplayMode:     "always",
		},
		{
			Case:            "DisplayMode always outside chart",
			HelmExists:      true,
			ExpectedEnabled: true,
			ExpectedString:  "Helm 3.12.3",
			DisplayMode:     "always",
		},
		{
			Case:            "DisplayMode files inside chart. Chart file Chart.yml",
			HelmExists:      true,
			ExpectedEnabled: true,
			ExpectedString:  "Helm 3.12.3",
			DisplayMode:     "files",
			ChartFile:       "Chart.yml",
		},
		{
			Case:            "DisplayMode always inside chart. Chart file Chart.yaml",
			HelmExists:      true,
			ExpectedEnabled: true,
			ExpectedString:  "Helm 3.12.3",
			DisplayMode:     "files",
			ChartFile:       "Chart.yaml",
		},
		{
			Case:            "DisplayMode always outside chart",
			HelmExists:      true,
			ExpectedEnabled: false,
			DisplayMode:     "files",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("HasCommand", "helm").Return(tc.HelmExists)
		env.On("RunCommand", "helm", []string{"version", "--short", "--template={{.Version}}"}).Return("v3.12.3", nil)

		env.On("HasParentFilePath", tc.ChartFile).Return(&platform.FileInfo{}, nil)
		env.On("HasParentFilePath", testify_mock.Anything).Return(&platform.FileInfo{}, errors.New("no such file or directory"))

		props := properties.Map{
			DisplayMode: tc.DisplayMode,
		}

		h := &Helm{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.ExpectedEnabled, h.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, h.Template(), h), tc.Case)
		}
	}
}
