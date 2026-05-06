//go:build linux && !darwin && !windows

package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

// spotifySMTCScript queries the Windows System Media Transport Controls (SMTC)
// from WSL via PowerShell.  It finds the Spotify session by matching
// SourceAppUserModelId against "Spotify" and emits three lines:
// artist, title, and playback status.
const spotifySMTCScript = `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; ` + //nolint:lll
	`$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType=WindowsRuntime]; ` +
	`$manager = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync().GetAwaiter().GetResult(); ` +
	`$session = $manager.GetSessions() | Where-Object { $_.SourceAppUserModelId -match 'Spotify' } | Select-Object -First 1; ` +
	`if ($null -ne $session) { $props = $session.TryGetMediaPropertiesAsync().GetAwaiter().GetResult(); ` +
	`$playback = $session.GetPlaybackInfo(); ` +
	`$props.Artist; $props.Title; $playback.PlaybackStatus }`

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

func (s *Spotify) enabledWsl() bool {
	output, err := s.env.RunCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", spotifySMTCScript)
	if err != nil {
		s.Status = stopped
		return false
	}

	artist, title, status, ok := parseSMTCOutput(output)
	if !ok {
		s.Status = stopped
		return false
	}

	s.Artist = artist
	s.Track = title
	s.Status = status
	s.resolveIcon()
	return true
}
