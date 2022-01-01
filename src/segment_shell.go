package main

import "strings"

type shell struct {
	props Properties
	env   environmentInfo
}

const (
	// MappedShellNames allows for custom text in place of shell names
	MappedShellNames Property = "mapped_shell_names"
)

func (s *shell) enabled() bool {
	return true
}

func (s *shell) string() string {
	mappedNames := s.props.getKeyValueMap(MappedShellNames, make(map[string]string))
	shellName := s.env.getShellName()
	for key, val := range mappedNames {
		if strings.EqualFold(shellName, key) {
			shellName = val
			break
		}
	}
	return shellName
}

func (s *shell) init(props Properties, env environmentInfo) {
	s.props = props
	s.env = env
}
