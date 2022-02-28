package console

import (
	"oh-my-posh/color"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTitle(t *testing.T) {
	cases := []struct {
		Style         Style
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
	}{
		{Style: FolderName, Cwd: "/usr/home", PathSeparator: "/", ShellName: "default", Expected: "\x1b]0;~\a"},
		{Style: FullPath, Cwd: "/usr/home/jan", PathSeparator: "/", ShellName: "default", Expected: "\x1b]0;~/jan\a"},
		{
			Style:         Template,
			Template:      "{{.Env.USERDOMAIN}} :: {{.PWD}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Root:          true,
			Expected:      "\x1b]0;MyCompany :: C:\\vagrant :: Admin :: PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.Folder}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;vagrant :: PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.UserName}}@{{.HostName}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Root:          true,
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@MyHost :: Admin :: PowerShell\a",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return(tc.Cwd)
		env.On("Home").Return("/usr/home")
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: map[string]string{
				"USERDOMAIN": "MyCompany",
			},
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "MyHost",
			PWD:      tc.Cwd,
			Folder:   "vagrant",
		})
		ansi := &color.Ansi{}
		ansi.Init(tc.ShellName)
		ct := &Title{
			Env:      env,
			Ansi:     ansi,
			Style:    tc.Style,
			Template: tc.Template,
		}
		got := ct.GetTitle()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetConsoleTitleIfGethostnameReturnsError(t *testing.T) {
	cases := []struct {
		Style         Style
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
	}{
		{
			Style:         Template,
			Template:      "Not using Host only {{.UserName}} and {{.Shell}}",
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;Not using Host only MyUser and PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.UserName}}@{{.HostName}} :: {{.Shell}}",
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@ :: PowerShell\a",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return(tc.Cwd)
		env.On("Home").Return("/usr/home")
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: map[string]string{
				"USERDOMAIN": "MyCompany",
			},
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "",
		})
		ansi := &color.Ansi{}
		ansi.Init(tc.ShellName)
		ct := &Title{
			Env:      env,
			Ansi:     ansi,
			Style:    tc.Style,
			Template: tc.Template,
		}
		got := ct.GetTitle()
		assert.Equal(t, tc.Expected, got)
	}
}
