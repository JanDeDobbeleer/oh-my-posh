package segments

type Nx struct {
	language
}

func (a *Nx) Template() string {
	return languageTemplate
}

func (a *Nx) Enabled() bool {
	a.extensions = []string{"workspace.json", "nx.json"}
	a.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: a.getVersion,
		},
	}
	a.versionURLTemplate = "https://github.com/nrwl/nx/releases/tag/{{.Full}}"

	return a.language.Enabled()
}

func (a *Nx) getVersion() (string, error) {
	return a.nodePackageVersion("nx")
}
