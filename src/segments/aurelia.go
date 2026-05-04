package segments

type Aurelia struct {
	Language
}

func (a *Aurelia) Template() string {
	return languageTemplate
}

func (a *Aurelia) Enabled() bool {
	a.extensions = []string{fileName}
	a.tooling = map[string]*cmd{
		"aurelia": {
			regex:      versionRegexPrefixed,
			getVersion: a.getVersion,
		},
	}
	a.defaultTooling = []string{"aurelia"}
	a.versionURLTemplate = "https://github.com/aurelia/aurelia/releases/tag/v{{ .Full }}"

	if !a.hasNodePackage("aurelia") {
		return false
	}

	return a.Language.Enabled()
}

func (a *Aurelia) getVersion() (string, error) {
	return a.nodePackageVersion("aurelia")
}
