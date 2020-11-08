// +build windows

package main

import (
	"strings"
)

const (
	//AutoHotkeyScript script to read spotify windows title
	AutoHotkeyScript Property = "autohotkey_script"
)

func (s *spotify) enabled() bool {
	if !s.env.hasCommand("AutoHotkey") {
		return false
	}
	spotifyWindowTitle, err := s.env.runCommand("AutoHotkey", s.props.getString(AutoHotkeyScript, ""))
	if err != nil || spotifyWindowTitle == "" {
		return false
	}
	if !strings.Contains(spotifyWindowTitle, " - ") {
		s.status = "stopped"
		return false
	}
	infos := strings.Split(spotifyWindowTitle, " - ")
	s.artist = infos[0]
	// remove first element
	s.track = strings.Join(infos[1:], " - ")
	s.status = "playing"
	return true
}
