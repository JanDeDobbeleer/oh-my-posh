package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

type Rust struct {
	language
}

func (r *Rust) Template() string {
	return languageTemplate
}

func (r *Rust) Init(props properties.Properties, env platform.Environment) {
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.rs", "Cargo.toml", "Cargo.lock"},
		commands: []*cmd{
			{
				executable: "rustc",
				args:       []string{"--version"},
				regex:      `rustc (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)(( \((?P<buildmetadata>[0-9a-f]+ [0-9]+-[0-9]+-[0-9]+)\))?)`,
			},
		},
	}
}

func (r *Rust) Enabled() bool {
	return r.language.Enabled()
}
