package main

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
	return p.Executable()
}

func (s *shell) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}
