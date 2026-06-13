//go:build windows

package segments

import "strings"

func (s *Spotify) Enabled() bool {
	if s.querySMTC() {
		return true
	}

	// Fall back to scraping the Edge window title for the Spotify Web Player PWA.
	windowTitle, err := s.env.QueryWindowTitles("msedge.exe", `^(Spotify.*)`)
	if err != nil {
		return false
	}
	return s.parseWebTitle(windowTitle)
}

func (s *Spotify) parseWebTitle(windowTitle string) bool {
	windowTitle = strings.TrimPrefix(windowTitle, "Spotify - ")
	separator := " • "

	if !strings.Contains(windowTitle, separator) {
		s.Status = stopped
		return false
	}

	before, after, _ := strings.Cut(windowTitle, separator)
	s.Track = before
	s.Artist = after
	s.Status = playing
	s.resolveIcon()
	return true
}
