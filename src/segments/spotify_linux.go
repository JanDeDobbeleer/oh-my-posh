//go:build linux && !darwin && !windows

package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

func (s *Spotify) Enabled() bool {
	var err error
	running := s.runLinuxScriptCommand("")

	if strings.HasPrefix(running, "ERROR") {
		return false
	}
	if strings.Contains(running, "Error.ServiceUnknown") || strings.HasSuffix(running, "-") {
		s.Status = stopped
		return false
	}
	s.Status = playing
	if err != nil {
		s.Status = stopped
		return false
	}
	if s.Status == stopped {
		return false
	}

	s.Artist = s.runLinuxScriptCommand(" | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'")
	s.Track = s.runLinuxScriptCommand(" | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'")
	s.resolveIcon()
	return true
}

func (s *Spotify) runLinuxScriptCommand(command string) string {
	val := s.env.RunShellCommand(shell.BASH, "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata"+command)
	return val
}
