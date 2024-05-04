package segments

import (
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestTalosctlSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ActiveConfig    string
		ExpectedEnabled bool
		ExpectedString  string
	}{
		{
			Case:            "happy path",
			ExpectedEnabled: true,
			ActiveConfig:    "context: context-name",
			ExpectedString:  "context-name",
		},
		{
			Case:            "no active config",
			ExpectedEnabled: false,
		},
		{
			Case:            "empty config",
			ActiveConfig:    "",
			ExpectedEnabled: false,
		},
		{
			Case:            "bad config",
			ActiveConfig:    "other-yaml: not-expected",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return("home")
		fcPath := filepath.Join("home", ".talos", "config")
		env.On("FileContent", fcPath).Return(tc.ActiveConfig)
		env.On("Error", mock2.Anything).Return()
		talos := TalosCTL{
			env: env,
		}
		talos.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, talos.Enabled())
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, talos.Template(), talos), tc.Case)
		}
	}
}

func TestGetTalosctlActiveConfig(t *testing.T) {
	cases := []struct {
		Case           string
		ActiveConfig   string
		ExpectedString string
		ExpectedError  string
	}{
		{
			Case:           "happy path",
			ActiveConfig:   "context: context-name",
			ExpectedString: "context: context-name",
		},
		{
			Case:          "no active config",
			ActiveConfig:  "",
			ExpectedError: "NO ACTIVE CONFIG FOUND",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return("home")
		configPath := filepath.Join("home", ".talos")
		contentPath := filepath.Join(configPath, "config")
		env.On("FileContent", contentPath).Return(tc.ActiveConfig)
		env.On("Error", mock2.Anything).Return()
		talos := TalosCTL{
			env: env,
		}
		got, err := talos.getActiveConfig(configPath)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
		if len(tc.ExpectedError) > 0 {
			assert.EqualError(t, err, tc.ExpectedError, tc.Case)
		} else {
			assert.NoError(t, err, tc.Case)
		}
	}
}
