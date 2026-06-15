//go:build windows || (linux && !darwin)

package segments

import "github.com/jandedobbeleer/oh-my-posh/src/runtime"

// applyMediaInfo maps a media session read from SMTC (native WinRT on Windows
// via runtime.QueryMediaPlayer, or the WSL PowerShell script in spotify_smtc.go)
// onto s and returns whether the segment should be displayed.
func (s *Spotify) applyMediaInfo(info *runtime.MediaInfo) bool {
	if info == nil {
		return false
	}

	switch info.Status {
	case playing:
		s.Status = playing
	case paused:
		s.Status = paused
	default:
		// stopped, closed, opened, changing — segment hidden, matching macOS/Linux behavior.
		s.Status = stopped
		return false
	}

	s.Track = info.Title
	s.Artist = info.Artist

	// Spotify ads expose the campaign as a regular "Music" SMTC entry but with no
	// AlbumTitle and TrackNumber=0 — neither is true for real tracks. Use the pair
	// to flag ads (single condition would false-positive on Spotify Singles etc.).
	if s.Status == playing && info.Album == "" && info.TrackNumber == 0 {
		s.Status = ad
	}

	s.resolveIcon()
	return true
}
