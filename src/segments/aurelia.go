package segments

type Aurelia struct {
	language
}

func (a *Aurelia) Template() string {
	return languageTemplate
}

func (a *Aurelia) Enabled() bool {
	a.extensions = []string{"package.json"}
	a.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(-(?P<prerelease>[a-z]+).(?P<buildmetadata>[0-9]+))?)))`,
			getVersion: a.getVersion,
		},
	}

	a.versionURLTemplate = "https://github.com/aurelia/aurelia/releases/tag/v{{ .Full }}"

	return a.language.Enabled()
}

func (a *Aurelia) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(a.language.env, "aurelia")
}
