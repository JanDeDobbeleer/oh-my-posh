package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionSegmentTemplate(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		UserName        string
		DefaultUserName string
		ComputerName    string
		SSHSession      bool
		Root            bool
		Template        string
	}{
		{
			Case:            "user and computer",
			ExpectedString:  "john@company-laptop",
			ComputerName:    "company-laptop",
			UserName:        "john",
			Template:        "{{.UserName}}@{{.ComputerName}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "user only",
			ExpectedString:  "john",
			UserName:        "john",
			Template:        "{{.UserName}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "user with ssh",
			ExpectedString:  "john on remote",
			UserName:        "john",
			SSHSession:      true,
			ComputerName:    "remote",
			Template:        "{{.UserName}}{{if .SSHSession}} on {{.ComputerName}}{{end}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "user without ssh",
			ExpectedString:  "john",
			UserName:        "john",
			SSHSession:      false,
			ComputerName:    "remote",
			Template:        "{{.UserName}}{{if .SSHSession}} on {{.ComputerName}}{{end}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "user with root and ssh",
			ExpectedString:  "super john on remote",
			UserName:        "john",
			SSHSession:      true,
			ComputerName:    "remote",
			Root:            true,
			Template:        "{{if .Root}}super {{end}}{{.UserName}}{{if .SSHSession}} on {{.ComputerName}}{{end}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "no template",
			ExpectedString:  "\uf817 john@remote",
			UserName:        "john",
			SSHSession:      true,
			ComputerName:    "remote",
			Root:            true,
			ExpectedEnabled: true,
		},
		{
			Case:            "default user not equal",
			ExpectedString:  "john",
			UserName:        "john",
			DefaultUserName: "jack",
			SSHSession:      true,
			ComputerName:    "remote",
			Root:            true,
			Template:        "{{if ne .Env.POSH_SESSION_DEFAULT_USER .UserName}}{{.UserName}}{{end}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "default user equal",
			ExpectedString:  "",
			UserName:        "john",
			DefaultUserName: "john",
			SSHSession:      true,
			ComputerName:    "remote",
			Root:            true,
			Template:        "{{if ne .Env.POSH_SESSION_DEFAULT_USER .UserName}}{{.UserName}}{{end}}",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getCurrentUser").Return(tc.UserName)
		env.On("getRuntimeGOOS").Return("burp")
		env.On("getHostName").Return(tc.ComputerName, nil)
		var SSHSession string
		if tc.SSHSession {
			SSHSession = "zezzion"
		}
		env.On("getenv", "SSH_CONNECTION").Return(SSHSession)
		env.On("getenv", "SSH_CLIENT").Return(SSHSession)
		env.On("isRunningAsRoot").Return(tc.Root)
		env.On("getenv", defaultUserEnvVar).Return(tc.DefaultUserName)
		env.onTemplate()
		session := &session{
			env: env,
			props: properties{
				SegmentTemplate: tc.Template,
			},
		}
		assert.Equal(t, tc.ExpectedEnabled, session.enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, session.string(), tc.Case)
		}
	}
}
