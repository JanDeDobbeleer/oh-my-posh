package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestRuby(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		HasRbenv       bool
		HasRvmprompt   bool
		HasChruby      bool
		HasAsdf        bool
		HasRuby        bool
		Version        string
		HasRubyFiles   bool
		HasRakeFile    bool
		HasGemFile     bool
		FetchVersion   bool
	}{
		{Case: "Ruby files", ExpectedString: "", FetchVersion: false, HasRubyFiles: true},
		{Case: "Rakefile", ExpectedString: "", FetchVersion: false, HasRakeFile: true},
		{Case: "Gemfile", ExpectedString: "", FetchVersion: false, HasGemFile: true},
		{Case: "Gemfile with version", ExpectedString: "err parsing info from ruby with", FetchVersion: true, HasGemFile: true},
		{
			Case:           "Version with chruby",
			ExpectedString: "ruby-2.6.3",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasChruby:      true,
			Version: ` * ruby-2.6.3
			ruby-1.9.3-p392
			jruby-1.7.0
			rubinius-2.0.0-rc1`,
		},
		{
			Case:           "Version with chruby line 2",
			ExpectedString: "ruby-1.9.3-p392",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasChruby:      true,
			Version: ` ruby-2.6.3
			* ruby-1.9.3-p392
			jruby-1.7.0
			rubinius-2.0.0-rc1`,
		},
		{
			Case:           "Version with asdf",
			ExpectedString: "2.6.3",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasAsdf:        true,
			Version:        "ruby            2.6.3           /Users/jan/Projects/oh-my-posh/.tool-versions",
		},
		{
			Case:           "Version with asdf not set",
			ExpectedString: "",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasAsdf:        true,
			Version:        "ruby            ______          No version set. Run \"asdf <global|shell|local> ruby <version>\"",
		},
		{
			Case:           "Version with ruby",
			ExpectedString: "2.6.3",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasRuby:        true,
			Version:        "ruby  2.6.3 (2019-04-16 revision 67580) [universal.x86_64-darwin20]",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "ruby",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.rb",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("HasCommand", "rbenv").Return(tc.HasRbenv)
		env.On("RunCommand", "rbenv", []string{"version-name"}).Return(tc.Version, nil)
		env.On("HasCommand", "rvm-prompt").Return(tc.HasRvmprompt)
		env.On("RunCommand", "rvm-prompt", []string{"i", "v", "g"}).Return(tc.Version, nil)
		env.On("HasCommand", "chruby").Return(tc.HasChruby)
		env.On("RunCommand", "chruby", []string(nil)).Return(tc.Version, nil)
		env.On("HasCommand", "asdf").Return(tc.HasAsdf)
		env.On("RunCommand", "asdf", []string{"current", "ruby"}).Return(tc.Version, nil)
		env.On("HasFiles", "Rakefile").Return(tc.HasRakeFile)
		env.On("HasFiles", "Gemfile").Return(tc.HasGemFile)

		props[properties.FetchVersion] = tc.FetchVersion

		ruby := &Ruby{}
		ruby.Init(props, env)

		assert.True(t, ruby.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, ruby.Template(), ruby), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
