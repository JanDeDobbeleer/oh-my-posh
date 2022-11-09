package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

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

func (d *Dart) Init(props properties.Properties, env platform.Environment) {
	d.language = language{
		env:        env,
		props:      props,
		extensions: dartExtensions,
		folders:    dartFolders,
		commands: []*cmd{
			{
				executable: "dart",
				args:       []string{"--version"},
				regex:      `Dart SDK version: (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://dart.dev/guides/language/evolution#dart-{{ .Major }}{{ .Minor }}",
	}
}

func (d *Dart) Enabled() bool {
	return d.language.Enabled()
}
