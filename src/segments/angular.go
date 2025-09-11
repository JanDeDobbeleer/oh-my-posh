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
	a.extensions = []string{"angular.json"}
	a.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: a.getVersion,
		},
	}
	a.versionURLTemplate = "https://github.com/angular/angular/releases/tag/{{.Full}}"

	return a.Language.Enabled()
}

func (a *Angular) getVersion() (string, error) {
	return a.nodePackageVersion(filepath.Join("@angular", "core"))
}
