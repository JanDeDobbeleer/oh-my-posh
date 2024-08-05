package segments

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	testify_mock "github.com/stretchr/testify/mock"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
)

func TestHelmSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		Template        string
		DisplayMode     string
		ChartFile       string
		HelmExists      bool
		ExpectedEnabled bool
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
			Case:            "DisplayMode always inside chart. Chart file helmfile.yaml",
			HelmExists:      true,
			ExpectedEnabled: true,
			ExpectedString:  "Helm 3.12.3",
			DisplayMode:     "files",
			ChartFile:       "helmfile.yaml",
		},
		{
			Case:            "DisplayMode always inside chart. Chart file helmfile.yml",
			HelmExists:      true,
			ExpectedEnabled: true,
			ExpectedString:  "Helm 3.12.3",
			DisplayMode:     "files",
			ChartFile:       "helmfile.yml",
		},
		{
			Case:            "DisplayMode always outside chart",
			HelmExists:      true,
			ExpectedEnabled: false,
			DisplayMode:     "files",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", "helm").Return(tc.HelmExists)
		env.On("RunCommand", "helm", []string{"version", "--short", "--template={{.Version}}"}).Return("v3.12.3", nil)

		env.On("HasParentFilePath", tc.ChartFile, false).Return(&runtime.FileInfo{}, nil)
		env.On("HasParentFilePath", testify_mock.Anything, false).Return(&runtime.FileInfo{}, errors.New("no such file or directory"))

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
