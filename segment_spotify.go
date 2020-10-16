package main

import (
	"fmt"
)

type spotify struct {
	props  *properties
	env    environmentInfo
	status string
	artist string
	track  string
}

const (
	//PlayingIcon indicates a song is playing
	PlayingIcon Property = "playing_icon"
	//PausedIcon indicates a song is paused
	PausedIcon Property = "paused_icon"
	//TrackSeparator is put between the artist and the track
	TrackSeparator Property = "track_separator"
)

func (s *spotify) enabled() bool {
	if s.env.getRuntimeGOOS() != "darwin" {
		return false
	}
	var err error
	// Check if running
	running := s.runAppleScriptCommand("application \"Spotify\" is running")
	if running == "false" || running == "" {
		return false
	}
	s.status = s.runAppleScriptCommand("tell application \"Spotify\" to player state as string")
	if err != nil {
		return false
	}
	if s.status == "stopped" {
		return false
	}
	s.artist = s.runAppleScriptCommand("tell application \"Spotify\" to artist of current track as string")
	s.track = s.runAppleScriptCommand("tell application \"Spotify\" to name of current track as string")
	return true
}

func (s *spotify) string() string {
	icon := ""
	switch s.status {
	case "paused":
		icon = s.props.getString(PausedIcon, "\uF8E3 ")
	case "playing":
		icon = s.props.getString(PlayingIcon, "\uE602 ")
	}
	separator := s.props.getString(TrackSeparator, " - ")
	return fmt.Sprintf("%s%s%s%s", icon, s.artist, separator, s.track)
}

func (s *spotify) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}

func (s *spotify) runAppleScriptCommand(command string) string {
	return s.env.runCommand("osascript", "-e", command)
}
