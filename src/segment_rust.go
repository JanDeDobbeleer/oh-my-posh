package main

import "oh-my-posh/runtime"

type rust struct {
	language *language
}

func (r *rust) string() string {
	return r.language.string()
}

func (r *rust) init(props *properties, env runtime.Environment) {
	r.language = &language{
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
