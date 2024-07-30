package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/alecthomas/assert"
)

func TestGitversion(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		Response        string
		HasGitversion   bool
		Template        string
		CommandError    error
	}{
		{Case: "GitVersion not installed"},
		{Case: "GitVersion installed, no GitVersion.yml file", HasGitversion: true, Response: "Cannot find the .git directory"},
		{
			Case:            "Version",
			ExpectedEnabled: true,
			ExpectedString:  "number",
			HasGitversion:   true,
			Response:        "{ \"FullSemVer\": \"0.1.0\", \"SemVer\": \"number\" }",
			Template:        "{{ .SemVer }}",
		},
		{
			Case:          "Command Error",
			HasGitversion: true,
			CommandError:  errors.New("error"),
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		env.On("HasCommand", "gitversion").Return(tc.HasGitversion)
		env.On("Pwd").Return("test-dir")
		env.On("RunCommand", "gitversion", []string{"-output", "json"}).Return(tc.Response, tc.CommandError)

		gitversion := &GitVersion{
			env:   env,
			props: properties.Map{},
		}
		if len(tc.Template) == 0 {
			tc.Template = gitversion.Template()
		}

		enabled := gitversion.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if enabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, gitversion), tc.Case)
		}
	}
}
