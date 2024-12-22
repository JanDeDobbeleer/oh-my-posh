package segments

type Vite struct {
	language
}

func (v *Vite) Template() string {
	return languageTemplate
}

func (v *Vite) Enabled() bool {
	v.extensions = []string{"vite.config.*"}
	v.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: v.getVersion,
		},
	}
	v.versionURLTemplate = "https://github.com/vitejs/vite/releases/tag/{{.Full}}"

	return v.language.Enabled()
}

func (v *Vite) getVersion() (string, error) {
	return v.nodePackageVersion("vite")
}
