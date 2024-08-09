//go:build darwin

package segments

func (s *Spotify) Enabled() bool {
	// Check if running
	running := s.runAppleScriptCommand("application \"Spotify\" is running")
	if running == "false" || running == "" {
		s.Status = stopped
		return false
	}

	s.Status = s.runAppleScriptCommand("tell application \"Spotify\" to player state as string")

	if len(s.Status) == 0 {
		s.Status = stopped
		return false
	}

	if s.Status == stopped {
		return false
	}

	s.Artist = s.runAppleScriptCommand("tell application \"Spotify\" to artist of current track as string")
	s.Track = s.runAppleScriptCommand("tell application \"Spotify\" to name of current track as string")

	s.resolveIcon()

	return true
}

func (s *Spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.RunCommand("osascript", "-e", command)
	return val
}
