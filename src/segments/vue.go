package segments

type Vue struct {
	language
}

func (v *Vue) Template() string {
	return languageTemplate
}

func (v *Vue) Enabled() bool {
	v.extensions = []string{"package.json"}
	v.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: v.getVersion,
		},
	}
	v.versionURLTemplate = "https://github.com/vuejs/core/releases/tag/v{{.Full}}"

	if !v.hasNodePackage("vue") {
		return false
	}

	return v.language.Enabled()
}

func (v *Vue) getVersion() (string, error) {
	return v.nodePackageVersion("vue")
}
