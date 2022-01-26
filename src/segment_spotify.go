package main

import "oh-my-posh/environment"

type spotify struct {
	props Properties
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
	PlayingIcon Property = "playing_icon"
	// PausedIcon indicates a song is paused
	PausedIcon Property = "paused_icon"
	// StoppedIcon indicates a song is stopped
	StoppedIcon Property = "stopped_icon"

	playing = "playing"
	stopped = "stopped"
	paused  = "paused"
)

func (s *spotify) template() string {
	return "{{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }}"
}

func (s *spotify) resolveIcon() {
	switch s.Status {
	case stopped:
		// in this case, no artist or track info
		s.Icon = s.props.getString(StoppedIcon, "\uF04D ")
	case paused:
		s.Icon = s.props.getString(PausedIcon, "\uF8E3 ")
	case playing:
		s.Icon = s.props.getString(PlayingIcon, "\uE602 ")
	}
}

func (s *spotify) init(props Properties, env environment.Environment) {
	s.props = props
	s.env = env
}
