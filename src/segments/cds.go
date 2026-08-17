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
	c.loadSpec()

	return c.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (c *Cds) Activation() Activation {
	c.loadSpec()

	return c.activation()
}

func (c *Cds) loadSpec() {
	c.extensions = []string{".cdsrc.json", ".cdsrc-private.json", "*.cds"}
	// Not marked versionCacheable: `cds --version` reports the @sap/cds
	// dependency version resolved from the nearest node_modules (the exact
	// line this regex targets), not just the cds-dk CLI's own version - so
	// the same globally resolved binary reports a different version per
	// project.
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
	// the context callback reads package.json in the cwd for a @sap/cds
	// dependency, so its presence gates the context part
	c.contextFiles = []string{fileName}
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
