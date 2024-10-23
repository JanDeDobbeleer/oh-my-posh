package segments

var (
	dartExtensions = []string{"*.dart", "pubspec.yaml", "pubspec.yml", "pubspec.lock"}
	dartFolders    = []string{".dart_tool"}
)

type Dart struct {
	language
}

func (d *Dart) Template() string {
	return languageTemplate
}

func (d *Dart) Enabled() bool {
	d.extensions = dartExtensions
	d.folders = dartFolders
	d.commands = []*cmd{
		{
			executable: "dart",
			args:       []string{"--version"},
			regex:      `Dart SDK version: (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	d.versionURLTemplate = "https://dart.dev/guides/language/evolution#dart-{{ .Major }}{{ .Minor }}"

	return d.language.Enabled()
}
