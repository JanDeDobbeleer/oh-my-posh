package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropertySessionSegment(t *testing.T) {
	cases := []struct {
		Case               string
		ExpectedEnabled    bool
		ExpectedString     string
		UserName           string
		Host               string
		DefaultUserName    string
		DefaultUserNameEnv string
		SSHSession         bool
		SSHClient          bool
		Root               bool
		DisplayUser        bool
		DisplayHost        bool
		DisplayDefault     bool
		HostColor          string
		UserColor          string
		GOOS               string
		HostError          bool
	}{
		{
			Case:            "user and computer",
			ExpectedString:  "john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			ExpectedEnabled: true,
		},
		{
			Case:            "user and computer with host color",
			ExpectedString:  "john at <yellow>company-laptop</>",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			HostColor:       "yellow",
			ExpectedEnabled: true,
		},
		{
			Case:            "user and computer with user color",
			ExpectedString:  "<yellow>john</> at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			UserColor:       "yellow",
			ExpectedEnabled: true,
		},
		{
			Case:            "user and computer with both colors",
			ExpectedString:  "<yellow>john</> at <green>company-laptop</>",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			UserColor:       "yellow",
			HostColor:       "green",
			ExpectedEnabled: true,
		},
		{
			Case:            "SSH Session",
			ExpectedString:  "ssh john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			SSHSession:      true,
			ExpectedEnabled: true,
		},
		{
			Case:            "SSH Client",
			ExpectedString:  "ssh john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			SSHClient:       true,
			ExpectedEnabled: true,
		},
		{
			Case:            "SSH Client",
			ExpectedString:  "ssh john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			SSHClient:       true,
			ExpectedEnabled: true,
		},
		{
			Case:            "only user name",
			ExpectedString:  "john",
			Host:            "company-laptop",
			UserName:        "john",
			DisplayUser:     true,
			ExpectedEnabled: true,
		},
		{
			Case:            "windows user name",
			ExpectedString:  "john at company-laptop",
			Host:            "company-laptop",
			UserName:        "surface\\john",
			DisplayHost:     true,
			DisplayUser:     true,
			ExpectedEnabled: true,
			GOOS:            string(Windows),
		},
		{
			Case:            "only host name",
			ExpectedString:  "company-laptop",
			Host:            "company-laptop",
			UserName:        "john",
			DisplayDefault:  true,
			DisplayHost:     true,
			ExpectedEnabled: true,
		},
		{
			Case:            "display default - hidden",
			Host:            "company-laptop",
			UserName:        "john",
			DefaultUserName: "john",
			DisplayDefault:  false,
			DisplayHost:     true,
			DisplayUser:     true,
			ExpectedEnabled: false,
		},
		{
			Case:               "display default with env var - hidden",
			Host:               "company-laptop",
			UserName:           "john",
			DefaultUserNameEnv: "john",
			DefaultUserName:    "jake",
			DisplayDefault:     false,
			DisplayHost:        true,
			DisplayUser:        true,
			ExpectedEnabled:    false,
		},
		{
			Case:            "host error",
			ExpectedString:  "john at unknown",
			Host:            "company-laptop",
			HostError:       true,
			UserName:        "john",
			DisplayHost:     true,
			DisplayUser:     true,
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getCurrentUser", nil).Return(tc.UserName)
		env.On("getRuntimeGOOS", nil).Return(tc.GOOS)
		if tc.HostError {
			env.On("getHostName", nil).Return(tc.Host, errors.New("oh snap"))
		} else {
			env.On("getHostName", nil).Return(tc.Host, nil)
		}
		var SSHSession string
		if tc.SSHSession {
			SSHSession = "zezzion"
		}
		var SSHClient string
		if tc.SSHClient {
			SSHClient = "clientz"
		}
		env.On("getenv", "SSH_CONNECTION").Return(SSHSession)
		env.On("getenv", "SSH_CLIENT").Return(SSHClient)
		env.On("getenv", "SSH_CLIENT").Return(SSHSession)
		env.On("getenv", defaultUserEnvVar).Return(tc.DefaultUserNameEnv)
		env.On("isRunningAsRoot", nil).Return(tc.Root)
		var props properties = map[Property]interface{}{
			UserInfoSeparator: " at ",
			SSHIcon:           "ssh ",
			DefaultUserName:   tc.DefaultUserName,
			DisplayDefault:    tc.DisplayDefault,
			DisplayUser:       tc.DisplayUser,
			DisplayHost:       tc.DisplayHost,
			HostColor:         tc.HostColor,
			UserColor:         tc.UserColor,
		}
		session := &session{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.ExpectedEnabled, session.enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, session.string(), tc.Case)
		}
	}
}

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
			Template:        "{{.UserName}}{{if .ComputerName}}@{{.ComputerName}}{{end}}",
			ExpectedEnabled: true,
		},
		{
			Case:            "user only",
			ExpectedString:  "john",
			UserName:        "john",
			Template:        "{{.UserName}}{{if .ComputerName}}@{{.ComputerName}}{{end}}",
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
			Template:        "{{if ne .DefaultUserName .UserName}}{{.UserName}}{{end}}",
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
			Template:        "{{if ne .DefaultUserName .UserName}}{{.UserName}}{{end}}",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getCurrentUser", nil).Return(tc.UserName)
		env.On("getRuntimeGOOS", nil).Return("burp")
		env.On("getHostName", nil).Return(tc.ComputerName, nil)
		var SSHSession string
		if tc.SSHSession {
			SSHSession = "zezzion"
		}
		env.On("getenv", "SSH_CONNECTION").Return(SSHSession)
		env.On("getenv", "SSH_CLIENT").Return(SSHSession)
		env.On("isRunningAsRoot", nil).Return(tc.Root)
		env.On("getenv", defaultUserEnvVar).Return(tc.DefaultUserName)
		session := &session{
			env: env,
			props: map[Property]interface{}{
				SegmentTemplate: tc.Template,
			},
		}
		assert.Equal(t, tc.ExpectedEnabled, session.enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, session.string(), tc.Case)
		}
	}
}
