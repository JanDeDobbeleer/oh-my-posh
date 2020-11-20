package main

import (
	"encoding/json"
	"io/ioutil"
)

type ytmdaQuery struct {
	player `json:"player"`
	track  `json:"track"`
}

type player struct {
	HasSong                     bool    `json:"hasSong"`
	IsPaused                    bool    `json:"isPaused"`
	VolumePercent               int     `json:"volumePercent"`
	SeekbarCurrentPosition      int     `json:"seekbarCurrentPosition"`
	SeekbarCurrentPositionHuman string  `json:"seekbarCurrentPositionHuman"`
	StatePercent                float64 `json:"statePercent"`
	LikeStatus                  string  `json:"likeStatus"`
	RepeatType                  string  `json:"repeatType"`
}

type track struct {
	Author          string `json:"author"`
	Title           string `json:"title"`
	Album           string `json:"album"`
	Cover           string `json:"cover"`
	Duration        int    `json:"duration"`
	DurationHuman   string `json:"durationHuman"`
	URL             string `json:"url"`
	ID              string `json:"id"`
	IsVideo         bool   `json:"isVideo"`
	IsAdvertisement bool   `json:"isAdvertisement"`
	InLibrary       bool   `json:"inLibrary"`
}

type ytmdaStatusService struct {
	url string
}

func (s *ytmdaStatusService) Get() (*ytmStatus, error) {
	resp, err := Get(s.url+"/query", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	q := new(ytmdaQuery)
	err = json.Unmarshal(body, &q)
	if err != nil {
		return nil, err
	}

	return getYTMStatus(q), err
}

func getYTMStatus(q *ytmdaQuery) *ytmStatus {
	state := playing
	if !q.player.HasSong {
		state = stopped
	} else if q.player.IsPaused {
		state = paused
	}

	return &ytmStatus{
		state:  state,
		author: q.track.Author,
		title:  q.track.Title,
	}
}

func newYTMDAStatusService(url string) ytmStatusService {
	return &ytmdaStatusService{
		url: url,
	}
}
