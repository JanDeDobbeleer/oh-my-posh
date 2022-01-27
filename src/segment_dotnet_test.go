package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotnetSegment(t *testing.T) {
	cases := []struct {
		Case         string
		Expected     string
		ExitCode     int
		HasCommand   bool
		Version      string
		FetchVersion bool
	}{
		{Case: "Unsupported version", Expected: "\uf071", HasCommand: true, FetchVersion: true, ExitCode: dotnetExitCode, Version: "3.1.402"},
		{Case: "Regular version", Expected: "3.1.402", HasCommand: true, FetchVersion: true, Version: "3.1.402"},
		{Case: "Regular version", Expected: "", HasCommand: true, FetchVersion: false, Version: "3.1.402"},
		{Case: "Regular version", Expected: "", HasCommand: false, FetchVersion: false, Version: "3.1.402"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("HasCommand", "dotnet").Return(tc.HasCommand)
		if tc.ExitCode != 0 {
			err := &commandError{exitCode: tc.ExitCode}
			env.On("RunCommand", "dotnet", []string{"--version"}).Return("", err)
		} else {
			env.On("RunCommand", "dotnet", []string{"--version"}).Return(tc.Version, nil)
		}

		env.On("HasFiles", "*.cs").Return(true)
		env.On("PathSeperator").Return("")
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("TemplateCache").Return(&TemplateCache{
			Env: make(map[string]string),
		})
		props := properties{
			FetchVersion: tc.FetchVersion,
		}
		dotnet := &dotnet{}
		dotnet.init(props, env)
		assert.True(t, dotnet.enabled())
		assert.Equal(t, tc.Expected, renderTemplate(env, dotnet.template(), dotnet), tc.Case)
	}
}
