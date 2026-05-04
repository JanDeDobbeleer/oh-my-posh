package segments

type Flutter struct {
	Language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

const flutterToolName = "flutter"

func (f *Flutter) Enabled() bool {
	f.extensions = dartExtensions
	f.folders = dartFolders
	f.tooling = map[string]*cmd{
		fvmToolName: {
			executable: fvmToolName,
			args:       []string{flutterToolName, versionFlagArg},
			regex:      `Flutter ` + versionRegex,
		},
		flutterToolName: {
			executable: flutterToolName,
			args:       []string{versionFlagArg},
			regex:      `Flutter ` + versionRegex,
		},
	}
	f.defaultTooling = []string{fvmToolName, flutterToolName}
	f.versionURLTemplate = "https://github.com/flutter/flutter/releases/tag/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"

	return f.Language.Enabled()
}
