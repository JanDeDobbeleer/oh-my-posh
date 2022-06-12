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
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
	}{
		{
			Template:      "{{.Env.USERDOMAIN}} :: {{.PWD}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Root:          true,
			Expected:      "\x1b]0;MyCompany :: C:\\vagrant :: Admin :: PowerShell\a",
		},
		{
			Template:      "{{.Folder}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;vagrant :: PowerShell\a",
		},
		{
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
		ansi.InitPlain()
		ct := &Title{
			Env:      env,
			Ansi:     ansi,
			Template: tc.Template,
		}
		got := ct.GetTitle()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetConsoleTitleIfGethostnameReturnsError(t *testing.T) {
	cases := []struct {
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
	}{
		{
			Template:      "Not using Host only {{.UserName}} and {{.Shell}}",
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;Not using Host only MyUser and PowerShell\a",
		},
		{
			Template:      "{{.UserName}}@{{.HostName}} :: {{.Shell}}",
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@ :: PowerShell\a",
		},
		{
			Template: "\x1b[93m[\x1b[39m\x1b[96mconsole-title\x1b[39m\x1b[96m ≡\x1b[39m\x1b[31m +0\x1b[39m\x1b[31m ~1\x1b[39m\x1b[31m -0\x1b[39m\x1b[31m !\x1b[39m\x1b[93m]\x1b[39m",
			Expected: "\x1b]0;[console-title ≡ +0 ~1 -0 !]\a",
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
		ansi.InitPlain()
		ct := &Title{
			Env:      env,
			Ansi:     ansi,
			Template: tc.Template,
		}
		got := ct.GetTitle()
		assert.Equal(t, tc.Expected, got)
	}
}
