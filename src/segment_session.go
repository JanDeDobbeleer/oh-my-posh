package main

import "strings"

type session struct {
	props Properties
	env   Environment
	// text  string

	userName   string
	hostName   string
	SSHSession bool

	// Deprecated
	DefaultUserName string
}

func (s *session) enabled() bool {
	s.SSHSession = s.activeSSHSession()
	segmentTemplate := s.props.getString(SegmentTemplate, "")
	if segmentTemplate == "" {
		return s.legacyEnabled()
	}
	return true
}

func (s *session) string() string {
	segmentTemplate := s.props.getString(SegmentTemplate, "")
	if segmentTemplate == "" {
		return s.legacyString()
	}
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

func (s *session) getUserName() string {
	user := s.env.getCurrentUser()
	username := strings.TrimSpace(user)
	if s.env.getRuntimeGOOS() == "windows" && strings.Contains(username, "\\") {
		username = strings.Split(username, "\\")[1]
	}
	return username
}

func (s *session) getComputerName() string {
	computername, err := s.env.getHostName()
	if err != nil {
		computername = "unknown"
	}
	return strings.TrimSpace(computername)
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
