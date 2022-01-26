package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
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
		env := new(mock.MockedEnvironment)
		env.On("HasCommand", "dotnet").Return(tc.HasCommand)
		if tc.ExitCode != 0 {
			err := &environment.CommandError{ExitCode: tc.ExitCode}
			env.On("RunCommand", "dotnet", []string{"--version"}).Return("", err)
		} else {
			env.On("RunCommand", "dotnet", []string{"--version"}).Return(tc.Version, nil)
		}

		env.On("HasFiles", "*.cs").Return(true)
		env.On("PathSeperator").Return("")
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})
		props := properties.Map{
			properties.FetchVersion: tc.FetchVersion,
		}
		dotnet := &Dotnet{}
		dotnet.Init(props, env)
		assert.True(t, dotnet.Enabled())
		assert.Equal(t, tc.Expected, renderTemplate(env, dotnet.Template(), dotnet), tc.Case)
	}
}
