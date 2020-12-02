package main

import (
	"fmt"
	"strings"
)

type session struct {
	props    *properties
	env      environmentInfo
	userName string
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
	// DisplayDefaultUser hides or shows the user name when it's the user set in DefaultUserName
	DisplayDefaultUser Property = "display_default_user"
)

func (s *session) enabled() bool {
	s.userName = s.getUserName()
	showDefaultUser := s.props.getBool(DisplayDefaultUser, true)
	defaultUser := s.props.getString(DefaultUserName, "")
	if !showDefaultUser && defaultUser == s.userName {
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
	username := s.userName
	computername := s.getComputerName()
	separator := ""
	if s.props.getBool(DisplayHost, true) && s.props.getBool(DisplayUser, true) {
		separator = s.props.getString(UserInfoSeparator, "@")
	}
	var ssh string
	if s.activeSSHSession() {
		ssh = s.props.getString(SSHIcon, "\uF817 ")
	}
	userColor := s.props.getColor(UserColor, s.props.foreground)
	hostColor := s.props.getColor(HostColor, s.props.foreground)
	return fmt.Sprintf("%s<%s>%s</>%s<%s>%s</>", ssh, userColor, username, separator, hostColor, computername)
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
