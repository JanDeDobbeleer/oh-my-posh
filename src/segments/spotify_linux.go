//go:build linux && !darwin && !windows

package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

func (s *Spotify) Enabled() bool {
	// Check if we're in WSL and handle that separately
	if s.env.IsWsl() {
		return s.enabledWsl()
	}

	// Standard Linux implementation
	running := s.runLinuxScriptCommand(" string:PlaybackStatus | awk -F '\"' '/string/ {print tolower($2)}'")

	if strings.HasPrefix(running, "Error") || running == "" {
		return false
	}

	if strings.Contains(running, "Error.ServiceUnknown") || strings.HasSuffix(running, "-") {
		s.Status = stopped
		return false
	}

	if running == stopped {
		s.Status = stopped
		return false
	}

	s.Status = running
	s.Artist = s.runLinuxScriptCommand(" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'")
	s.Track = s.runLinuxScriptCommand(" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'")
	s.resolveIcon()

	return true
}

func (s *Spotify) runLinuxScriptCommand(command string) string {
	dbusCMD := "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player"
	val := s.env.RunShellCommand(shell.BASH, dbusCMD+command)
	return val
}

// enabledWsl reads the Windows host's SMTC sessions through the WSL Windows
// interop powershell.exe. The Linux side sees the same data the native
// Windows segment does (playing/paused/stopped/ad).
func (s *Spotify) enabledWsl() bool {
	return s.querySMTC()
}
