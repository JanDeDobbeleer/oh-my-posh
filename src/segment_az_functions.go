package main

type azfunc struct {
	language
}

func (az *azfunc) string() string {
	segmentTemplate := az.language.props.getString(SegmentTemplate, "{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}")
	return az.language.string(segmentTemplate, az)
}

func (az *azfunc) init(props Properties, env Environment) {
	az.language = language{
		env:        env,
		props:      props,
		extensions: []string{"host.json", "local.settings.json", "function.json"},
		commands: []*cmd{
			{
				executable: "func",
				args:       []string{"--version"},
				regex:      `(?P<version>[0-9.]+)`,
			},
		},
	}
}

func (az *azfunc) enabled() bool {
	return az.language.enabled()
}
