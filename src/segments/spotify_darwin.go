//go:build darwin

package segments

import "strings"

func (s *Spotify) Enabled() bool {
	// Batching commands to reduce latency. Each individual call to `osascript` creates additional delays.
	// Using '|' as a delimiter in the batched command since it's unlikely
	// to appear in track or artist names, making it safe for splitting the output
	batchedCommand := `
	if application "Spotify" is running then
		tell application "Spotify"
			set playerState to player state as string
			set artistName to ""
			set trackName to ""
			if playerState is not "stopped" then
				set artistName to artist of current track as string
				set trackName to name of current track as string
			end if
			return "true|" & playerState & "|" & artistName & "|" & trackName
		end tell
	else
		return "false|||"
	end if
	`

	batchedOutput := s.runAppleScriptCommand(batchedCommand)

	outputStrings := strings.SplitN(batchedOutput, "|", 4)
	if outputStrings[0] == "false" || outputStrings[0] == "" || len(outputStrings) != 4 {
		s.Status = stopped
		return false
	}

	s.Status = outputStrings[1]

	// Check if running
	if s.Status == "" {
		s.Status = stopped
		return false
	}

	if s.Status == stopped {
		return false
	}

	s.Artist = outputStrings[2]
	s.Track = outputStrings[3]
	s.resolveIcon()

	return true
}

func (s *Spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.RunCommand("osascript", "-e", command)
	return val
}
