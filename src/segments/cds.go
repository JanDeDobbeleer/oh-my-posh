package segments

type Cds struct {
	Language
	HasDependency bool
}

func (c *Cds) Template() string {
	return languageTemplate
}

const cdsToolName = "cds"

func (c *Cds) Enabled() bool {
	c.extensions = []string{".cdsrc.json", ".cdsrc-private.json", "*.cds"}
	c.tooling = map[string]*cmd{
		cdsToolName: {
			executable: cdsToolName,
			args:       []string{versionFlagArg},
			regex:      `@sap/cds: ` + versionRegexPrefixed,
		},
	}
	c.defaultTooling = []string{cdsToolName}
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
