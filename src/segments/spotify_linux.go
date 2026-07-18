//go:build linux && !darwin && !windows

package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

const (
	spotifyDBusCommand = "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify " +
		"/org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player"
	spotifyPlaybackStatusCommand = " string:PlaybackStatus | awk -F '\"' '/string/ {print tolower($2)}'"
	spotifyArtistCommand         = " string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'"
	spotifyTrackCommand          = " string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'"
	spotifyAlbumCommand          = " string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:album'/ {a=$4} END {print a}'"
	spotifyTrackNumberCommand    = " string:Metadata | awk 'BEGIN {RS=\"entry\"}; /'xesam:trackNumber'/ " +
		"{for (i=1; i<=NF; i++) if ($i ~ /^[0-9]+$/) n=$i} END {print n}'"
)

func (s *Spotify) Enabled() bool {
	// Check if we're in WSL and handle that separately
	if s.env.IsWsl() {
		return s.enabledWsl()
	}

	// Standard Linux implementation
	running := s.runLinuxScriptCommand(spotifyPlaybackStatusCommand)

	if strings.HasPrefix(running, "Error") || running == "" {
		return false
	}

	if strings.Contains(running, "Error.ServiceUnknown") || strings.HasSuffix(running, "-") {
		s.Status = stopped
		return false
	}
	if running != playing && running != paused {
		return s.applyMediaInfo(&runtime.MediaInfo{Status: running})
	}

	artist := s.runLinuxScriptCommand(spotifyArtistCommand)
	track := s.runLinuxScriptCommand(spotifyTrackCommand)
	album := s.runLinuxScriptCommand(spotifyAlbumCommand)
	trackNumberOutput := s.runLinuxScriptCommand(spotifyTrackNumberCommand)
	trackNumber, _ := strconv.Atoi(trackNumberOutput)

	return s.applyMediaInfo(&runtime.MediaInfo{
		Status:      running,
		Artist:      artist,
		Title:       track,
		Album:       album,
		TrackNumber: trackNumber,
	})
}

func (s *Spotify) runLinuxScriptCommand(command string) string {
	val := s.env.RunShellCommand(shell.BASH, spotifyDBusCommand+command)
	return val
}

// enabledWsl reads the Windows host's SMTC sessions through the WSL Windows
// interop powershell.exe. The Linux side sees the same data the native
// Windows segment does (playing/paused/stopped/ad).
func (s *Spotify) enabledWsl() bool {
	return s.querySMTC()
}
