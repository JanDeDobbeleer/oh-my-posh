package main

import (
	"fmt"
	"testing"

	"oh-my-posh/runtime"

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
			Template:      "{{.Env.USERDOMAIN}} :: {{.Path}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
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
			Template:      "{{.User}}@{{.Host}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
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
		env := new(runtime.MockedEnvironment)
		env.On("Getcwd", nil).Return(tc.Cwd)
		env.On("HomeDir", nil).Return("/usr/home")
		env.On("GetPathSeperator", nil).Return(tc.PathSeperator)
		env.On("IsRunningAsRoot", nil).Return(tc.Root)
		env.On("GetShellName", nil).Return(tc.ShellName)
		env.On("Getenv", "USERDOMAIN").Return("MyCompany")
		env.On("GetCurrentUser", nil).Return("MyUser")
		env.On("GetHostName", nil).Return("MyHost", nil)
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
			Template:      "Not using Host only {{.User}} and {{.Shell}}",
			User:          "MyUser",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;Not using Host only MyUser and PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.User}}@{{.Host}} :: {{.Shell}}",
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
		env := new(runtime.MockedEnvironment)
		env.On("Getcwd", nil).Return(tc.Cwd)
		env.On("HomeDir", nil).Return("/usr/home")
		env.On("GetPathSeperator", nil).Return(tc.PathSeperator)
		env.On("IsRunningAsRoot", nil).Return(tc.Root)
		env.On("GetShellName", nil).Return(tc.ShellName)
		env.On("Getenv", "USERDOMAIN").Return("MyCompany")
		env.On("GetCurrentUser", nil).Return("MyUser")
		env.On("GetHostName", nil).Return("", fmt.Errorf("I have a bad feeling about this"))
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
