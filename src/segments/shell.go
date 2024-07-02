package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Shell struct {
	props properties.Properties
	env   runtime.Environment

	Name    string
	Version string
}

const (
	// MappedShellNames allows for custom text in place of shell names
	MappedShellNames properties.Property = "mapped_shell_names"
)

func (s *Shell) Template() string {
	return NameTemplate
}

func (s *Shell) Enabled() bool {
	mappedNames := s.props.GetKeyValueMap(MappedShellNames, make(map[string]string))
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

func (s *Shell) Init(props properties.Properties, env runtime.Environment) {
	s.props = props
	s.env = env
}
