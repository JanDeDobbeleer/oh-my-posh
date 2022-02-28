package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
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
			ExpectedString: "",
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
		env := new(mock.MockedEnvironment)
		env.On("User").Return(tc.UserName)
		env.On("GOOS").Return("burp")
		env.On("Host").Return(tc.ComputerName, nil)
		var SSHSession string
		if tc.SSHSession {
			SSHSession = "zezzion"
		}
		env.On("Getenv", "SSH_CONNECTION").Return(SSHSession)
		env.On("Getenv", "SSH_CLIENT").Return(SSHSession)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			UserName: tc.UserName,
			HostName: tc.ComputerName,
			Env: map[string]string{
				"SSH_CONNECTION":            SSHSession,
				"SSH_CLIENT":                SSHSession,
				"POSH_SESSION_DEFAULT_USER": tc.DefaultUserName,
			},
			Root: tc.Root,
		})
		session := &Session{
			env:   env,
			props: properties.Map{},
		}
		_ = session.Enabled()
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, session), tc.Case)
	}
}
