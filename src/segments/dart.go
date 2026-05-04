package segments

var (
	dartExtensions = []string{"*.dart", "pubspec.yaml", "pubspec.yml", "pubspec.lock"}
	dartFolders    = []string{".dart_tool"}
)

type Dart struct {
	Language
}

func (d *Dart) Template() string {
	return languageTemplate
}

func (d *Dart) Enabled() bool {
	d.extensions = dartExtensions
	d.folders = dartFolders
	d.tooling = map[string]*cmd{
		fvmToolName: {
			executable: fvmToolName,
			args:       []string{dartToolName, versionFlagArg},
			regex:      `Dart SDK version: ` + versionRegex,
		},
		dartToolName: {
			executable: dartToolName,
			args:       []string{versionFlagArg},
			regex:      `Dart SDK version: ` + versionRegex,
		},
	}
	d.defaultTooling = []string{fvmToolName, dartToolName}
	d.versionURLTemplate = "https://dart.dev/guides/language/evolution#dart-{{ .Major }}{{ .Minor }}"

	return d.Language.Enabled()
}
