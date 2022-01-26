package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Ruby struct {
	language
}

func (r *Ruby) Template() string {
	return languageTemplate
}

func (r *Ruby) Init(props properties.Properties, env environment.Environment) {
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.rb", "Rakefile", "Gemfile"},
		commands: []*cmd{
			{
				executable: "rbenv",
				args:       []string{"version-name"},
				regex:      `(?P<version>.+)`,
			},
			{
				executable: "rvm-prompt",
				args:       []string{"i", "v", "g"},
				regex:      `(?P<version>.+)`,
			},
			{
				executable: "chruby",
				args:       []string(nil),
				regex:      `\* (?P<version>.+)\n`,
			},
			{
				executable: "asdf",
				args:       []string{"current", "ruby"},
				regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
			},
			{
				executable: "ruby",
				args:       []string{"--version"},
				regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
			},
		},
	}
}

func (r *Ruby) Enabled() bool {
	enabled := r.language.Enabled()
	// this happens when no version is set
	if r.Full == "______" {
		r.Full = ""
	}
	return enabled
}
