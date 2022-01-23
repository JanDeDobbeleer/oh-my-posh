package main

type dart struct {
	language
}

func (d *dart) string() string {
	segmentTemplate := d.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		return d.language.string()
	}
	return d.language.renderTemplate(segmentTemplate, d)
}

func (d *dart) init(props Properties, env Environment) {
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

func (d *dart) enabled() bool {
	return d.language.enabled()
}
