//go:build windows

package main

import (
	"strings"
)

func (s *Spotify) Enabled() bool {
	// search for spotify window to retrieve the title
	// Can be either "Spotify xxx" or the song name "Candlemass - Spellbreaker"
	spotifyWindowTitle, err := s.env.WindowTitle("spotify.exe", "^(Spotify.*)|(.*\\s-\\s.*)$")
	if err != nil {
		return false
	}

	if !strings.Contains(spotifyWindowTitle, " - ") {
		s.Status = stopped
		return false
	}

	infos := strings.Split(spotifyWindowTitle, " - ")
	s.Artist = infos[0]
	// remove first element and concat others(a song can contains also a " - ")
	s.Track = strings.Join(infos[1:], " - ")
	s.Status = playing
	s.resolveIcon()
	return true
}
