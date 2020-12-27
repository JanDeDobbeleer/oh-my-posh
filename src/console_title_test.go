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
		Cwd           string
		PathSeperator string
		ShellName     string
		Expected      string
	}{
		{Style: FolderName, Cwd: "/usr/home", PathSeperator: "/", ShellName: "default", Expected: "\x1b]0;home\a"},
		{Style: FullPath, Cwd: "/usr/home/jan", PathSeperator: "/", ShellName: "default", Expected: "\x1b]0;/usr/home/jan\a"},
		{
			Style:         Template,
			Template:      "{{.Path}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Root:          true,
			Expected:      "\x1b]0;C:\\vagrant :: Admin :: PowerShell\a",
		},
		{
			Style:         Template,
			Template:      "{{.Folder}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeperator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;vagrant :: PowerShell\a",
		},
	}

	for _, tc := range cases {
		settings := &Settings{
			ConsoleTitleStyle:    tc.Style,
			ConsoleTitleTemplate: tc.Template,
		}
		env := new(MockedEnvironment)
		env.On("getcwd", nil).Return(tc.Cwd)
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getPathSeperator", nil).Return(tc.PathSeperator)
		env.On("isRunningAsRoot", nil).Return(tc.Root)
		env.On("getShellName", nil).Return(tc.ShellName)
		formats := &ansiFormats{}
		formats.init(tc.ShellName)
		ct := &consoleTitle{
			env:      env,
			settings: settings,
			formats:  formats,
		}
		got := ct.getConsoleTitle()
		assert.Equal(t, tc.Expected, got)
	}
}
