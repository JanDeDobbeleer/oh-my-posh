package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

const UI5ToolingYamlPattern = "*ui5*.y*ml"

type UI5Tooling struct {
	language
	HasUI5YamlInParentDir bool
}

func (u *UI5Tooling) Template() string {
	return languageTemplate
}

func (u *UI5Tooling) Init(props properties.Properties, env environment.Environment) {
	u.language = language{
		env:         env,
		props:       props,
		extensions:  []string{UI5ToolingYamlPattern},
		loadContext: u.loadContext,
		inContext:   u.inContext,
		displayMode: props.GetString(DisplayMode, DisplayModeContext),
		commands: []*cmd{
			{
				executable: "ui5",
				args:       []string{"--version"},
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "https://github.com/SAP/ui5-cli/releases/tag/v{{ .Full }}",
	}
}

func (u *UI5Tooling) Enabled() bool {
	return u.language.Enabled()
}

func (u *UI5Tooling) loadContext() {
	// for searching ui5 yaml from subdirectories of UI5 project root - up to 4 levels
	u.HasUI5YamlInParentDir = u.env.HasFileInParentDirs(UI5ToolingYamlPattern, 4)
}

func (u *UI5Tooling) inContext() bool {
	return u.HasUI5YamlInParentDir
}
