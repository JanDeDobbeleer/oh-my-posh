package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestPythonTemplate(t *testing.T) {
	cases := []struct {
		Case             string
		Expected         string
		ExpectedDisabled bool
		Template         string
		VirtualEnvName   string
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
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasCommand", "python").Return(true)
		env.On("runCommand", "python", []string{"--version"}).Return("Python 3.8.4", nil)
		env.On("hasFiles", "*.py").Return(true)
		env.On("getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("getenv", "CONDA_ENV_PATH").Return(tc.VirtualEnvName)
		env.On("getenv", "CONDA_DEFAULT_ENV").Return(tc.VirtualEnvName)
		env.On("getenv", "PYENV_VERSION").Return(tc.VirtualEnvName)
		env.On("getPathSeperator", nil).Return("")
		env.On("getcwd", nil).Return("/usr/home/project")
		env.On("homeDir", nil).Return("/usr/home")
		var props properties = map[Property]interface{}{
			FetchVersion:    tc.FetchVersion,
			SegmentTemplate: tc.Template,
			DisplayMode:     DisplayModeAlways,
		}
		python := &python{}
		python.init(props, env)
		assert.Equal(t, !tc.ExpectedDisabled, python.enabled(), tc.Case)
		assert.Equal(t, tc.Expected, python.string(), tc.Case)
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
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return("")
		env.On("getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("getenv", "CONDA_ENV_PATH").Return("")
		env.On("getenv", "CONDA_DEFAULT_ENV").Return("")
		env.On("getenv", "PYENV_VERSION").Return("")
		python := &python{}
		python.init(nil, env)
		python.loadContext()
		assert.Equal(t, tc.Expected, python.inContext())
	}
}
