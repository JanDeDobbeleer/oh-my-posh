package main

import "oh-my-posh/runtime"

type shell struct {
	props *properties
	env   runtime.Environment
}

func (s *shell) enabled() bool {
	return true
}

func (s *shell) string() string {
	return s.env.GetShellName()
}

func (s *shell) init(props *properties, env runtime.Environment) {
	s.props = props
	s.env = env
}
