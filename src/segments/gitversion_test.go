package segments

import (
	"errors"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/alecthomas/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestGitversion(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		Response        string
		CacheResponse   string
		CacheError      error
		HasGitversion   bool
		Template        string
		CommandError    error
		CacheTimeout    int
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
			Case:            "Cache Version",
			ExpectedEnabled: true,
			ExpectedString:  "number2",
			HasGitversion:   true,
			CacheResponse:   "{ \"FullSemVer\": \"0.1.2\", \"SemVer\": \"number2\" }",
			Response:        "{ \"FullSemVer\": \"0.1.0\", \"SemVer\": \"number\" }",
			Template:        "{{ .SemVer }}",
		},
		{
			Case:            "No Cache enabled",
			ExpectedEnabled: true,
			CacheTimeout:    -1,
			ExpectedString:  "number",
			HasGitversion:   true,
			CacheResponse:   "{ \"FullSemVer\": \"0.1.2\", \"SemVer\": \"number2\" }",
			Response:        "{ \"FullSemVer\": \"0.1.0\", \"SemVer\": \"number\" }",
			Template:        "{{ .SemVer }}",
		},
		{
			Case:          "Command Error",
			HasGitversion: true,
			CommandError:  errors.New("error"),
		},
		{
			Case:            "Bad cache",
			ExpectedEnabled: true,
			ExpectedString:  "number",
			HasGitversion:   true,
			CacheResponse:   "{{",
			Response:        "{ \"FullSemVer\": \"0.1.0\", \"SemVer\": \"number\" }",
			Template:        "{{ .SemVer }}",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		cache := &mock.MockedCache{}

		env.On("HasCommand", "gitversion").Return(tc.HasGitversion)
		env.On("Pwd").Return("test-dir")
		env.On("Cache").Return(cache)
		cache.On("Get", "test-dir").Return(tc.CacheResponse, len(tc.CacheResponse) != 0)
		cache.On("Set", mock2.Anything, mock2.Anything, mock2.Anything).Return(tc.Response, true)

		env.On("RunCommand", "gitversion", []string{"-output", "json"}).Return(tc.Response, tc.CommandError)

		if tc.CacheTimeout == 0 {
			tc.CacheTimeout = 30
		}
		gitversion := &GitVersion{
			env: env,
			props: properties.Map{
				properties.CacheTimeout: tc.CacheTimeout,
			},
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
