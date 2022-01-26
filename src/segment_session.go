package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type session struct {
	props properties.Properties
	env   environment.Environment
	// text  string

	SSHSession bool

	// Deprecated
	DefaultUserName string
}

func (s *session) enabled() bool {
	s.SSHSession = s.activeSSHSession()
	return true
}

func (s *session) template() string {
	return "{{ .UserName }}@{{ .HostName }}"
}

func (s *session) init(props properties.Properties, env environment.Environment) {
	s.props = props
	s.env = env
}

func (s *session) activeSSHSession() bool {
	keys := []string{
		"SSH_CONNECTION",
		"SSH_CLIENT",
	}
	for _, key := range keys {
		content := s.env.Getenv(key)
		if content != "" {
			return true
		}
	}
	return false
}
