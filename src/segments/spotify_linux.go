//go:build linux && !darwin && !windows

package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

func (s *Spotify) Enable() bool {
	command := "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata"
	tlist := s.env.RunShellCommand(shell.BASH, command)

	infos := strings.Split(tlist, " - ")
	s.Status = playing
	s.Artist = infos[0]
	s.Track = infos[1]
	s.resolveIcon()

	return true
}
