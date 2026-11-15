package segments

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

const (
	SSHAgentKeySeparator options.Option = "key_separator"
	SSHAgentKeyTemplate  options.Option = "key_template"
	SSHAgentMappedKeys   options.Option = "mapped_keys"

	sshAddNoIdentities = 1
)

type SSHAgent struct {
	Base

	Keys          template.Markup
	Count         int
	UnmappedCount int
	Running       bool
}

type SSHKey struct {
	Alias       template.Markup
	Type        string
	Comment     string
	Fingerprint string
	Bits        int
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

	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	s.Count = len(lines)
	s.Keys = s.renderKeys(lines)

	return true
}

func (s *SSHAgent) renderKeys(lines []string) template.Markup {
	aliases := s.options.KeyValueMap(SSHAgentMappedKeys, nil)
	separator := template.RawMarkup(s.options.String(SSHAgentKeySeparator, " "))
	keyTemplate := "{{ or .Alias .Comment }}"

	parts := make([]template.Markup, 0, 2*len(lines))
	s.UnmappedCount = 0
	renderedKeys := 0
	for _, line := range lines {
		key := parseSSHKey(line)
		alias, mapped := aliases[key.Fingerprint]
		if !mapped {
			s.UnmappedCount++
		}
		key.Alias = template.RawMarkup(alias)

		rendered := template.RawMarkup(s.options.Template(SSHAgentKeyTemplate, keyTemplate, key))
		if rendered == "" {
			continue
		}
		if renderedKeys > 0 {
			parts = append(parts, separator)
		}

		parts = append(parts, rendered)
		renderedKeys++
	}

	return template.JoinMarkup(parts...)
}

// parseSSHKey reads a `ssh-add -l` line: "<bits> <fingerprint> <comment> (<type>)",
// where the comment may contain spaces.
func parseSSHKey(line string) *SSHKey {
	key := &SSHKey{}

	bits, rest, _ := strings.Cut(line, " ")
	key.Bits, _ = strconv.Atoi(bits)
	key.Fingerprint, rest, _ = strings.Cut(rest, " ")

	index := strings.LastIndex(rest, " (")
	if index == -1 || !strings.HasSuffix(rest, ")") {
		key.Comment = rest
		return key
	}

	key.Comment = rest[:index]
	key.Type = rest[index+2 : len(rest)-1]

	return key
}
