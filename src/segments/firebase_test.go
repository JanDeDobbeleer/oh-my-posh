package segments

import (
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestFirebaseSegment(t *testing.T) {
	config := `{
		"activeProjects": {
			"path": "project-name"
		}
	}`
	cases := []struct {
		Case            string
		ActiveConfig    string
		ActivePath      string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "happy path",
			ExpectedEnabled: true,
			ActiveConfig:    config,
			ActivePath:      "path",
			ExpectedString:  "project-name",
		},
		{
			Case:            "happy subpath",
			ExpectedEnabled: true,
			ActiveConfig:    config,
			ActivePath:      "path/subpath",
			ExpectedString:  "project-name",
		},
		{
			Case:            "no active config",
			ExpectedEnabled: false,
		},
		{
			Case:            "empty config",
			ActiveConfig:    "{}",
			ExpectedEnabled: false,
		},
		{
			Case:            "bad config",
			ActiveConfig:    "{bad}",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return("home")
		env.On("Pwd").Return(tc.ActivePath)
		fcPath := filepath.Join("home", ".config", "configstore", "firebase-tools.json")
		env.On("FileContent", fcPath).Return(tc.ActiveConfig)
		env.On("Error", testify_.Anything).Return()
		f := Firebase{
			env: env,
		}
		f.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, f.Enabled())
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, f.Template(), f), tc.Case)
		}
	}
}

func TestGetFirebaseActiveConfig(t *testing.T) {
	data :=
		`{
			"activeProjects": {
				"path": "project-name"
			}
		}`
	cases := []struct {
		Case           string
		ActiveConfig   string
		ExpectedString string
		ExpectedError  string
	}{
		{
			Case:           "happy path",
			ActiveConfig:   data,
			ExpectedString: data,
		},
		{
			Case:          "no active config",
			ActiveConfig:  "",
			ExpectedError: FIREBASENOACTIVECONFIG,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return("home")
		configPath := filepath.Join("home", ".config", "configstore")
		contentPath := filepath.Join(configPath, "firebase-tools.json")
		env.On("FileContent", contentPath).Return(tc.ActiveConfig)
		env.On("Error", testify_.Anything).Return()
		f := Firebase{
			env: env,
		}
		got, err := f.getActiveConfig(configPath)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
		if len(tc.ExpectedError) > 0 {
			assert.EqualError(t, err, tc.ExpectedError, tc.Case)
		} else {
			assert.NoError(t, err, tc.Case)
		}
	}
}
