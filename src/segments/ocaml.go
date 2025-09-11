package segments

type OCaml struct {
	Language
}

func (o *OCaml) Template() string {
	return languageTemplate
}

func (o *OCaml) Enabled() bool {
	o.extensions = []string{"*.ml", "*.mli", "dune", "dune-project", "dune-workspace"}
	o.commands = []*cmd{
		{
			executable: "ocaml",
			args:       []string{"-version"},
			regex:      `The OCaml toplevel, version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)`,
		},
	}

	return o.Language.Enabled()
}
