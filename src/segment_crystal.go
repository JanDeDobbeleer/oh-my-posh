package main

type crystal struct {
	language
}

func (c *crystal) string() string {
	segmentTemplate := c.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		return c.language.string()
	}
	return c.language.renderTemplate(segmentTemplate, c)
}

func (c *crystal) init(props Properties, env Environment) {
	c.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.cr", "shard.yml"},
		commands: []*cmd{
			{
				executable: "crystal",
				args:       []string{"--version"},
				regex:      `Crystal (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/crystal-lang/crystal/releases/tag/{{ .Full }}",
	}
}

func (c *crystal) enabled() bool {
	return c.language.enabled()
}
