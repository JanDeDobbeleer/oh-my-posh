package segments

type Cf struct {
	Language
}

func (c *Cf) Template() string {
	return languageTemplate
}

func (c *Cf) Enabled() bool {
	c.loadSpec()

	return c.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (c *Cf) Activation() Activation {
	c.loadSpec()

	return c.activation()
}

func (c *Cf) loadSpec() {
	c.extensions = []string{"manifest.yml", "mta.yaml"}
	c.tooling = map[string]*cmd{
		"cf": {
			executable: "cf",
			args:       []string{versionArg},
			regex:      versionRegexPrefixed,
		},
	}
	c.defaultTooling = []string{"cf"}
	c.displayMode = c.options.String(DisplayMode, DisplayModeFiles)
	c.versionURLTemplate = "https://github.com/cloudfoundry/cli/releases/tag/v{{ .Full }}"
}
