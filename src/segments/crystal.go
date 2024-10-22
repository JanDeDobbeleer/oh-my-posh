package segments

type Crystal struct {
	language
}

func (c *Crystal) Template() string {
	return languageTemplate
}

func (c *Crystal) Enabled() bool {
	c.extensions = []string{"*.cr", "shard.yml"}
	c.commands = []*cmd{
		{
			executable: "crystal",
			args:       []string{"--version"},
			regex:      `Crystal (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	c.versionURLTemplate = "https://github.com/crystal-lang/crystal/releases/tag/{{ .Full }}"

	return c.language.Enabled()
}
