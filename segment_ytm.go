package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type ytm struct {
	props   *properties
	env     environmentInfo
	service ytmStatusService
	status  *ytmStatus
}

const (
	// YTMDARemoteControlAPIURL is the YTMDA Remote Control API URL property.
	YTMDARemoteControlAPIURL Property = "ytmda_remote_control_api_url"
)

func (y *ytm) string() string {
	if y.status.state == stopped {
		return y.props.getString(StoppedIcon, "\uF04D ")
	}

	icon := ""
	separator := y.props.getString(TrackSeparator, " - ")
	switch y.status.state {
	case paused:
		icon = y.props.getString(PausedIcon, "\uF8E3 ")
	default:
		icon = y.props.getString(PlayingIcon, "\uE602 ")
	}

	return fmt.Sprintf("%s%s%s%s", icon, y.status.author, separator, y.status.title)
}

func (y *ytm) enabled() bool {
	// See if the Remote Control API returns a response.
	// https://github.com/ytmdesktop/ytmdesktop/wiki/Remote-Control-API
	status, err := y.service.Get()

	// "Cache" the status so we don't have to call the RC API again.
	// We'll use this in the string() method.
	y.status = status

	// If we don't get a response back (error), the user isn't running
	// YTMDA, or they don't have the RC API enabled.
	return err == nil
}

func (y *ytm) init(props *properties, env environmentInfo) {
	y.props = props
	y.env = env
	y.service = newYTMDAStatusService(props.getString(YTMDARemoteControlAPIURL, "http://localhost:9863"))
}

type ytmPlayingState int

const (
	playing ytmPlayingState = iota
	paused
	stopped
)

type ytmStatus struct {
	state  ytmPlayingState
	author string
	title  string
}

type ytmStatusService interface {
	Get() (*ytmStatus, error)
}

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
	resp, err := doGet(s.url+"/query", nil)
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
