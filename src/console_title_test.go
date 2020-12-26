package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConsoleTitle(t *testing.T) {
	cases := []struct {
		Style         ConsoleTitleStyle
		Cwd           string
		PathSeperator string
		ShellName     string
		Expected      string
	}{
		{Style: FolderName, Cwd: "/usr/home", PathSeperator: "/", ShellName: "default", Expected: "\x1b]0;home\a"},
		{Style: FullPath, Cwd: "/usr/home/jan", PathSeperator: "/", ShellName: "default", Expected: "\x1b]0;/usr/home/jan\a"},
	}

	for _, tc := range cases {
		settings := &Settings{
			ConsoleTitleStyle: tc.Style,
		}
		env := new(MockedEnvironment)
		env.On("getcwd", nil).Return(tc.Cwd)
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getPathSeperator", nil).Return(tc.PathSeperator)
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
