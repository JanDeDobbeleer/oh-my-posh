package main

type crystal struct {
	language
}

func (c *crystal) string() string {
	return c.language.string()
}

func (c *crystal) init(props properties, env environmentInfo) {
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
		versionURLTemplate: "[%s](https://github.com/crystal-lang/crystal/releases/tag/%s.%s.%s)",
	}
}

func (c *crystal) enabled() bool {
	return c.language.enabled()
}
