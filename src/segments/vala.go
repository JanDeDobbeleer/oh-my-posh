package segments

type Vala struct {
	Language
}

func (v *Vala) Template() string {
	return languageTemplate
}

func (v *Vala) Enabled() bool {
	const valaToolName = "vala"

	v.extensions = []string{"*.vala"}
	v.tooling = map[string]*cmd{
		valaToolName: {
			executable: valaToolName,
			args:       []string{versionFlagArg},
			regex:      `Vala (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	v.defaultTooling = []string{valaToolName}
	v.versionURLTemplate = "https://gitlab.gnome.org/GNOME/vala/raw/{{ .Major }}.{{ .Minor }}/NEWS"

	return v.Language.Enabled()
}
