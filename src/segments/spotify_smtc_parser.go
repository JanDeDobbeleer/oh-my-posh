//go:build windows || (linux && !darwin)

package segments

import "strings"

// parseSMTCOutput consumes a "<status>|<title>|<artist>|<album>|<trackNumber>"
// line produced by either the WSL PowerShell script (spotify_smtc.go) or the
// native Windows WinRT path (runtime/smtc_windows.go), applies it to s, and
// returns whether the segment should be displayed.
func (s *Spotify) parseSMTCOutput(output string) bool {
	output = strings.TrimSpace(output)
	if output == "" {
		return false
	}

	parts := strings.SplitN(output, "|", 5)
	if len(parts) != 5 {
		return false
	}

	switch parts[0] {
	case playing:
		s.Status = playing
	case paused:
		s.Status = paused
	default:
		// stopped, closed, opened, changing — segment hidden, matching macOS/Linux behavior.
		s.Status = stopped
		return false
	}

	s.Track = parts[1]
	s.Artist = parts[2]

	// Spotify ads expose the campaign as a regular "Music" SMTC entry but with no
	// AlbumTitle and TrackNumber=0 — neither is true for real tracks. Use the pair
	// to flag ads (single condition would false-positive on Spotify Singles etc.).
	if s.Status == playing && parts[3] == "" && parts[4] == "0" {
		s.Status = ad
	}

	s.resolveIcon()
	return true
}
