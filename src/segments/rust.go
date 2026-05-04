package segments

type Rust struct {
	Language
}

func (r *Rust) Template() string {
	return languageTemplate
}

func (r *Rust) Enabled() bool {
	const rustcToolName = "rustc"

	r.extensions = []string{"*.rs", "Cargo.toml", "Cargo.lock"}
	r.tooling = map[string]*cmd{
		rustcToolName: {
			executable: rustcToolName,
			args:       []string{versionFlagArg},
			regex:      `(rust version|rustc) (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)(( \((?P<buildmetadata>[0-9a-f]+ [0-9]+-[0-9]+-[0-9]+)\))?)`, //nolint:lll
		},
	}
	r.defaultTooling = []string{rustcToolName}

	return r.Language.Enabled()
}
