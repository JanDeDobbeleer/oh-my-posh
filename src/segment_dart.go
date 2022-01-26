package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Dart struct {
	language
}

func (d *Dart) template() string {
	return languageTemplate
}

func (d *Dart) init(props properties.Properties, env environment.Environment) {
	d.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.dart", "pubspec.yaml", "pubspec.yml", "pubspec.lock", ".dart_tool"},
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

func (d *Dart) enabled() bool {
	return d.language.enabled()
}
