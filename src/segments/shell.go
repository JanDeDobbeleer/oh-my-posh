package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Shell struct {
	Base

	Name    string
	Version string
}

const (
	// MappedShellNames allows for custom text in place of shell names
	MappedShellNames options.Option = "mapped_shell_names"
)

func (s *Shell) Template() string {
	return NameTemplate
}

func (s *Shell) Enabled() bool {
	mappedNames := s.options.KeyValueMap(MappedShellNames, make(map[string]string))
	s.Name = s.env.Shell()
	s.Version = s.env.Flags().ShellVersion
	for key, val := range mappedNames {
		if strings.EqualFold(s.Name, key) {
			s.Name = val
			break
		}
	}
	return true
}
