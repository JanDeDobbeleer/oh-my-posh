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

// Activation implements the activation gate; see Language.activation.
func (f *Flutter) Activation() Activation {
	f.loadSpec()

	return f.activation()
}

func (f *Flutter) loadSpec() {
	f.extensions = []string{"*.dart", pubspecFileName, "pubspec.yml", "pubspec.lock"}
	f.folders = []string{".dart_tool"}
	f.tooling = map[string]*cmd{
		// Not marked versionCacheable: fvm (Flutter Version Management)
		// resolves the SDK to run from the project's .fvmrc/.fvm config, so
		// the same fvm binary reports a different Flutter version per
		// project.
		fvmToolName: {
			executable: fvmToolName,
			args:       []string{flutterToolName, versionFlagArg},
			regex:      `Flutter ` + versionRegex,
		},
		// Not marked versionCacheable: bin/flutter is a stable wrapper
		// script that dispatches into the SDK checkout next to it, and
		// `flutter upgrade`/`flutter channel` swap that SDK via git without
		// touching the script - so the wrapper's path/mtime/size cannot see
		// an upgrade (the same wrapper-dispatch class as mvnw/gradlew).
		flutterToolName: {
			executable: flutterToolName,
			args:       []string{versionFlagArg},
			regex:      `Flutter ` + versionRegex,
		},
	}
	f.defaultTooling = []string{fvmToolName, flutterToolName}
	f.versionURLTemplate = "https://github.com/flutter/flutter/releases/tag/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"
}
