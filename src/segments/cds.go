package segments

type Cds struct {
	Language
	HasDependency bool
}

func (c *Cds) Template() string {
	return languageTemplate
}

func (c *Cds) Enabled() bool {
	c.extensions = []string{".cdsrc.json", ".cdsrc-private.json", "*.cds"}
	c.commands = []*cmd{
		{
			executable: "cds",
			args:       []string{"--version"},
			regex:      `@sap/cds: (?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	c.Language.loadContext = c.loadContext
	c.Language.inContext = c.inContext
	c.displayMode = c.options.String(DisplayMode, DisplayModeContext)

	return c.Language.Enabled()
}

func (c *Cds) loadContext() {
	if !c.hasNodePackage("@sap/cds") {
		return
	}

	c.HasDependency = true
}

func (c *Cds) inContext() bool {
	return c.HasDependency
}
