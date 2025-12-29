package segments

type Rust struct {
	Language
}

func (r *Rust) Template() string {
	return languageTemplate
}

func (r *Rust) Enabled() bool {
	r.extensions = []string{"*.rs", "Cargo.toml", "Cargo.lock"}
	r.tooling = map[string]*cmd{
		"rustc": {
			executable: "rustc",
			args:       []string{"--version"},
			regex:      `(rust version|rustc) (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)(( \((?P<buildmetadata>[0-9a-f]+ [0-9]+-[0-9]+-[0-9]+)\))?)`, //nolint:lll
		},
	}
	r.defaultTooling = []string{"rustc"}

	return r.Language.Enabled()
}
