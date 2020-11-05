package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type spotify struct {
	props  *properties
	env    environmentInfo
	status string
	artist string
	track  string
}

const (
	//PlayingIcon indicates a song is playing
	PlayingIcon Property = "playing_icon"
	//PausedIcon indicates a song is paused
	PausedIcon Property = "paused_icon"
	//StoppedIcon indicates a song is stopped
	StoppedIcon Property = "stopped_icon"
	//TrackSeparator is put between the artist and the track
	TrackSeparator Property = "track_separator"
	//AutohotkeyTitleScript script to read spotify windows title
	AutohotkeyTitleScript string = "segment_spotify.ahk"
)

func (s *spotify) enabled() bool {
	if s.env.getRuntimeGOOS() == "darwin" {
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
	if s.env.getRuntimeGOOS() == "windows" {
		var err error

		// get the current executable path
		// https://stackoverflow.com/questions/18537257/how-to-get-the-directory-of-the-currently-running-file
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		scriptPath := filepath.Join(exPath, AutohotkeyTitleScript)
		// read windows title using autohotkey
		spotifyWindowTitle := s.runAutoHotkeyScript(scriptPath)
		if err != nil {
			return false
		}

		if !strings.Contains(spotifyWindowTitle, " - ") {
			s.status = "stopped"
			return true
		}

		infos := strings.Split(spotifyWindowTitle, " - ")
		s.artist = infos[0]
		s.track = infos[1]
		s.status = "playing"
		return true
	}
	return false
}

func (s *spotify) string() string {
	icon := ""
	switch s.status {
	case "stopped":
		icon = s.props.getString(StoppedIcon, "\uF04D ")
	case "paused":
		icon = s.props.getString(PausedIcon, "\uF8E3 ")
	case "playing":
		icon = s.props.getString(PlayingIcon, "\uE602 ")
	}
	separator := s.props.getString(TrackSeparator, " - ")
	return fmt.Sprintf("%s%s%s%s", icon, s.artist, separator, s.track)
}

func (s *spotify) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}

func (s *spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.runCommand("osascript", "-e", command)
	return val
}

func (s *spotify) runAutoHotkeyScript(scriptPath string) string {
	var err error

	// check if the script exists
	if _, err := os.Stat(scriptPath); err != nil {
		return "AutoHotkey - script missing"
	}

	// execute authotkey and pipe the output
	cmd := exec.Command("AutoHotkey", scriptPath)
	stdout, _ := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		return "AutoHotkey - " + err.Error()
	}
	out := make([]byte, 1024)
	n, _ := stdout.Read(out)
	return string(out[:n])
}
