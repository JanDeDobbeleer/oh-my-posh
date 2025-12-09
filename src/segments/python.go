package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Python struct {
	Venv string
	Language
}

const (
	// FetchVirtualEnv fetches the virtual env
	FetchVirtualEnv      options.Option = "fetch_virtual_env"
	UsePythonVersionFile options.Option = "use_python_version_file"
	FolderNameFallback   options.Option = "folder_name_fallback"
	DefaultVenvNames     options.Option = "default_venv_names"
)

func (p *Python) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}{{ end }} "
}

func (p *Python) Enabled() bool {
	p.extensions = []string{"*.py", "*.ipynb", "pyproject.toml", "venv.bak"}
	p.folders = []string{".venv", "venv", "virtualenv", "venv-win", "pyenv-win"}
	p.commands = []*cmd{
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
	}
	p.versionURLTemplate = "https://docs.python.org/release/{{ .Major }}.{{ .Minor }}.{{ .Patch }}/whatsnew/changelog.html#python-{{ .Major }}-{{ .Minor }}-{{ .Patch }}"
	p.displayMode = p.options.String(DisplayMode, DisplayModeEnvironment)
	p.Language.loadContext = p.loadContext
	p.Language.inContext = p.inContext

	return p.Language.Enabled()
}

func (p *Python) loadContext() {
	if !p.options.Bool(FetchVirtualEnv, true) {
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

	folderNameFallback := p.options.Bool(FolderNameFallback, true)
	defaultVenvNames := p.options.StringArray(DefaultVenvNames, []string{
		".venv",
		"venv",
	})

	var venv string
	for _, venvVar := range venvVars {
		venv = p.env.Getenv(venvVar)
		if venv == "" {
			continue
		}

		name := path.Base(venv)
		log.Debugf("virtual env name: %s", name)
		if folderNameFallback && slices.Contains(defaultVenvNames, name) {
			venv = strings.TrimSuffix(venv, name)
			name = path.Base(venv)
			log.Debugf("virtual env name (fallback): %s", name)
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
	if p.options.Bool(options.DisplayDefault, true) {
		return true
	}

	invalidNames := [2]string{"system", "base"}
	for _, a := range invalidNames {
		if a == name {
			log.Debugf("virtual env name %s is invalid", name)
			return false
		}
	}

	return true
}

func (p *Python) pyenvVersion() (string, error) {
	// Use `pyenv root` instead of $PYENV_ROOT?
	// Is our Python executable at $PYENV_ROOT/bin/python ?
	// Should p.env expose command paths?
	cmdPath := p.env.CommandPath("python")
	if cmdPath == "" {
		cmdPath = p.env.CommandPath("python3")
	}

	if cmdPath == "" {
		return "", errors.New("no python executable found")
	}

	pyEnvRoot := p.env.Getenv("PYENV_ROOT")
	if pyEnvRoot == "" || !strings.HasPrefix(cmdPath, pyEnvRoot) {
		return "", fmt.Errorf("executable at %s is not a pyenv shim", cmdPath)
	}

	// pyenv version-name will return current version or virtualenv
	cmdOutput, err := p.env.RunCommand("pyenv", "version-name")
	if err != nil {
		return "", err
	}

	versionString := strings.Split(cmdOutput, ":")[0]
	if versionString == "" {
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
	cmdPath := p.env.CommandPath("python")
	if cmdPath == "" {
		cmdPath = p.env.CommandPath("python3")
	}

	if cmdPath == "" {
		return ""
	}

	pyvenvDir := filepath.Dir(cmdPath)
	if !p.env.HasFilesInDir(pyvenvDir, "pyvenv.cfg") {
		pyvenvDir = filepath.Dir(pyvenvDir)
	}

	if !p.env.HasFilesInDir(pyvenvDir, "pyvenv.cfg") {
		return ""
	}

	pyvenvCfg := p.env.FileContent(filepath.Join(pyvenvDir, "pyvenv.cfg"))
	for line := range strings.SplitSeq(pyvenvCfg, "\n") {
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
