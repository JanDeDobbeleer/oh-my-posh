package segments

type Nx struct {
	Language
}

func (a *Nx) Template() string {
	return languageTemplate
}

func (a *Nx) Enabled() bool {
	a.extensions = []string{"workspace.json", "nx.json"}
	a.tooling = map[string]*cmd{
		"nx": {
			regex:      versionRegexPrefixed,
			getVersion: a.getVersion,
		},
	}
	a.defaultTooling = []string{"nx"}
	a.versionURLTemplate = "https://github.com/nrwl/nx/releases/tag/{{.Full}}"

	return a.Language.Enabled()
}

func (a *Nx) getVersion() (string, error) {
	return a.nodePackageVersion("nx")
}
