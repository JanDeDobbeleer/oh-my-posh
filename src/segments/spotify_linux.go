//go:build linux && !darwin && !windows

package segments

import (
	"encoding/json"
	"strings"
)

type Metadata struct {
	TrackID     string   `json:"mpris:trackid"`
	Length      uint64   `json:"mpris:length"`
	ArtURL      string   `json:"mpris:artUrl"`
	Album       string   `json:"xesam:album"`
	AlbumArtist []string `json:"xesam:albumArtist"`
	Artist      []string `json:"xesam:artist"`
	AutoRating  float64  `json:"xesam:autoRating"`
	DiscNumber  int32    `json:"xesam:discNumber"`
	Title       string   `json:"xesam:title"`
	TrackNumber int32    `json:"xesam:trackNumber"`
	URL         string   `json:"xesam:url"`
}

func (s *Spotify) Enable() bool {

	tlist, _ := s.env.RunCommand("dbus-send", "--print-reply", "--dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata")

	// Split the output into lines
	lines := strings.Split(tlist, "\n")

	// Extract the JSON-formatted data
	var jsonData string
	for _, line := range lines {
		if strings.Contains(line, "variant") {
			jsonData = strings.TrimSpace(line[len("   variant       array ["):])
			break
		}
	}

	var metadata []Metadata
	// Parsing JSON data from the output string
	if err := json.Unmarshal([]byte(jsonData), &metadata); err != nil {
		return true
	}

	return false
}
