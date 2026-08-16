package segments

import (
	"path/filepath"
)

type Angular struct {
	Language
}

func (a *Angular) Template() string {
	return languageTemplate
}

func (a *Angular) Enabled() bool {
	a.loadSpec()

	return a.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (a *Angular) Activation() Activation {
	a.loadSpec()

	return a.activation()
}

func (a *Angular) loadSpec() {
	a.extensions = []string{"angular.json"}
	a.tooling = map[string]*cmd{
		"angular": {
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: a.getVersion,
		},
	}
	a.defaultTooling = []string{"angular"}
	a.versionURLTemplate = "https://github.com/angular/angular/releases/tag/{{.Full}}"
}

func (a *Angular) getVersion() (string, error) {
	return a.nodePackageVersion(filepath.Join("@angular", "core"))
}
