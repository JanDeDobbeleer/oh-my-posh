package segments

type Flutter struct {
	Language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

func (f *Flutter) Enabled() bool {
	f.extensions = dartExtensions
	f.folders = dartFolders
	f.commands = []*cmd{
		{
			executable: "fvm",
			args:       []string{"flutter", "--version"},
			regex:      `Flutter (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
		{
			executable: "flutter",
			args:       []string{"--version"},
			regex:      `Flutter (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	f.versionURLTemplate = "https://github.com/flutter/flutter/releases/tag/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"

	return f.Language.Enabled()
}
