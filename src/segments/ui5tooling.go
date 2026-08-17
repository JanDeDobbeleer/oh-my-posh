package segments

const UI5ToolingYamlPattern = "*ui5*.y*ml"

type UI5Tooling struct {
	Language
	HasUI5YamlInParentDir bool
}

func (u *UI5Tooling) Template() string {
	return languageTemplate
}

func (u *UI5Tooling) Enabled() bool {
	u.loadSpec()

	return u.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation. The
// default context mode resolves to Always: the context callback is a
// depth-limited glob search through parent directories, which no condition
// kind expresses (ProjectFiles match exact names only).
func (u *UI5Tooling) Activation() Activation {
	u.loadSpec()

	return u.activation()
}

func (u *UI5Tooling) loadSpec() {
	const ui5ToolName = "ui5"

	u.extensions = []string{UI5ToolingYamlPattern}
	u.displayMode = u.options.String(DisplayMode, DisplayModeContext)
	u.tooling = map[string]*cmd{
		ui5ToolName: {
			executable:       ui5ToolName,
			args:             []string{versionFlagArg},
			regex:            versionRegexPrefixed,
			versionCacheable: true,
		},
	}
	u.defaultTooling = []string{ui5ToolName}
	u.versionURLTemplate = "https://github.com/SAP/ui5-cli/releases/tag/v{{ .Full }}"
	u.Language.loadContext = u.loadContext
	u.Language.inContext = u.inContext
}

func (u *UI5Tooling) loadContext() {
	// for searching ui5 yaml from subdirectories of UI5 project root - up to 4 levels
	u.HasUI5YamlInParentDir = u.env.HasFileInParentDirs(UI5ToolingYamlPattern, 4)
}

func (u *UI5Tooling) inContext() bool {
	return u.HasUI5YamlInParentDir
}
