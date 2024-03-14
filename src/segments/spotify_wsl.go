//go:build !darwin && !windows

package segments

import (
	"strings"
)

func (s *Spotify) Enabled() bool {
	if !s.env.IsWsl() {
		return false
	}

	tlist, err := s.env.RunCommand("tasklist.exe", "/V", "/FI", "Imagename eq Spotify.exe", "/FO", "CSV", "/NH")
	if err != nil || strings.HasPrefix(tlist, "INFO") {
		return false
	}

	records := strings.Split(tlist, "\n")
	if len(records) == 0 {
		return false
	}

	for _, record := range records {
		record = strings.TrimSpace(record)
		fields := strings.Split(record, ",")
		if len(fields) == 0 {
			continue
		}

		// last elemant is the title
		title := fields[len(fields)-1]
		// trim leading and trailing quotes from the field
		title = strings.TrimPrefix(title, `"`)
		title = strings.TrimSuffix(title, `"`)
		if !strings.Contains(title, " - ") {
			continue
		}

		infos := strings.Split(title, " - ")
		s.Artist = infos[0]
		s.Track = strings.Join(infos[1:], " - ")
		s.Status = playing
		s.resolveIcon()
		return true
	}

	return false
}
