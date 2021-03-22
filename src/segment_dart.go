package main

type dart struct {
	language *language
}

func (d *dart) string() string {
	return d.language.string()
}

func (d *dart) init(props *properties, env environmentInfo) {
	d.language = &language{
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
		versionURLTemplate: "[%s](https://dart.dev/guides/language/evolution#dart-%s%s)",
	}
}

func (d *dart) enabled() bool {
	return d.language.enabled()
}
