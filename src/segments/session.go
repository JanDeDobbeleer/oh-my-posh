package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Session struct {
	props properties.Properties
	env   environment.Environment
	// text  string

	SSHSession bool

	// Deprecated
	DefaultUserName string
}

func (s *Session) Enabled() bool {
	s.SSHSession = s.activeSSHSession()
	return true
}

func (s *Session) Template() string {
	return "{{ if .SSHSession }}\uf817 {{ end }}{{ .UserName }}@{{ .HostName }}"
}

func (s *Session) Init(props properties.Properties, env environment.Environment) {
	s.props = props
	s.env = env
}

func (s *Session) activeSSHSession() bool {
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
