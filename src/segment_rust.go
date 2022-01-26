package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Rust struct {
	language
}

func (r *Rust) template() string {
	return languageTemplate
}

func (r *Rust) init(props properties.Properties, env environment.Environment) {
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

func (r *Rust) enabled() bool {
	return r.language.enabled()
}
