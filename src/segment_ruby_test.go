package main

import (
	"fmt"
	"testing"

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
		env := new(MockedEnvironment)
		env.On("hasCommand", "rbenv").Return(tc.HasRbenv)
		env.On("runCommand", "rbenv", []string{"version-name"}).Return(tc.Version, nil)
		env.On("hasCommand", "rvm-prompt").Return(tc.HasRvmprompt)
		env.On("runCommand", "rvm-prompt", []string{"i", "v", "g"}).Return(tc.Version, nil)
		env.On("hasCommand", "chruby").Return(tc.HasChruby)
		env.On("runCommand", "chruby", []string(nil)).Return(tc.Version, nil)
		env.On("hasCommand", "asdf").Return(tc.HasAsdf)
		env.On("runCommand", "asdf", []string{"current", "ruby"}).Return(tc.Version, nil)
		env.On("hasCommand", "ruby").Return(tc.HasRuby)
		env.On("runCommand", "ruby", []string{"--version"}).Return(tc.Version, nil)
		env.On("hasFiles", "*.rb").Return(tc.HasRubyFiles)
		env.On("hasFiles", "Rakefile").Return(tc.HasRakeFile)
		env.On("hasFiles", "Gemfile").Return(tc.HasGemFile)
		env.On("getcwd", nil).Return("/usr/home/project")
		env.On("homeDir", nil).Return("/usr/home")
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
