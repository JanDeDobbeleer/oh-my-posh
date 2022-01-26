package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type rust struct {
	language
}

func (r *rust) template() string {
	return languageTemplate
}

func (r *rust) init(props properties.Properties, env environment.Environment) {
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.rs", "Cargo.toml", "Cargo.lock"},
		commands: []*cmd{
			{
				executable: "rustc",
				args:       []string{"--version"},
				regex:      `rustc (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
	}
}

func (r *rust) enabled() bool {
	return r.language.enabled()
}
