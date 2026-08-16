package segments

type Nx struct {
	Language
}

func (a *Nx) Template() string {
	return languageTemplate
}

func (a *Nx) Enabled() bool {
	a.loadSpec()

	return a.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (a *Nx) Activation() Activation {
	a.loadSpec()

	return a.activation()
}

func (a *Nx) loadSpec() {
	a.extensions = []string{"workspace.json", "nx.json"}
	a.tooling = map[string]*cmd{
		"nx": {
			regex:      versionRegexPrefixed,
			getVersion: a.getVersion,
		},
	}
	a.defaultTooling = []string{"nx"}
	a.versionURLTemplate = "https://github.com/nrwl/nx/releases/tag/{{.Full}}"
}

func (a *Nx) getVersion() (string, error) {
	return a.nodePackageVersion("nx")
}
