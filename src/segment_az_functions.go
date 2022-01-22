package main

type azfunc struct {
	language
}

func (az *azfunc) string() string {
	segmentTemplate := az.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		return az.language.string()
	}
	return az.language.renderTemplate(segmentTemplate, az)
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
