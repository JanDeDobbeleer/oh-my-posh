//go:build darwin

package main

func (s *spotify) enabled() bool {
	var err error
	// Check if running
	running := s.runAppleScriptCommand("application \"Spotify\" is running")
	if running == "false" || running == "" {
		return false
	}
	s.Status = s.runAppleScriptCommand("tell application \"Spotify\" to player state as string")
	if err != nil {
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

func (s *spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.runCommand("osascript", "-e", command)
	return val
}
