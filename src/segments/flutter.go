package segments

type Flutter struct {
	language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

func (f *Flutter) Enabled() bool {
	f.extensions = dartExtensions
	f.folders = dartFolders
	f.commands = []*cmd{
		{
			executable: "flutter",
			args:       []string{"--version"},
			regex:      `Flutter (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	f.versionURLTemplate = "https://github.com/flutter/flutter/releases/tag/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"

	return f.language.Enabled()
}
