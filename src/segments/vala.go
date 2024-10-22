package segments

type Vala struct {
	language
}

func (v *Vala) Template() string {
	return languageTemplate
}

func (v *Vala) Enabled() bool {
	v.extensions = []string{"*.vala"}
	v.commands = []*cmd{
		{
			executable: "vala",
			args:       []string{"--version"},
			regex:      `Vala (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	v.versionURLTemplate = "https://gitlab.gnome.org/GNOME/vala/raw/{{ .Major }}.{{ .Minor }}/NEWS"

	return v.language.Enabled()
}
