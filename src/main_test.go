package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleBackgroundColorTemplate(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Term     string
	}{
		{Case: "Inside vscode", Expected: "#123456", Term: "vscode"},
		{Case: "Outside vscode", Expected: "", Term: "windowsterminal"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: map[string]string{
				"TERM_PROGRAM": tc.Term,
			},
		})
		color := getConsoleBackgroundColor(env, "{{ if eq \"vscode\" .Env.TERM_PROGRAM }}#123456{{end}}")
		assert.Equal(t, tc.Expected, color, tc.Case)
	}
}

// This can only be tested here due to circular dependencies
// Which might be an indaction of a fault architecture but
// I honestly could not figure out how to do this better
// without making the solution worse.
func TestDirMatchesOneOf(t *testing.T) {
	cases := []struct {
		GOOS     string
		HomeDir  string
		Dir      string
		Pattern  string
		Expected bool
	}{
		{GOOS: environment.LinuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill", Pattern: "/home/bill", Expected: true},
		{GOOS: environment.LinuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/foo", Expected: true},
		{GOOS: environment.LinuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/Foo", Expected: false},
		{GOOS: environment.LinuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~\\\\foo", Expected: true},
		{GOOS: environment.LinuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo/bar", Pattern: "~/fo.*", Expected: true},
		{GOOS: environment.LinuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/fo\\w", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "C:\\\\Users\\\\Bill", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "C:/Users/Bill", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "c:/users/bill", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "~", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/Foo", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/foo", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo\\Bar", Pattern: "~/fo.*", Expected: true},
		{GOOS: environment.WindowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/fo\\w", Expected: true},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return(tc.GOOS)
		env.On("Home").Return(tc.HomeDir)
		got := environment.DirMatchesOneOf(env, tc.Dir, []string{tc.Pattern})
		assert.Equal(t, tc.Expected, got)
	}
}
