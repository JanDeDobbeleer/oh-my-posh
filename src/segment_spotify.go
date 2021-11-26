package main

import (
	"fmt"
)

type spotify struct {
	props  properties
	env    environmentInfo
	status string
	artist string
	track  string
}

const (
	// PlayingIcon indicates a song is playing
	PlayingIcon Property = "playing_icon"
	// PausedIcon indicates a song is paused
	PausedIcon Property = "paused_icon"
	// StoppedIcon indicates a song is stopped
	StoppedIcon Property = "stopped_icon"
	// TrackSeparator is put between the artist and the track
	TrackSeparator Property = "track_separator"
)

func (s *spotify) string() string {
	icon := ""
	switch s.status {
	case "stopped":
		// in this case, no artist or track info
		icon = s.props.getString(StoppedIcon, "\uF04D ")
		return icon
	case "paused":
		icon = s.props.getString(PausedIcon, "\uF8E3 ")
	case "playing":
		icon = s.props.getString(PlayingIcon, "\uE602 ")
	}
	separator := s.props.getString(TrackSeparator, " - ")
	return fmt.Sprintf("%s%s%s%s", icon, s.artist, separator, s.track)
}

func (s *spotify) init(props properties, env environmentInfo) {
	s.props = props
	s.env = env
}
