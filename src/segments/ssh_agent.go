package segments

import (
	"errors"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

const sshAddNoIdentities = 1

type SSHAgent struct {
	Base

	Count   int
	Running bool
}

func (s *SSHAgent) Template() string {
	return "{{ if .Running }} \uf43d {{ .Count }} {{ end }}"
}

func (s *SSHAgent) Enabled() bool {
	// Windows OpenSSH serves its agent on a named pipe without setting SSH_AUTH_SOCK.
	if s.env.GOOS() != runtime.WINDOWS && s.env.Getenv("SSH_AUTH_SOCK") == "" {
		return true
	}

	output, err := s.env.RunCommand("ssh-add", "-l")
	if err != nil {
		var exitErr interface{ ExitCode() int }
		s.Running = errors.As(err, &exitErr) && exitErr.ExitCode() == sshAddNoIdentities
		return true
	}

	s.Running = true
	if output == "" {
		return true
	}

	s.Count = len(strings.Split(output, "\n"))

	return true
}
