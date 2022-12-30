package segments

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/mock"
	"github.com/jandedobbeleer/oh-my-posh/properties"

	"github.com/stretchr/testify/assert"
)

func TestCFTargetSegment(t *testing.T) {
	cases := []struct {
		Case            string
		Template        string
		ExpectedString  string
		ExpectedEnabled bool
		TargetOutput    string
		CommandError    error
	}{
		{
			Case:            "1) not logged in to CF account",
			ExpectedString:  "",
			ExpectedEnabled: false,
			TargetOutput:    `Not logged in`,
			CommandError:    &exec.ExitError{},
		},
		{
			Case:            "2) logged in, default template",
			ExpectedString:  "12345678trial/dev",
			ExpectedEnabled: true,
			TargetOutput:    "API endpoint: https://api.cf.eu10.hana.ondemand.com\nAPI version: 3.109.0\nuser: user@some.com\norg: 12345678trial\nspace: dev",
			CommandError:    nil,
		},
		{
			Case:            "3) logged in, full template",
			Template:        "{{.URL}} {{.User}} {{.Org}} {{.Space}}",
			ExpectedString:  "https://api.cf.eu10.hana.ondemand.com user@some.com 12345678trial dev",
			ExpectedEnabled: true,
			TargetOutput:    "API endpoint: https://api.cf.eu10.hana.ondemand.com\nAPI version: 3.109.0\nuser: user@some.com\norg: 12345678trial\nspace: dev",
			CommandError:    nil,
		},
	}

	for _, tc := range cases {
		var env = new(mock.MockedEnvironment)
		env.On("HasCommand", "cf").Return(true)
		env.On("RunCommand", "cf", []string{"target"}).Return(tc.TargetOutput, tc.CommandError)
		env.On("Pwd", nil).Return("/usr/home/dev/my-app")
		env.On("Home", nil).Return("/usr/home")

		cfTarget := &CfTarget{}
		props := properties.Map{}

		if tc.Template == "" {
			tc.Template = cfTarget.Template()
		}

		cfTarget.Init(props, env)

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.Equal(t, tc.ExpectedEnabled, cfTarget.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, cfTarget), failMsg)
	}
}
