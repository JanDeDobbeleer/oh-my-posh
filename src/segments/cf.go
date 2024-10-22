package segments

type Cf struct {
	language
}

func (c *Cf) Template() string {
	return languageTemplate
}

func (c *Cf) Enabled() bool {
	c.extensions = []string{"manifest.yml", "mta.yaml"}
	c.commands = []*cmd{
		{
			executable: "cf",
			args:       []string{"version"},
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	c.displayMode = c.props.GetString(DisplayMode, DisplayModeFiles)
	c.versionURLTemplate = "https://github.com/cloudfoundry/cli/releases/tag/v{{ .Full }}"

	return c.language.Enabled()
}
