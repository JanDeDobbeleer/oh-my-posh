package main

import (
	"oh-my-posh/environment"
	"strings"
)

type shell struct {
	props Properties
	env   environment.Environment

	Name string
}

const (
	// MappedShellNames allows for custom text in place of shell names
	MappedShellNames Property = "mapped_shell_names"
)

func (s *shell) template() string {
	return "{{ .Name }}"
}

func (s *shell) enabled() bool {
	mappedNames := s.props.GetKeyValueMap(MappedShellNames, make(map[string]string))
	s.Name = s.env.Shell()
	for key, val := range mappedNames {
		if strings.EqualFold(s.Name, key) {
			s.Name = val
			break
		}
	}
	return true
}

func (s *shell) init(props Properties, env environment.Environment) {
	s.props = props
	s.env = env
}
