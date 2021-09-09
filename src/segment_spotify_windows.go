//go:build windows

package main

import (
	"strings"
)

func (s *spotify) enabled() bool {
	// search for spotify window to retrieve the title
	// Can be either "Spotify xxx" or the song name "Candlemass - Spellbreaker"
	spotifyWindowTitle, err := s.env.getWindowTitle("spotify.exe", "^(Spotify.*)|(.*\\s-\\s.*)$")
	if err != nil {
		return false
	}

	if !strings.Contains(spotifyWindowTitle, " - ") {
		s.status = "stopped"
		return false
	}

	infos := strings.Split(spotifyWindowTitle, " - ")
	s.artist = infos[0]
	// remove first element and concat others(a song can contains also a " - ")
	s.track = strings.Join(infos[1:], " - ")
	s.status = "playing"
	return true
}
