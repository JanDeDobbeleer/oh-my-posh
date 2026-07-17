package segments

type Flutter struct {
	Language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

const flutterToolName = "flutter"

func (f *Flutter) Enabled() bool {
	f.extensions = []string{"*.dart", pubspecFileName, "pubspec.yml", "pubspec.lock"}
	f.folders = []string{".dart_tool"}
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
