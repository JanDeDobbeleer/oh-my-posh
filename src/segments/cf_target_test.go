package segments

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestCFTargetSegment(t *testing.T) {
	cases := []struct {
		Case           string
		Template       string
		ExpectedString string
		DisplayMode    string
		FileInfo       *platform.FileInfo
		TargetOutput   string
		CommandError   error
	}{
		{
			Case:         "not logged in to CF account",
			TargetOutput: `Not logged in`,
			CommandError: &exec.ExitError{},
		},
		{
			Case:           "logged in, default template",
			ExpectedString: "12345678trial/dev",
			TargetOutput:   "API endpoint: https://api.cf.eu10.hana.ondemand.com\nAPI version: 3.109.0\nuser: user@some.com\norg: 12345678trial\nspace: dev",
		},
		{
			Case: "no output from command",
		},
		{
			Case:           "logged in, full template",
			Template:       "{{.URL}} {{.User}} {{.Org}} {{.Space}}",
			ExpectedString: "https://api.cf.eu10.hana.ondemand.com user@some.com 12345678trial dev",
			TargetOutput:   "API endpoint: https://api.cf.eu10.hana.ondemand.com\nAPI version: 3.109.0\nuser: user@some.com\norg: 12345678trial\nspace: dev",
		},
		{
			Case:         "files and no manifest file",
			DisplayMode:  DisplayModeFiles,
			TargetOutput: "API endpoint: https://api.cf.eu10.hana.ondemand.com\nAPI version: 3.109.0\nuser: user@some.com\norg: 12345678trial\nspace: dev",
		},
		{
			Case:           "files and a manifest file",
			ExpectedString: "12345678trial/dev",
			DisplayMode:    DisplayModeFiles,
			FileInfo:       &platform.FileInfo{},
			TargetOutput:   "API endpoint: https://api.cf.eu10.hana.ondemand.com\nAPI version: 3.109.0\nuser: user@some.com\norg: 12345678trial\nspace: dev",
		},
		{
			Case:        "files and a manifest directory",
			DisplayMode: DisplayModeFiles,
			FileInfo: &platform.FileInfo{
				IsDir: true,
			},
		},
	}

	for _, tc := range cases {
		var env = new(mock.MockedEnvironment)
		env.On("HasCommand", "cf").Return(true)
		env.On("RunCommand", "cf", []string{"target"}).Return(tc.TargetOutput, tc.CommandError)
		env.On("Pwd", nil).Return("/usr/home/dev/my-app")
		env.On("Home", nil).Return("/usr/home")
		var err error
		if tc.FileInfo == nil {
			err = errors.New("no such file or directory")
		}
		env.On("HasParentFilePath", "manifest.yml").Return(tc.FileInfo, err)

		cfTarget := &CfTarget{}
		props := properties.Map{
			DisplayMode: tc.DisplayMode,
		}

		if tc.Template == "" {
			tc.Template = cfTarget.Template()
		}

		cfTarget.Init(props, env)

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.Equal(t, len(tc.ExpectedString) > 0, cfTarget.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, cfTarget), failMsg)
	}
}
