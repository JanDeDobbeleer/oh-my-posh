package main

type shell struct {
	props *properties
	env   environmentInfo
}

func (s *shell) enabled() bool {
	return true
}

func (s *shell) string() string {
	return s.env.getShellName()
}

func (s *shell) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}
