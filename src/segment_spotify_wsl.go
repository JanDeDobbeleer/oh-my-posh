//go:build !darwin && !windows

package main

import (
	"encoding/csv"
	"strings"
)

func (s *spotify) enabled() bool {
	if !s.env.isWsl() {
		return false
	}
	tlist, err := s.env.runCommand("tasklist.exe", "/V", "/FI", "Imagename eq Spotify.exe", "/FO", "CSV", "/NH")
	if err != nil || strings.HasPrefix(tlist, "INFO") {
		return false
	}
	records, err := csv.NewReader(strings.NewReader(tlist)).ReadAll()
	if err != nil || len(records) == 0 {
		return false
	}
	for _, record := range records {
		title := record[len(record)-1]
		if strings.Contains(title, " - ") {
			infos := strings.Split(title, " - ")
			s.Artist = infos[0]
			s.Track = strings.Join(infos[1:], " - ")
			s.Status = playing
			s.resolveIcon()
			return true
		}
	}
	return false
}
