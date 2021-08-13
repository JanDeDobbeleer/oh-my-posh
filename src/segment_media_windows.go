// +build windows

package main

import (
	"encoding/json"
	"os/exec"
)

func (s *media) enabled() bool {
	var cmd = exec.Command("sys-media-info")
	var str, err = cmd.Output()
	if err == nil {
		json.Unmarshal(str, &s.info)
		return true
	}
	s.other, err = s.env.getWindowTitle("cloudmusic.exe", "^(.*\\s-\\s.*)$")
	if err == nil && s.other != "" {
		return true
	}
	s.other, err = s.env.getWindowTitle("qqmusic.exe", "^(.*\\s-\\s.*)$")
	if err == nil && s.other != "" {
		return true
	}
	s.other, err = s.env.getWindowTitle("spotify.exe", "^(.*\\s-\\s.*)$")
	if err == nil && s.other != "" {
		return true
	}
	return false
}
