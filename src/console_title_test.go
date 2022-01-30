package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConsoleTitle(t *testing.T) {
	cases := []struct {
		Style         ConsoleTitleStyle
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeperator string
		ShellName     string
		Expected      string
	}{
		{Style: FolderName, Cwd: "/usr/home", PathSeperator: "/", ShellName: "default", Expected: "\x1b]0;~\a"},
		{Style: FullPath, Cwd: "/usr/home/jan", PathSeperator: "/", ShellName: "default", Expected: "\x1b]0;~/jan\a"},
		{
			Style:         Template,
			Template:      "{{.Env.USERDOMAIN}} :: {{.PWD}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Root:          true,
			Expected:      "\x1b]0;MyCompany :: C:\\vagrant :: Admin :: PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.Folder}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;vagrant :: PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.UserName}}@{{.HostName}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Root:          true,
			User:          "MyUser",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@MyHost :: Admin :: PowerShell\a",
		},
	}

	for _, tc := range cases {
		config := &Config{
			ConsoleTitleStyle:    tc.Style,
			ConsoleTitleTemplate: tc.Template,
		}
		env := new(MockedEnvironment)
		env.On("pwd").Return(tc.Cwd)
		env.On("homeDir").Return("/usr/home")
		env.On("getPathSeperator").Return(tc.PathSeperator)
		env.On("templateCache").Return(&templateCache{
			Env: map[string]string{
				"USERDOMAIN": "MyCompany",
			},
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "MyHost",
			PWD:      tc.Cwd,
			Folder:   base(tc.Cwd, env),
		})
		env.onTemplate()
		ansi := &ansiUtils{}
		ansi.init(tc.ShellName)
		ct := &consoleTitle{
			env:    env,
			config: config,
			ansi:   ansi,
		}
		got := ct.getConsoleTitle()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetConsoleTitleIfGethostnameReturnsError(t *testing.T) {
	cases := []struct {
		Style         ConsoleTitleStyle
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeperator string
		ShellName     string
		Expected      string
	}{
		{
			Style:         Template,
			Template:      "Not using Host only {{.UserName}} and {{.Shell}}",
			User:          "MyUser",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;Not using Host only MyUser and PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.UserName}}@{{.HostName}} :: {{.Shell}}",
			User:          "MyUser",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@ :: PowerShell\a",
		},
	}

	for _, tc := range cases {
		config := &Config{
			ConsoleTitleStyle:    tc.Style,
			ConsoleTitleTemplate: tc.Template,
		}
		env := new(MockedEnvironment)
		env.On("pwd").Return(tc.Cwd)
		env.On("homeDir").Return("/usr/home")
		env.On("templateCache").Return(&templateCache{
			Env: map[string]string{
				"USERDOMAIN": "MyCompany",
			},
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "",
		})
		env.onTemplate()
		ansi := &ansiUtils{}
		ansi.init(tc.ShellName)
		ct := &consoleTitle{
			env:    env,
			config: config,
			ansi:   ansi,
		}
		got := ct.getConsoleTitle()
		assert.Equal(t, tc.Expected, got)
	}
}
