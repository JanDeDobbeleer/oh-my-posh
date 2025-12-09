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
	dbusCMD := "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.options.Get string:org.mpris.MediaPlayer2.Player"
	val := s.env.RunShellCommand(shell.BASH, dbusCMD+command)
	return val
}

func (s *Spotify) enabledWsl() bool {
	psCommand := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; (Get-Process Spotify -ErrorAction SilentlyContinue | Where-Object {$_.MainWindowTitle -ne ""} | Select-Object -First 1).MainWindowTitle` //nolint: lll

	windowName, err := s.env.RunCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", psCommand)
	if err != nil {
		s.Status = stopped
		return false
	}

	title := strings.TrimSpace(windowName)
	if title == "" || !strings.Contains(title, " - ") {
		s.Status = stopped
		return false
	}

	infos := strings.SplitN(title, " - ", 2)
	if len(infos) < 2 {
		s.Status = stopped
		return false
	}

	s.Artist = strings.TrimSpace(infos[0])
	s.Track = strings.TrimSpace(infos[1])

	if s.Artist == "" || s.Track == "" {
		s.Status = stopped
		return false
	}

	s.Status = playing
	s.resolveIcon()

	return true
}
