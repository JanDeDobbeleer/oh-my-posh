package main

import "strings"

type shell struct {
	props *properties
	env   environmentInfo
}

func (s *shell) enabled() bool {
	return true
}

func (s *shell) string() string {
	p, err := s.env.getParentProcess()

	if err != nil {
		return "unknown"
	}
	shell := strings.Replace(p.Executable(), ".exe", "", 1)
	return shell
}

func (s *shell) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}
