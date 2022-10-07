package segments

import (
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/alecthomas/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestPythonTemplate(t *testing.T) {
	type ResolveSymlink struct {
		Path string
		Err  error
	}
	cases := []struct {
		Case             string
		Expected         string
		ExpectedDisabled bool
		Template         string
		VirtualEnvName   string
		FetchVersion     bool
		PythonPath       string
		ResolveSymlink   ResolveSymlink
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
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return("")
		env.On("HasCommand", "python").Return(true)
		env.On("CommandPath", mock2.Anything).Return(tc.PythonPath)
		env.On("RunCommand", "python", []string{"--version"}).Return("Python 3.8.4", nil)
		env.On("RunCommand", "pyenv", []string{"version-name"}).Return(tc.VirtualEnvName, nil)
		env.On("HasFiles", "*.py").Return(true)
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "PYENV_ROOT").Return("/home/user/.pyenv")
		env.On("PathSeparator").Return("")
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("ResolveSymlink", mock2.Anything).Return(tc.ResolveSymlink.Path, tc.ResolveSymlink.Err)
		props := properties.Map{
			properties.FetchVersion: tc.FetchVersion,
			UsePythonVersionFile:    true,
			DisplayMode:             DisplayModeAlways,
		}
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})
		python := &Python{}
		python.Init(props, env)
		assert.Equal(t, !tc.ExpectedDisabled, python.Enabled(), tc.Case)
		assert.Equal(t, tc.Expected, renderTemplate(env, tc.Template, python), tc.Case)
	}
}

func TestPythonPythonInContext(t *testing.T) {
	cases := []struct {
		Expected       bool
		VirtualEnvName string
	}{
		{Expected: true, VirtualEnvName: "VENV"},
		{Expected: false, VirtualEnvName: ""},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return("")
		env.On("PathSeparator").Return("/")
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return("")
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return("")
		env.On("Getenv", "PYENV_VERSION").Return("")
		env.On("HasParentFilePath", ".python-version").Return(&environment.FileInfo{}, errors.New("no match at root level"))
		python := &Python{}
		python.Init(properties.Map{}, env)
		python.loadContext()
		assert.Equal(t, tc.Expected, python.inContext())
	}
}
