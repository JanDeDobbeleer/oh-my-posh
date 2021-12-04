package main

import "fmt"

type python struct {
	language

	Venv string
}

const (
	// DisplayVirtualEnv shows or hides the virtual env
	DisplayVirtualEnv Property = "display_virtual_env"
)

func (p *python) string() string {
	if p.Venv == "" {
		return p.language.string()
	}
	version := p.language.string()
	if version == "" {
		return p.Venv
	}
	return fmt.Sprintf("%s %s", p.Venv, version)
}

func (p *python) init(props properties, env environmentInfo) {
	p.language = language{
		env:         env,
		props:       props,
		extensions:  []string{"*.py", "*.ipynb", "pyproject.toml", "venv.bak", "venv", ".venv"},
		loadContext: p.loadContext,
		inContext:   p.inContext,
		commands: []*cmd{
			{
				executable: "python",
				args:       []string{"--version"},
				regex:      `(?:Python (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
			{
				executable: "python3",
				args:       []string{"--version"},
				regex:      `(?:Python (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "[%s](https://www.python.org/downloads/release/python-%s%s%s/)",
		displayMode:        props.getString(DisplayMode, DisplayModeEnvironment),
		homeEnabled:        props.getBool(HomeEnabled, true),
	}
}

func (p *python) enabled() bool {
	return p.language.enabled()
}

func (p *python) loadContext() {
	if !p.language.props.getBool(DisplayVirtualEnv, true) {
		return
	}
	venvVars := []string{
		"VIRTUAL_ENV",
		"CONDA_ENV_PATH",
		"CONDA_DEFAULT_ENV",
		"PYENV_VERSION",
	}
	var venv string
	for _, venvVar := range venvVars {
		venv = p.language.env.getenv(venvVar)
		name := base(venv, p.language.env)
		if p.canUseVenvName(name) {
			p.Venv = name
			break
		}
	}
}

func (p *python) inContext() bool {
	return p.Venv != ""
}

func (p *python) canUseVenvName(name string) bool {
	if name == "" || name == "." {
		return false
	}
	if p.language.props.getBool(DisplayDefault, true) {
		return true
	}
	invalidNames := [2]string{"system", "base"}
	for _, a := range invalidNames {
		if a == name {
			return false
		}
	}
	return true
}
