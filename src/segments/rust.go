package segments

type Rust struct {
	language
}

func (r *Rust) Template() string {
	return languageTemplate
}

func (r *Rust) Enabled() bool {
	r.extensions = []string{"*.rs", "Cargo.toml", "Cargo.lock"}
	r.commands = []*cmd{
		{
			executable: "rustc",
			args:       []string{"--version"},
			regex:      `rustc (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)(( \((?P<buildmetadata>[0-9a-f]+ [0-9]+-[0-9]+-[0-9]+)\))?)`,
		},
	}

	return r.language.Enabled()
}
