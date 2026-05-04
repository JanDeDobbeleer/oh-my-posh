package segments

type OCaml struct {
	Language
}

func (o *OCaml) Template() string {
	return languageTemplate
}

func (o *OCaml) Enabled() bool {
	const ocamlToolName = "ocaml"

	o.extensions = []string{"*.ml", "*.mli", "dune", "dune-project", "dune-workspace"}
	o.tooling = map[string]*cmd{
		ocamlToolName: {
			executable: ocamlToolName,
			args:       []string{versionFlagShortArg},
			regex:      `The OCaml toplevel, version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)`,
		},
	}
	o.defaultTooling = []string{ocamlToolName}

	return o.Language.Enabled()
}
