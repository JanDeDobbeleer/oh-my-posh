package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type MusicPlayer struct {
	Status string
	Artist string
	Track  string
	Icon   template.Markup
}

const (
	PlayingIcon options.Option = "playing_icon"
	PausedIcon  options.Option = "paused_icon"
	StoppedIcon options.Option = "stopped_icon"
	AdIcon      options.Option = "ad_icon"

	playing = "playing"
	stopped = "stopped"
	paused  = "paused"
	ad      = "ad"
)

func (m *MusicPlayer) resolveIcon(opts options.Provider) {
	switch m.Status {
	case stopped:
		// in this case, no artist or track info
		m.Icon = opts.Markup(StoppedIcon, "\uf04d ")
	case paused:
		m.Icon = opts.Markup(PausedIcon, "\uf04c ")
	case playing:
		m.Icon = opts.Markup(PlayingIcon, "\uf04b ")
	case ad:
		m.Icon = opts.Markup(AdIcon, "\ueebb ")
	}
}

// Enable runs the segment's status fetch so every music segment's Enabled()
// is a one-liner with identical error handling.
func (m *MusicPlayer) Enable(setStatus func() error) bool {
	if err := setStatus(); err != nil {
		log.Error(err)
		return false
	}

	return true
}
