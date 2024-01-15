package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type OCaml struct {
	language
}

func (o *OCaml) Template() string {
	return languageTemplate
}

func (o *OCaml) Init(props properties.Properties, env platform.Environment) {
	o.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.ml", "*.mli", "dune", "dune-project", "dune-workspace"},
		commands: []*cmd{
			{
				executable: "ocaml",
				args:       []string{"-version"},
				regex:      `The OCaml toplevel, version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)`,
			},
		},
	}
}

func (o *OCaml) Enabled() bool {
	return o.language.Enabled()
}
