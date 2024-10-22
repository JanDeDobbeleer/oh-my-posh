package segments

const UI5ToolingYamlPattern = "*ui5*.y*ml"

type UI5Tooling struct {
	language
	HasUI5YamlInParentDir bool
}

func (u *UI5Tooling) Template() string {
	return languageTemplate
}

func (u *UI5Tooling) Enabled() bool {
	u.extensions = []string{UI5ToolingYamlPattern}
	u.displayMode = u.props.GetString(DisplayMode, DisplayModeContext)
	u.commands = []*cmd{
		{
			executable: "ui5",
			args:       []string{"--version"},
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	u.versionURLTemplate = "https://github.com/SAP/ui5-cli/releases/tag/v{{ .Full }}"
	u.language.loadContext = u.loadContext
	u.language.inContext = u.inContext

	return u.language.Enabled()
}

func (u *UI5Tooling) loadContext() {
	// for searching ui5 yaml from subdirectories of UI5 project root - up to 4 levels
	u.HasUI5YamlInParentDir = u.env.HasFileInParentDirs(UI5ToolingYamlPattern, 4)
}

func (u *UI5Tooling) inContext() bool {
	return u.HasUI5YamlInParentDir
}
