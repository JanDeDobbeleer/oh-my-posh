package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type sessionArgs struct {
	userInfoSeparator string
	username          string
	hostname          string
	goos              string
	connection        string
	client            string
	sshIcon           string
	defaultUserName   string
}

func setupSession(args *sessionArgs) session {
	env := new(MockedEnvironment)
	env.On("getCurrentUser", nil).Return(args.username)
	env.On("getHostName", nil).Return(args.hostname, nil)
	env.On("getRuntimeGOOS", nil).Return(args.goos)
	env.On("getenv", "SSH_CONNECTION").Return(args.connection)
	env.On("getenv", "SSH_CLIENT").Return(args.client)
	env.On("getenv", defaultUserEnvVar).Return(args.defaultUserName)
	props := &properties{
		values: map[Property]interface{}{
			UserInfoSeparator: args.userInfoSeparator,
			SSHIcon:           args.sshIcon,
		},
		foreground: "#fff",
		background: "#000",
	}
	s := session{
		env:   env,
		props: props,
	}
	return s
}

func testUserInfoWriter(args *sessionArgs) string {
	s := setupSession(args)
	_ = s.enabled()
	return s.getFormattedText()
}

func TestWriteUserInfo(t *testing.T) {
	want := "<#fff>bill</>@<#fff>surface</>"
	args := &sessionArgs{
		userInfoSeparator: "@",
		username:          "bill",
		hostname:          "surface",
		goos:              "windows",
	}
	got := testUserInfoWriter(args)
	assert.EqualValues(t, want, got)
}

func TestWriteUserInfoWindowsIncludingHostname(t *testing.T) {
	want := "<#fff>bill</>@<#fff>surface</>"
	args := &sessionArgs{
		userInfoSeparator: "@",
		username:          "surface\\bill",
		hostname:          "surface",
		goos:              "windows",
	}
	got := testUserInfoWriter(args)
	assert.EqualValues(t, want, got)
}

func TestWriteOnlyUsername(t *testing.T) {
	args := &sessionArgs{
		userInfoSeparator: "@",
		username:          "surface\\bill",
		hostname:          "surface",
		goos:              "windows",
	}
	s := setupSession(args)
	s.props.values[DisplayHost] = false
	want := "<#fff>bill</><#fff></>"
	assert.True(t, s.enabled())
	got := s.getFormattedText()
	assert.EqualValues(t, want, got)
}

func TestWriteOnlyHostname(t *testing.T) {
	args := &sessionArgs{
		userInfoSeparator: "@",
		username:          "surface\\bill",
		hostname:          "surface",
		goos:              "windows",
	}
	s := setupSession(args)
	s.props.values[DisplayUser] = false
	want := "<#fff></><#fff>surface</>"
	assert.True(t, s.enabled())
	got := s.getFormattedText()
	assert.EqualValues(t, want, got)
}

func TestWriteActiveSSHSession(t *testing.T) {
	want := "ssh <#fff>bill</>@<#fff>surface</>"
	args := &sessionArgs{
		userInfoSeparator: "@",
		username:          "bill",
		hostname:          "surface",
		goos:              "windows",
		sshIcon:           "ssh ",
		connection:        "1.1.1.1",
	}
	got := testUserInfoWriter(args)
	assert.EqualValues(t, want, got)
}

func TestActiveSSHSessionInactive(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getenv", "SSH_CONNECTION").Return("")
	env.On("getenv", "SSH_CLIENT").Return("")
	s := &session{
		env: env,
	}
	assert.False(t, s.activeSSHSession())
}

func TestActiveSSHSessionActiveConnection(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getenv", "SSH_CONNECTION").Return("1.1.1.1")
	env.On("getenv", "SSH_CLIENT").Return("")
	s := &session{
		env: env,
	}
	assert.True(t, s.activeSSHSession())
}

func TestActiveSSHSessionActiveClient(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getenv", "SSH_CONNECTION").Return("")
	env.On("getenv", "SSH_CLIENT").Return("1.1.1.1")
	s := &session{
		env: env,
	}
	assert.True(t, s.activeSSHSession())
}

func TestActiveSSHSessionActiveBoth(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getenv", "SSH_CONNECTION").Return("2.2.2.2")
	env.On("getenv", "SSH_CLIENT").Return("1.1.1.1")
	s := &session{
		env: env,
	}
	assert.True(t, s.activeSSHSession())
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
			ExpectedString:  "\uf817 <>john</>@<>remote</>",
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
		props := &properties{
			values: map[Property]interface{}{
				SegmentTemplate: tc.Template,
			},
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
