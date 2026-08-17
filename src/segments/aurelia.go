//nolint:dupl // react and aurelia are deliberately parallel: identical node-package detection, differing only in package name and version metadata
package segments

type Aurelia struct {
	Language
}

func (a *Aurelia) Template() string {
	return languageTemplate
}

func (a *Aurelia) Enabled() bool {
	a.loadSpec()

	if !a.hasNodePackage("aurelia") {
		return false
	}

	return a.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (a *Aurelia) Activation() Activation {
	a.loadSpec()

	return a.activation()
}

func (a *Aurelia) loadSpec() {
	a.extensions = []string{fileName}
	a.tooling = map[string]*cmd{
		"aurelia": {
			regex:      versionRegexSemver,
			getVersion: a.getVersion,
		},
	}
	a.defaultTooling = []string{"aurelia"}
	a.versionURLTemplate = "https://github.com/aurelia/aurelia/releases/tag/v{{ .Full }}"
}

func (a *Aurelia) getVersion() (string, error) {
	return a.nodePackageVersion("aurelia")
}
