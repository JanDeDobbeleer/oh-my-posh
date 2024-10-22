package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Session struct {
	base

	DefaultUserName string
	SSHSession      bool
}

func (s *Session) Enabled() bool {
	s.SSHSession = s.activeSSHSession()
	return true
}

func (s *Session) Template() string {
	return " {{ if .SSHSession }}\ueba9 {{ end }}{{ .UserName }}@{{ .HostName }} "
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

	if s.env.Platform() == runtime.WINDOWS {
		return false
	}

	whoAmI, err := s.env.RunCommand("who", "am", "i")
	if err != nil {
		return false
	}

	return regex.MatchString(`\(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\)`, whoAmI)
}
