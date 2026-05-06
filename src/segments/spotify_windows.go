//go:build windows

package segments

import (
	"strings"
)

func (s *Spotify) Enabled() bool {
	// Primary path: query SMTC directly via WinRT, which gives accurate
	// artist/title/status without relying on window titles.
	if output, err := s.env.QuerySMTC(); err == nil {
		artist, title, status, ok := parseSMTCOutput(output)
		if ok {
			s.Artist = artist
			s.Track = title
			s.Status = status
			s.resolveIcon()
			return true
		}
	}

	// Fallback: Spotify PWA running inside Microsoft Edge does not register
	// with SMTC under a "Spotify" app model ID, so we fall back to reading
	// the Edge window title.
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
