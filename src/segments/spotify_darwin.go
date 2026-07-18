//go:build darwin

package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

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
			set albumName to ""
			set trackNumber to 0
			if playerState is not "stopped" then
				try
					set artistName to artist of current track as string
				end try
				try
					set trackName to name of current track as string
				end try
				try
					set albumName to album of current track as string
				end try
				try
					set trackNumber to track number of current track as integer
				end try
			end if
			return "true|" & playerState & "|" & artistName & "|" & trackName & "|" & albumName & "|" & trackNumber
		end tell
	else
		return "false|||||0"
	end if
	`

	batchedOutput := s.runAppleScriptCommand(batchedCommand)

	outputStrings := strings.SplitN(batchedOutput, "|", 6)
	if len(outputStrings) != 6 || outputStrings[0] == "false" || outputStrings[0] == "" {
		s.Status = stopped
		return false
	}

	if outputStrings[1] == "" {
		s.Status = stopped
		return false
	}

	trackNumber, _ := strconv.Atoi(outputStrings[5])
	return s.applyMediaInfo(&runtime.MediaInfo{
		Status:      outputStrings[1],
		Artist:      outputStrings[2],
		Title:       outputStrings[3],
		Album:       outputStrings[4],
		TrackNumber: trackNumber,
	})
}

func (s *Spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.RunCommand("osascript", "-e", command)
	return val
}
