package main

import (
	"fmt"
	"strings"
)

type session struct {
	props           *properties
	env             environmentInfo
	UserName        string
	DefaultUserName string
	ComputerName    string
	SSHSession      bool
	Root            bool
	templateText    string
}

const (
	// UserInfoSeparator is put between the user and computer name
	UserInfoSeparator Property = "user_info_separator"
	// UserColor if set, is used to color the user name
	UserColor Property = "user_color"
	// HostColor if set, is used to color the computer name
	HostColor Property = "host_color"
	// DisplayHost hides or show the computer name
	DisplayHost Property = "display_host"
	// DisplayUser hides or shows the user name
	DisplayUser Property = "display_user"
	// SSHIcon shows when in an SSH session
	SSHIcon Property = "ssh_icon"
	// DefaultUserName holds the default user of the platform
	DefaultUserName Property = "default_user_name"

	defaultUserEnvVar = "POSH_SESSION_DEFAULT_USER"
)

func (s *session) enabled() bool {
	s.UserName = s.getUserName()
	s.ComputerName = s.getComputerName()
	s.SSHSession = s.activeSSHSession()
	s.DefaultUserName = s.getDefaultUser()
	segmentTemplate := s.props.getString(SegmentTemplate, "")
	if segmentTemplate != "" {
		s.Root = s.env.isRunningAsRoot()
		template := &textTemplate{
			Template: segmentTemplate,
			Context:  s,
			Env:      s.env,
		}
		var err error
		s.templateText, err = template.render()
		if err != nil {
			s.templateText = err.Error()
		}
		return len(s.templateText) > 0
	}
	showDefaultUser := s.props.getBool(DisplayDefault, true)
	if !showDefaultUser && s.DefaultUserName == s.UserName {
		return false
	}
	return true
}

func (s *session) string() string {
	return s.getFormattedText()
}

func (s *session) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}

func (s *session) getFormattedText() string {
	if len(s.templateText) > 0 {
		return s.templateText
	}
	separator := ""
	if s.props.getBool(DisplayHost, true) && s.props.getBool(DisplayUser, true) {
		separator = s.props.getString(UserInfoSeparator, "@")
	}
	var sshIcon string
	if s.SSHSession {
		sshIcon = s.props.getString(SSHIcon, "\uF817 ")
	}
	userColor := s.props.getColor(UserColor, s.props.foreground)
	hostColor := s.props.getColor(HostColor, s.props.foreground)
	if len(userColor) > 0 && len(hostColor) > 0 {
		return fmt.Sprintf("%s<%s>%s</>%s<%s>%s</>", sshIcon, userColor, s.UserName, separator, hostColor, s.ComputerName)
	}
	if len(userColor) > 0 {
		return fmt.Sprintf("%s<%s>%s</>%s%s", sshIcon, userColor, s.UserName, separator, s.ComputerName)
	}
	if len(hostColor) > 0 {
		return fmt.Sprintf("%s%s%s<%s>%s</>", sshIcon, s.UserName, separator, hostColor, s.ComputerName)
	}
	return fmt.Sprintf("%s%s%s%s", sshIcon, s.UserName, separator, s.ComputerName)
}

func (s *session) getComputerName() string {
	if !s.props.getBool(DisplayHost, true) {
		return ""
	}
	computername, err := s.env.getHostName()
	if err != nil {
		computername = "unknown"
	}
	return strings.TrimSpace(computername)
}

func (s *session) getUserName() string {
	if !s.props.getBool(DisplayUser, true) {
		return ""
	}
	user := s.env.getCurrentUser()
	username := strings.TrimSpace(user)
	if s.env.getRuntimeGOOS() == "windows" && strings.Contains(username, "\\") {
		username = strings.Split(username, "\\")[1]
	}
	return username
}

func (s *session) getDefaultUser() string {
	user := s.env.getenv(defaultUserEnvVar)
	if len(user) == 0 {
		user = s.props.getString(DefaultUserName, "")
	}
	return user
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
