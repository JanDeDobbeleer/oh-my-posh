package segments

import (
	"path/filepath"
)

type Tauri struct {
	language
}

func (t *Tauri) Template() string {
	return languageTemplate
}

func (t *Tauri) Enabled() bool {
	t.extensions = []string{"package.json"}
	t.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: t.getVersion,
		},
	}
	t.versionURLTemplate = "https://github.com/tauri-apps/tauri/releases/tag/tauri-v{{.Full}}"

	return t.language.Enabled()
}

func (t *Tauri) getVersion() (string, error) {
	return t.nodePackageVersion(filepath.Join("@tauri-apps", "api"))
}
