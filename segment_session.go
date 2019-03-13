package main

import (
	"fmt"
	"strings"
)

type session struct {
	props *properties
	env   environmentInfo
}

const (
	//UserInfoSeparator is put between the user and computer name
	UserInfoSeparator Property = "user_info_separator"
	//UserColor if set, is used to color the user name
	UserColor Property = "user_color"
	//ComputerColor if set, is used to color the computer name
	ComputerColor Property = "computer_color"
	//DisplayComputer hides or show the computer name
	DisplayComputer Property = "display_computer"
	//DisplayUser hides or shows the user name
	DisplayUser Property = "display_user"
)

func (s *session) enabled() bool {
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
	username := s.getUserName()
	computername := s.getComputerName()
	return fmt.Sprintf("<%s>%s</>%s<%s>%s</>", s.props.getColor(UserColor, s.props.foreground), username, s.props.getString(UserInfoSeparator, "@"), s.props.getColor(ComputerColor, s.props.foreground), computername)
}

func (s *session) getComputerName() string {
	if !s.props.getBool(DisplayComputer, true) {
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
	user, err := s.env.getCurrentUser()
	if err != nil {
		return "unknown"
	}
	username := strings.TrimSpace(user.Username)
	if s.env.getRuntimeGOOS() == "windows" && strings.Contains(username, "\\") {
		username = strings.Split(username, "\\")[1]
	}
	return username
}
