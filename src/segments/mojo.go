package segments

import (
	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Mojo struct {
	Venv string
	Language
}

func (m *Mojo) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}{{ end }} "
}

func (m *Mojo) Enabled() bool {
	m.loadSpec()

	return m.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (m *Mojo) Activation() Activation {
	m.loadSpec()

	return m.activation()
}

func (m *Mojo) loadSpec() {
	m.extensions = []string{"*.🔥", "*.mojo", "mojoproject.toml"}
	m.tooling = map[string]*cmd{
		mojoToolName: {
			executable:       mojoToolName,
			args:             []string{versionFlagArg},
			regex:            `(?:mojo (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			versionCacheable: true,
		},
	}
	m.defaultTooling = []string{mojoToolName}
	m.displayMode = m.options.String(DisplayMode, DisplayModeEnvironment)
	m.Language.loadContext = m.loadContext
	m.Language.inContext = m.inContext
	// the declared trigger for the pixi/magic virtual env context, keeping
	// the activation gate and loadContext in sync
	m.contextEnvVars = []string{"PIXI_ENVIRONMENT_NAME"}
}

func (m *Mojo) loadContext() {
	if !m.options.Bool(FetchVirtualEnv, true) {
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
	defaultNames := []string{defaultStr}

	if m.options.Bool(options.DisplayDefault, true) ||
		!slices.Contains(defaultNames, name) {
		return true
	}

	return false
}
