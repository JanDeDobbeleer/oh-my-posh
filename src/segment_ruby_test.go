package main

import (
	"fmt"
	"testing"

	"oh-my-posh/runtime"

	"github.com/stretchr/testify/assert"
)

func TestRuby(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		HasRbenv        bool
		HasRvmprompt    bool
		HasChruby       bool
		HasAsdf         bool
		HasRuby         bool
		Version         string
		HasRubyFiles    bool
		HasRakeFile     bool
		HasGemFile      bool
		DisplayVersion  bool
	}{
		{Case: "No files", ExpectedString: "", ExpectedEnabled: false},
		{Case: "Ruby files", ExpectedString: "", ExpectedEnabled: true, DisplayVersion: false, HasRubyFiles: true},
		{Case: "Rakefile", ExpectedString: "", ExpectedEnabled: true, DisplayVersion: false, HasRakeFile: true},
		{Case: "Gemfile", ExpectedString: "", ExpectedEnabled: true, DisplayVersion: false, HasGemFile: true},
		{Case: "Gemfile with version", ExpectedString: "", ExpectedEnabled: true, DisplayVersion: true, HasGemFile: true},
		{Case: "No files with version", ExpectedString: "", ExpectedEnabled: false, DisplayVersion: true},
		{
			Case:            "Version with chruby",
			ExpectedString:  "ruby-2.6.3",
			ExpectedEnabled: true,
			DisplayVersion:  true,
			HasRubyFiles:    true,
			HasChruby:       true,
			Version: ` * ruby-2.6.3
			ruby-1.9.3-p392
			jruby-1.7.0
			rubinius-2.0.0-rc1`,
		},
		{
			Case:            "Version with chruby line 2",
			ExpectedString:  "ruby-1.9.3-p392",
			ExpectedEnabled: true,
			DisplayVersion:  true,
			HasRubyFiles:    true,
			HasChruby:       true,
			Version: ` ruby-2.6.3
			* ruby-1.9.3-p392
			jruby-1.7.0
			rubinius-2.0.0-rc1`,
		},
		{
			Case:            "Version with asdf",
			ExpectedString:  "2.6.3",
			ExpectedEnabled: true,
			DisplayVersion:  true,
			HasRubyFiles:    true,
			HasAsdf:         true,
			Version:         "ruby            2.6.3           /Users/jan/Projects/oh-my-posh/.tool-versions",
		},
		{
			Case:            "Version with asdf not set",
			ExpectedString:  "",
			ExpectedEnabled: true,
			DisplayVersion:  true,
			HasRubyFiles:    true,
			HasAsdf:         true,
			Version:         "ruby            ______          No version set. Run \"asdf <global|shell|local> ruby <version>\"",
		},
		{
			Case:            "Version with ruby",
			ExpectedString:  "2.6.3",
			ExpectedEnabled: true,
			DisplayVersion:  true,
			HasRubyFiles:    true,
			HasRuby:         true,
			Version:         "ruby  2.6.3 (2019-04-16 revision 67580) [universal.x86_64-darwin20]",
		},
	}
	for _, tc := range cases {
		env := new(runtime.MockedEnvironment)
		env.On("HasCommand", "rbenv").Return(tc.HasRbenv)
		env.On("RunCommand", "rbenv", []string{"version-name"}).Return(tc.Version, nil)
		env.On("HasCommand", "rvm-prompt").Return(tc.HasRvmprompt)
		env.On("RunCommand", "rvm-prompt", []string{"i", "v", "g"}).Return(tc.Version, nil)
		env.On("HasCommand", "chruby").Return(tc.HasChruby)
		env.On("RunCommand", "chruby", []string(nil)).Return(tc.Version, nil)
		env.On("HasCommand", "asdf").Return(tc.HasAsdf)
		env.On("RunCommand", "asdf", []string{"current", "ruby"}).Return(tc.Version, nil)
		env.On("HasCommand", "ruby").Return(tc.HasRuby)
		env.On("RunCommand", "ruby", []string{"--version"}).Return(tc.Version, nil)
		env.On("HasFiles", "*.rb").Return(tc.HasRubyFiles)
		env.On("HasFiles", "Rakefile").Return(tc.HasRakeFile)
		env.On("HasFiles", "Gemfile").Return(tc.HasGemFile)
		env.On("Getcwd", nil).Return("/usr/home/project")
		env.On("HomeDir", nil).Return("/usr/home")
		props := &properties{
			values: map[Property]interface{}{
				DisplayVersion: tc.DisplayVersion,
			},
		}
		ruby := &ruby{}
		ruby.init(props, env)
		assert.Equal(t, tc.ExpectedEnabled, ruby.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, ruby.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
