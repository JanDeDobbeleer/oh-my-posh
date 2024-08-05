package segments

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/alecthomas/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestPythonTemplate(t *testing.T) {
	type ResolveSymlink struct {
		Err  error
		Path string
	}
	cases := []struct {
		ResolveSymlink   ResolveSymlink
		Case             string
		Expected         string
		Template         string
		VirtualEnvName   string
		PythonPath       string
		PyvenvCfg        string
		ExpectedDisabled bool
		FetchVersion     bool
	}{
		{Case: "No virtual env present", FetchVersion: true, Expected: "3.8.4", Template: "{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}"},
		{Case: "Virtual env present", FetchVersion: true, Expected: "VENV 3.8.4", VirtualEnvName: "VENV", Template: "{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Full }}"},
		{
			Case:           "Virtual env major and minor dot",
			FetchVersion:   true,
			Expected:       "VENV 3.8",
			VirtualEnvName: "VENV",
			Template:       "{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Major }}.{{ .Minor }}",
		},
		{
			Case:           "Virtual env hide on default",
			FetchVersion:   true,
			Expected:       "3.8",
			VirtualEnvName: "default",
			Template:       "{{ if ne .Venv \"default\" }}{{ .Venv }} {{ end }}{{ .Major }}.{{ .Minor }}",
		},
		{
			Case:           "Virtual env show non default",
			FetchVersion:   true,
			Expected:       "billy 3.8",
			VirtualEnvName: "billy",
			Template:       "{{ if ne .Venv \"default\" }}{{ .Venv }} {{ end }}{{ .Major }}.{{ .Minor }}",
		},
		{
			Case:           "Pyenv show env",
			FetchVersion:   true,
			Expected:       "VENV 3.8",
			PythonPath:     "/home/user/.pyenv/shims/python",
			VirtualEnvName: "VENV",
			Template:       "{{ if ne .Venv \"default\" }}{{ .Venv }} {{ end }}{{ .Major }}.{{ .Minor }}",
			ResolveSymlink: ResolveSymlink{Path: "/home/user/.pyenv/versions/3.8.8/envs/VENV", Err: nil},
		},
		{
			Case:           "Pyenv no venv",
			FetchVersion:   true,
			Expected:       "3.8",
			PythonPath:     "/home/user/.pyenv/shims/python",
			Template:       "{{ if ne .Venv \"default\" }}{{ .Venv }} {{ end }}{{ .Major }}.{{ .Minor }}",
			ResolveSymlink: ResolveSymlink{Path: "/home/user.pyenv/versions/3.8.8", Err: nil},
		},
		{
			Case:           "Pyenv virtual env version name",
			FetchVersion:   true,
			VirtualEnvName: "demo",
			Expected:       "demo 3.8.4",
			PythonPath:     "/home/user/.pyenv/shims/python",
			Template:       "{{ .Venv }} {{ .Full }}",
			ResolveSymlink: ResolveSymlink{Path: "/home/user/.pyenv/versions/demo", Err: nil},
		},
		{
			Case:           "pyvenv.cfg prompt",
			FetchVersion:   true,
			VirtualEnvName: "VENV",
			PythonPath:     "/home/user/.pyenv/shims/python",
			Template:       "{{ if .Venv }}{{ .Venv }} {{ end }}{{ .Major }}.{{ .Minor }}",
			PyvenvCfg:      "home = /usr/bin/\nprompt = pyvenvCfgPrompt\n",
			Expected:       "pyvenvCfgPrompt 3.8",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "python",
			versionParam:  "--version",
			versionOutput: "Python 3.8.4",
			extension:     "*.py",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("GOOS").Return("")
		env.On("CommandPath", testify_.Anything).Return(tc.PythonPath)
		env.On("RunCommand", "pyenv", []string{"version-name"}).Return(tc.VirtualEnvName, nil)
		env.On("HasFilesInDir", testify_.Anything, "pyvenv.cfg").Return(len(tc.PyvenvCfg) > 0)
		env.On("FileContent", filepath.Join(filepath.Dir(tc.PythonPath), "pyvenv.cfg")).Return(tc.PyvenvCfg)
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "PYENV_ROOT").Return("/home/user/.pyenv")
		env.On("PathSeparator").Return("")
		env.On("ResolveSymlink", testify_.Anything).Return(tc.ResolveSymlink.Path, tc.ResolveSymlink.Err)

		props[properties.FetchVersion] = tc.FetchVersion
		props[UsePythonVersionFile] = true
		props[DisplayMode] = DisplayModeAlways

		python := &Python{}
		python.Init(props, env)
		assert.Equal(t, !tc.ExpectedDisabled, python.Enabled(), tc.Case)
		assert.Equal(t, tc.Expected, renderTemplate(env, tc.Template, python), tc.Case)
	}
}

func TestPythonPythonInContext(t *testing.T) {
	cases := []struct {
		VirtualEnvName string
		Expected       bool
	}{
		{Expected: true, VirtualEnvName: "VENV"},
		{Expected: false, VirtualEnvName: ""},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return("")
		env.On("PathSeparator").Return("/")
		env.On("CommandPath", testify_.Anything).Return("")
		env.On("HasFilesInDir", testify_.Anything, "pyvenv.cfg").Return(false)
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return("")
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return("")
		env.On("Getenv", "PYENV_VERSION").Return("")
		env.On("HasParentFilePath", ".python-version", false).Return(&runtime.FileInfo{}, errors.New("no match at root level"))
		python := &Python{}
		python.Init(properties.Map{}, env)
		python.loadContext()
		assert.Equal(t, tc.Expected, python.inContext())
	}
}

func TestPythonVirtualEnvIgnoreDefaultVenvNames(t *testing.T) {
	cases := []struct {
		Expected           string
		VirtualEnvName     string
		FolderNameFallback bool
	}{
		{
			Expected:           "folder",
			FolderNameFallback: true,
			VirtualEnvName:     "/path/to/folder/.venv",
		},
		{
			Expected:           "folder",
			FolderNameFallback: true,
			VirtualEnvName:     "/path/to/folder/venv",
		},
		{
			Expected:           ".venv",
			FolderNameFallback: false,
			VirtualEnvName:     "/path/to/folder/.venv",
		},
		{
			Expected:           "venv",
			FolderNameFallback: false,
			VirtualEnvName:     "/path/to/folder/venv",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{}
		env, props := getMockedLanguageEnv(params)

		env.On("GOOS").Return("")
		env.On("PathSeparator").Return("/")
		env.On("CommandPath", testify_.Anything).Return("")
		env.On("HasFilesInDir", testify_.Anything, "pyvenv.cfg").Return(false)
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return("")
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return("")
		env.On("Getenv", "PYENV_VERSION").Return("")
		env.On("HasParentFilePath", ".python-version", false).Return(&runtime.FileInfo{}, errors.New("no match at root level"))

		props[FolderNameFallback] = tc.FolderNameFallback

		python := &Python{}
		python.Init(props, env)
		python.loadContext()
		assert.Equal(t, tc.Expected, python.Venv)
	}
}
