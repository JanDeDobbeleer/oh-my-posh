package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/alecthomas/assert"
)

func TestNbgv(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		ExpectedString  string
		Response        string
		Template        string
		ExpectedEnabled bool
		HasNbgv         bool
	}{
		{Case: "nbgv not installed"},
		{Case: "nbgv installed, no version file", HasNbgv: true, Response: "{ \"VersionFileFound\": false }"},
		{Case: "nbgv installed with version file", ExpectedEnabled: true, HasNbgv: true, Response: "{ \"VersionFileFound\": true }"},
		{
			Case:            "invalid template",
			ExpectedEnabled: true,
			ExpectedString:  "invalid template text",
			HasNbgv:         true,
			Response:        "{ \"VersionFileFound\": true }",
			Template:        "{{ err }}",
		},
		{
			Case:    "command error",
			HasNbgv: true,
			Error:   errors.New("oh noes"),
		},
		{
			Case:     "invalid json",
			HasNbgv:  true,
			Response: "><<<>>>",
		},
		{
			Case:            "Version",
			ExpectedEnabled: true,
			ExpectedString:  "bump",
			HasNbgv:         true,
			Response:        "{ \"VersionFileFound\": true, \"Version\": \"bump\" }",
			Template:        "{{ .Version }}",
		},
		{
			Case:            "AssemblyVersion",
			ExpectedEnabled: true,
			ExpectedString:  "bump",
			HasNbgv:         true,
			Response:        "{ \"VersionFileFound\": true, \"AssemblyVersion\": \"bump\" }",
			Template:        "{{ .AssemblyVersion }}",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", "nbgv").Return(tc.HasNbgv)
		env.On("RunCommand", "nbgv", []string{"get-version", "--format=json"}).Return(tc.Response, tc.Error)
		nbgv := &Nbgv{
			env:   env,
			props: properties.Map{},
		}
		enabled := nbgv.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if tc.Template == "" {
			tc.Template = nbgv.Template()
		}
		if enabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, nbgv), tc.Case)
		}
	}
}
