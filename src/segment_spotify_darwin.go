//go:build darwin

package main

func (s *spotify) enabled() bool {
	var err error
	// Check if running
	running := s.runAppleScriptCommand("application \"Spotify\" is running")
	if running == "false" || running == "" {
		return false
	}
	s.status = s.runAppleScriptCommand("tell application \"Spotify\" to player state as string")
	if err != nil {
		return false
	}
	if s.status == "stopped" {
		return false
	}
	s.artist = s.runAppleScriptCommand("tell application \"Spotify\" to artist of current track as string")
	s.track = s.runAppleScriptCommand("tell application \"Spotify\" to name of current track as string")
	return true
}

func (s *spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.runCommand("osascript", "-e", command)
	return val
}
