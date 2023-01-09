//go:build windows

package segments

import (
	"strings"
)

func (s *Spotify) Enabled() bool {
	// search for spotify window to retrieve the title
	// Can be either "Spotify xxx" or the song name "Candlemass - Spellbreaker"
	windowTitle, err := s.env.QueryWindowTitles("spotify.exe", `^(Spotify.*)|(.*\s-\s.*)$`)
	if err == nil {
		return s.parseNativeTitle(windowTitle)
	}
	windowTitle, err = s.env.QueryWindowTitles("msedge.exe", `^(Spotify.*)`)
	if err != nil {
		return false
	}
	return s.parseWebTitle(windowTitle)
}

func (s *Spotify) parseNativeTitle(windowTitle string) bool {
	separator := " - "

	if !strings.Contains(windowTitle, separator) {
		s.Status = stopped
		return false
	}

	index := strings.Index(windowTitle, separator)
	s.Artist = windowTitle[0:index]
	s.Track = windowTitle[index+len(separator):]
	s.Status = playing
	s.resolveIcon()
	return true
}

func (s *Spotify) parseWebTitle(windowTitle string) bool {
	windowTitle = strings.TrimPrefix(windowTitle, "Spotify - ")
	separator := " • "

	if !strings.Contains(windowTitle, separator) {
		s.Status = stopped
		return false
	}

	index := strings.Index(windowTitle, separator)
	s.Track = windowTitle[0:index]
	s.Artist = windowTitle[index+len(separator):]
	s.Status = playing
	s.resolveIcon()
	return true
}
