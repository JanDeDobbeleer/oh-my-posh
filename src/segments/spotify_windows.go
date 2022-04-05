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
		return s.parseSpotifyTitle(windowTitle, " - ")
	}
	windowTitle, err = s.env.QueryWindowTitles("msedge.exe", `^(Spotify.*)`)
	if err != nil {
		return false
	}
	return s.parseWebSpotifyTitle(windowTitle)
}

func (s *Spotify) parseWebSpotifyTitle(windowTitle string) bool {
	windowTitle = strings.TrimPrefix(windowTitle, "Spotify - ")
	return s.parseSpotifyTitle(windowTitle, " • ")
}

func (s *Spotify) parseSpotifyTitle(windowTitle, separator string) bool {
	if !strings.Contains(windowTitle, separator) {
		s.Status = stopped
		return false
	}

	infos := strings.Split(windowTitle, separator)
	s.Artist = infos[0]
	// remove first element and concat others(a song can contains also a " - ")
	s.Track = strings.Join(infos[1:], separator)
	s.Status = playing
	s.resolveIcon()
	return true
}
