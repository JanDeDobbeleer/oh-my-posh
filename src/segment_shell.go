package main

import "strings"

type shell struct {
	props *properties
	env   environmentInfo
}

const (
	// CustomText allows for custom text in place of shell names
	CustomText Property = "custom_text"
)

func (s *shell) enabled() bool {
	return true
}

func (s *shell) string() string {
	customText := s.props.getKeyValueMap(CustomText, make(map[string]string))
	shellName := s.env.getShellName()
	for key, val := range customText {
		if strings.EqualFold(shellName, key) {
			shellName = val
			break
		}
	}
	return shellName
}

func (s *shell) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}
