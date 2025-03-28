package segments

import (
	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Mojo struct {
	Venv string
	language
}

func (m *Mojo) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}{{ end }} "
}

func (m *Mojo) Enabled() bool {
	m.extensions = []string{"*.🔥", "*.mojo", "mojoproject.toml"}
	m.commands = []*cmd{
		{
			executable: "mojo",
			args:       []string{"--version"},
			regex:      `(?:mojo (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	m.displayMode = m.props.GetString(DisplayMode, DisplayModeEnvironment)
	m.language.loadContext = m.loadContext
	m.language.inContext = m.inContext

	return m.language.Enabled()
}

func (m *Mojo) loadContext() {
	if !m.props.GetBool(FetchVirtualEnv, true) {
		return
	}

	// Magic, the official package manager and virtual env manager,
	// is built on top of pixi: https://github.com/prefix-dev/pixi
	venv := m.env.Getenv("PIXI_ENVIRONMENT_NAME")
	if len(venv) > 0 && m.canUseVenvName(venv) {
		m.Venv = venv
	}
}

func (m *Mojo) inContext() bool {
	return m.Venv != ""
}

func (m *Mojo) canUseVenvName(name string) bool {
	defaultNames := []string{"default"}

	if m.props.GetBool(properties.DisplayDefault, true) ||
		!slices.Contains(defaultNames, name) {
		return true
	}

	return false
}
