package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Python struct {
	Venv string
	language
}

const (
	// FetchVirtualEnv fetches the virtual env
	FetchVirtualEnv      properties.Property = "fetch_virtual_env"
	UsePythonVersionFile properties.Property = "use_python_version_file"
	FolderNameFallback   properties.Property = "folder_name_fallback"
)

func (p *Python) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}{{ end }} "
}

func (p *Python) Init(props properties.Properties, env runtime.Environment) {
	p.language = language{
		env:         env,
		props:       props,
		extensions:  []string{"*.py", "*.ipynb", "pyproject.toml", "venv.bak"},
		folders:     []string{".venv", "venv", "virtualenv", "venv-win", "pyenv-win"},
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
			{
				executable: "py",
				args:       []string{"--version"},
				regex:      `(?:Python (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "https://docs.python.org/release/{{ .Major }}.{{ .Minor }}.{{ .Patch }}/whatsnew/changelog.html#python-{{ .Major }}-{{ .Minor }}-{{ .Patch }}",
		displayMode:        props.GetString(DisplayMode, DisplayModeEnvironment),
	}
}

func (p *Python) Enabled() bool {
	return p.language.Enabled()
}

func (p *Python) loadContext() {
	if !p.language.props.GetBool(FetchVirtualEnv, true) {
		return
	}
	if prompt := p.pyvenvCfgPrompt(); len(prompt) > 0 {
		p.Venv = prompt
		return
	}
	venvVars := []string{
		"VIRTUAL_ENV",
		"CONDA_ENV_PATH",
		"CONDA_DEFAULT_ENV",
	}

	folderNameFallback := p.language.props.GetBool(FolderNameFallback, true)
	defaultVenvNames := []string{
		".venv",
		"venv",
	}

	var venv string
	for _, venvVar := range venvVars {
		venv = p.language.env.Getenv(venvVar)
		if len(venv) == 0 {
			continue
		}

		name := runtime.Base(p.language.env, venv)
		if folderNameFallback && slices.Contains(defaultVenvNames, name) {
			venv = strings.TrimSuffix(venv, name)
			name = runtime.Base(p.language.env, venv)
		}

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

func (p *Python) pyvenvCfgPrompt() string {
	path := p.language.env.CommandPath("python")
	if len(path) == 0 {
		path = p.language.env.CommandPath("python3")
	}
	if len(path) == 0 {
		return ""
	}
	pyvenvDir := filepath.Dir(path)
	if !p.language.env.HasFilesInDir(pyvenvDir, "pyvenv.cfg") {
		pyvenvDir = filepath.Dir(pyvenvDir)
	}
	if !p.language.env.HasFilesInDir(pyvenvDir, "pyvenv.cfg") {
		return ""
	}
	pyvenvCfg := p.env.FileContent(filepath.Join(pyvenvDir, "pyvenv.cfg"))
	for _, line := range strings.Split(pyvenvCfg, "\n") {
		lineSplit := strings.SplitN(line, "=", 2)
		if len(lineSplit) != 2 {
			continue
		}
		key := strings.TrimSpace(lineSplit[0])
		if key == "prompt" {
			value := strings.TrimSpace(lineSplit[1])
			return value
		}
	}
	return ""
}
