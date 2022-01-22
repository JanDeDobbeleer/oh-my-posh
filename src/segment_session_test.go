package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionSegmentTemplate(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		UserName        string
		DefaultUserName string
		ComputerName    string
		SSHSession      bool
		Root            bool
		Template        string
	}{
		{
			Case:           "user and computer",
			ExpectedString: "john@company-laptop",
			ComputerName:   "company-laptop",
			UserName:       "john",
			Template:       "{{.UserName}}@{{.HostName}}",
		},
		{
			Case:           "user only",
			ExpectedString: "john",
			UserName:       "john",
			Template:       "{{.UserName}}",
		},
		{
			Case:           "user with ssh",
			ExpectedString: "john on remote",
			UserName:       "john",
			SSHSession:     true,
			ComputerName:   "remote",
			Template:       "{{.UserName}}{{if .SSHSession}} on {{.HostName}}{{end}}",
		},
		{
			Case:           "user without ssh",
			ExpectedString: "john",
			UserName:       "john",
			SSHSession:     false,
			ComputerName:   "remote",
			Template:       "{{.UserName}}{{if .SSHSession}} on {{.HostName}}{{end}}",
		},
		{
			Case:           "user with root and ssh",
			ExpectedString: "super john on remote",
			UserName:       "john",
			SSHSession:     true,
			ComputerName:   "remote",
			Root:           true,
			Template:       "{{if .Root}}super {{end}}{{.UserName}}{{if .SSHSession}} on {{.HostName}}{{end}}",
		},
		{
			Case:           "no template",
			ExpectedString: "\uf817 john@remote",
			UserName:       "john",
			SSHSession:     true,
			ComputerName:   "remote",
			Root:           true,
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
		env.On("getenv", defaultUserEnvVar).Return(tc.DefaultUserName)
		env.On("templateCache").Return(&templateCache{
			UserName: tc.UserName,
			HostName: tc.ComputerName,
			Env: map[string]string{
				"SSH_CONNECTION":  SSHSession,
				"SSH_CLIENT":      SSHSession,
				defaultUserEnvVar: tc.DefaultUserName,
			},
			Root: tc.Root,
		})
		session := &session{
			env: env,
			props: properties{
				SegmentTemplate: tc.Template,
			},
		}
		_ = session.enabled()
		assert.Equal(t, tc.ExpectedString, session.string(), tc.Case)
	}
}
