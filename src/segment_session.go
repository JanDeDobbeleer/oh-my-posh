package main

type session struct {
	props Properties
	env   Environment
	// text  string

	SSHSession bool

	// Deprecated
	DefaultUserName string
}

func (s *session) enabled() bool {
	s.SSHSession = s.activeSSHSession()
	return true
}

func (s *session) string() string {
	segmentTemplate := s.props.getString(SegmentTemplate, "{{ .UserName}}@{{ .HostName }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  s,
		Env:      s.env,
	}
	text, err := template.render()
	if err != nil {
		text = err.Error()
	}
	return text
}

func (s *session) init(props Properties, env Environment) {
	s.props = props
	s.env = env
}

func (s *session) activeSSHSession() bool {
	keys := []string{
		"SSH_CONNECTION",
		"SSH_CLIENT",
	}
	for _, key := range keys {
		content := s.env.getenv(key)
		if content != "" {
			return true
		}
	}
	return false
}
