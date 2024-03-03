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

	if strings.HasPrefix(running, "Error") {
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

// Implementation for WSL - moved from spotify_wsl.go
func (s *Spotify) enabledWsl() bool {
	tlist, err := s.env.RunCommand("tasklist.exe", "/V", "/FI", "Imagename eq Spotify.exe", "/FO", "CSV", "/NH")
	if err != nil || strings.HasPrefix(tlist, "INFO") {
		return false
	}

	records := strings.Split(tlist, "\n")
	if len(records) == 0 {
		return false
	}

	for _, record := range records {
		record = strings.TrimSpace(record)
		fields := strings.Split(record, ",")
		if len(fields) == 0 {
			continue
		}

		// last elemant is the title
		title := fields[len(fields)-1]
		// trim leading and trailing quotes from the field
		title = strings.TrimPrefix(title, `"`)
		title = strings.TrimSuffix(title, `"`)
		if !strings.Contains(title, " - ") {
			continue
		}

		infos := strings.Split(title, " - ")
		s.Artist = infos[0]
		s.Track = strings.Join(infos[1:], " - ")
		s.Status = playing
		s.resolveIcon()

		return true
	}

	return false
}
