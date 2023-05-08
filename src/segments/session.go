package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Session struct {
	props properties.Properties
	env   platform.Environment
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
	return " {{ if .SSHSession }}\U000f0318 {{ end }}{{ .UserName }}@{{ .HostName }} "
}

func (s *Session) Init(props properties.Properties, env platform.Environment) {
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
