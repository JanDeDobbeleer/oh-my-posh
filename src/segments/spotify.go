package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

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
		s.Icon = s.options.String(StoppedIcon, "\uf04d ")
	case paused:
		s.Icon = s.options.String(PausedIcon, "\uf04c ")
	case playing:
		s.Icon = s.options.String(PlayingIcon, "\ue602 ")
	}
}

// parseSMTCOutput parses the output from a SMTC query (WinRT or PowerShell).
// The expected format is three newline-separated values: artist, title, and
// playback status ("Playing", "Paused", or any other value treated as stopped).
// Returns ok=false when the output is missing, malformed, or indicates stopped.
func parseSMTCOutput(output string) (artist, title, status string, ok bool) {
	// Normalize CRLF (PowerShell on Windows) to LF.
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")
	output = strings.TrimSpace(output)

	lines := strings.SplitN(output, "\n", 3)
	if len(lines) < 3 {
		return
	}

	artist = strings.TrimSpace(lines[0])
	title = strings.TrimSpace(lines[1])

	switch strings.ToLower(strings.TrimSpace(lines[2])) {
	case "playing":
		status = playing
	case "paused":
		status = paused
	default:
		return
	}

	if artist == "" || title == "" {
		return
	}

	ok = true
	return
}
