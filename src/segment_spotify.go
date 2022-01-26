package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Spotify struct {
	props properties.Properties
	env   environment.Environment

	MusicPlayer
}

type MusicPlayer struct {
	Status string
	Artist string
	Track  string
	Icon   string
}

const (
	// PlayingIcon indicates a song is playing
	PlayingIcon properties.Property = "playing_icon"
	// PausedIcon indicates a song is paused
	PausedIcon properties.Property = "paused_icon"
	// StoppedIcon indicates a song is stopped
	StoppedIcon properties.Property = "stopped_icon"

	playing = "playing"
	stopped = "stopped"
	paused  = "paused"
)

func (s *Spotify) Template() string {
	return "{{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }}"
}

func (s *Spotify) resolveIcon() {
	switch s.Status {
	case stopped:
		// in this case, no artist or track info
		s.Icon = s.props.GetString(StoppedIcon, "\uF04D ")
	case paused:
		s.Icon = s.props.GetString(PausedIcon, "\uF8E3 ")
	case playing:
		s.Icon = s.props.GetString(PlayingIcon, "\uE602 ")
	}
}

func (s *Spotify) Init(props properties.Properties, env environment.Environment) {
	s.props = props
	s.env = env
}
