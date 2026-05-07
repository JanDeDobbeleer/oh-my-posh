//go:build windows

package segments

import "strings"

func (s *Spotify) Enabled() bool {
	// Primary path: native WinRT call into SMTC. See runtime/smtc_windows.go
	// for the combase.dll binding. PowerShell startup latency made the
	// previous approach unsuitable for per-prompt rendering.
	if output, err := s.env.QuerySpotifySMTC(); err == nil && s.parseSMTCOutput(output) {
		return true
	}

	// Fall back to scraping the Edge window title for the Spotify Web Player PWA,
	// whose SMTC entry surfaces under "Microsoft.MicrosoftEdge_*" rather than
	// "Spotify*", so the SMTC walker above skips it.
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
