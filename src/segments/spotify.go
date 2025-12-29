package segments

import "github.com/jandedobbeleer/oh-my-posh/src/segments/options"

type Spotify struct {
	Base

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
	PlayingIcon options.Option = "playing_icon"
	// PausedIcon indicates a song is paused
	PausedIcon options.Option = "paused_icon"
	// StoppedIcon indicates a song is stopped
	StoppedIcon options.Option = "stopped_icon"
	// AdIcon indicates an advertisement is playing
	AdIcon options.Option = "ad_icon"

	playing = "playing"
	stopped = "stopped"
	paused  = "paused"
)

func (s *Spotify) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }} "
}

func (s *Spotify) resolveIcon() {
	switch s.Status {
	case stopped:
		// in this case, no artist or track info
		s.Icon = s.options.String(StoppedIcon, "\uF04D ")
	case paused:
		s.Icon = s.options.String(PausedIcon, "\uF8E3 ")
	case playing:
		s.Icon = s.options.String(PlayingIcon, "\uE602 ")
	}
}
