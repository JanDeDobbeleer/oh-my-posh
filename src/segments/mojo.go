package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Mojo struct {
	language

	Venv string
}

func (m *Mojo) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}{{ end }} "
}

func (m *Mojo) Init(props properties.Properties, env runtime.Environment) {
	m.language = language{
		env:         env,
		props:       props,
		extensions:  []string{"*.🔥", "*.mojo", "mojoproject.toml"},
		loadContext: m.loadContext,
		inContext:   m.inContext,
		commands: []*cmd{
			{
				executable: "mojo",
				args:       []string{"--version"},
				regex:      `(?:mojo (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		displayMode: props.GetString(DisplayMode, DisplayModeEnvironment),
	}
}

func (m *Mojo) Enabled() bool {
	return m.language.Enabled()
}

func (m *Mojo) loadContext() {
	if !m.language.props.GetBool(FetchVirtualEnv, true) {
		return
	}
	// Magic, the official package manager and virtual env manager,
	// is built on top of pixi: https://github.com/prefix-dev/pixi
	venv := m.language.env.Getenv("PIXI_ENVIRONMENT_NAME")
	if m.canUseVenvName(venv) {
		m.Venv = venv
	}
}

func (m *Mojo) inContext() bool {
	return m.Venv != ""
}

func (m *Mojo) canUseVenvName(name string) bool {
	if m.language.props.GetBool(properties.DisplayDefault, true) || name != "default" {
		return true
	}
	return false
}
