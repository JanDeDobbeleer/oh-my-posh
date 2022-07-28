package segments

import (
	"errors"
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"path/filepath"
	"strings"
)

type Python struct {
	language

	Venv string
}

const (
	// FetchVirtualEnv fetches the virtual env
	FetchVirtualEnv      properties.Property = "fetch_virtual_env"
	UsePythonVersionFile properties.Property = "use_python_version_file"
)

func (p *Python) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}{{ end }} "
}

func (p *Python) Init(props properties.Properties, env environment.Environment) {
	p.language = language{
		env:         env,
		props:       props,
		extensions:  []string{"*.py", "*.ipynb", "pyproject.toml", "venv.bak"},
		folders:     []string{".venv", "venv", "virtualenv", "env", "venv-win", "pyenv-win"},
		loadContext: p.loadContext,
		inContext:   p.inContext,
		commands: []*cmd{
			{
				getVersion: p.pyenvVersion,
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
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
		versionURLTemplate: "https://docs.python.org/release/{{ .Major }}.{{ .Minor }}.{{ .Patch }}/whatsnew/changelog.html#python-{{ .Major }}-{{ .Minor }}-{{ .Patch }}",
		displayMode:        props.GetString(DisplayMode, DisplayModeEnvironment),
		homeEnabled:        props.GetBool(HomeEnabled, true),
	}
}

func (p *Python) Enabled() bool {
	return p.language.Enabled()
}

func (p *Python) loadContext() {
	if !p.language.props.GetBool(FetchVirtualEnv, true) {
		return
	}
	venvVars := []string{
		"VIRTUAL_ENV",
		"CONDA_ENV_PATH",
		"CONDA_DEFAULT_ENV",
	}
	var venv string
	for _, venvVar := range venvVars {
		venv = p.language.env.Getenv(venvVar)
		name := environment.Base(p.language.env, venv)
		if p.canUseVenvName(name) {
			p.Venv = name
			break
		}
	}
}

func (p *Python) inContext() bool {
	return p.Venv != ""
}

func (p *Python) canUseVenvName(name string) bool {
	if name == "" || name == "." {
		return false
	}
	if p.language.props.GetBool(properties.DisplayDefault, true) {
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

func (p *Python) pyenvVersion() (string, error) {
	// Use `pyenv root` instead of $PYENV_ROOT?
	// Is our Python executable at $PYENV_ROOT/bin/python ?
	// Should p.env expose command paths?
	path := p.env.CommandPath("python")
	if len(path) == 0 {
		path = p.env.CommandPath("python3")
	}
	if len(path) == 0 {
		return "", errors.New("no python executable found")
	}
	pyEnvRoot := p.env.Getenv("PYENV_ROOT")
	// TODO:  pyenv-win has this at $PYENV_ROOT/pyenv-win/shims
	if path != filepath.Join(pyEnvRoot, "shims", "python") {
		return "", fmt.Errorf("executable at %s is not a pyenv shim", path)
	}
	// pyenv version-name will return current version or virtualenv
	cmdOutput, err := p.env.RunCommand("pyenv", "version-name")
	if err != nil {
		return "", err
	}
	versionString := strings.Split(cmdOutput, ":")[0]
	if len(versionString) == 0 {
		return "", errors.New("no pyenv version-name found")
	}

	// $PYENV_ROOT/versions + versionString (symlinks resolved) == $PYENV_ROOT/versions/(version)[/envs/(virtualenv)]
	realPath, err := p.env.ResolveSymlink(filepath.Join(pyEnvRoot, "versions", versionString))
	if err != nil {
		return "", err
	}
	// ../versions/(version)[/envs/(virtualenv)]
	shortPath, err := filepath.Rel(filepath.Join(pyEnvRoot, "versions"), realPath)
	if err != nil {
		return "", err
	}
	// override virtualenv if pyenv set one
	parts := strings.Split(shortPath, string(filepath.Separator))
	if len(parts) > 2 && p.canUseVenvName(parts[2]) {
		p.Venv = parts[2]
	}
	return parts[0], nil
}
