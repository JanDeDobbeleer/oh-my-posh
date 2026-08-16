package segments

type Flutter struct {
	Language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

const flutterToolName = "flutter"

func (f *Flutter) Enabled() bool {
	f.loadSpec()

	return f.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation. Flutter
// declares folders, so this resolves to Always unless the user empties them.
func (f *Flutter) Activation() Activation {
	f.loadSpec()

	return f.activation()
}

func (f *Flutter) loadSpec() {
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
}
