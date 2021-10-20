package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirMatchesOneOf(t *testing.T) {
	cases := []struct {
		GOOS     string
		HomeDir  string
		Dir      string
		Pattern  string
		Expected bool
	}{
		{GOOS: linuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill", Pattern: "/home/bill", Expected: true},
		{GOOS: linuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/foo", Expected: true},
		{GOOS: linuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/Foo", Expected: false},
		{GOOS: linuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~\\\\foo", Expected: true},
		{GOOS: linuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo/bar", Pattern: "~/fo.*", Expected: true},
		{GOOS: linuxPlatform, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/fo\\w", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "C:\\\\Users\\\\Bill", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "C:/Users/Bill", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "c:/users/bill", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "~", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/Foo", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/foo", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo\\Bar", Pattern: "~/fo.*", Expected: true},
		{GOOS: windowsPlatform, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/fo\\w", Expected: true},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getRuntimeGOOS", nil).Return(tc.GOOS)
		env.On("homeDir", nil).Return(tc.HomeDir)
		got := dirMatchesOneOf(env, tc.Dir, []string{tc.Pattern})
		assert.Equal(t, tc.Expected, got)
	}
}
