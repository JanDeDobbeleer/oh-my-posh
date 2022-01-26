package segments

import (
	"encoding/json"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Ytm struct {
	props properties.Properties
	env   environment.Environment

	MusicPlayer
}

const (
	// APIURL is the YTMDA Remote Control API URL property.
	APIURL properties.Property = "api_url"
)

func (y *Ytm) Template() string {
	return "{{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }}"
}

func (y *Ytm) Enabled() bool {
	err := y.setStatus()
	// If we don't get a response back (error), the user isn't running
	// YTMDA, or they don't have the RC API enabled.
	return err == nil
}

func (y *Ytm) Init(props properties.Properties, env environment.Environment) {
	y.props = props
	y.env = env
}

type ytmdaStatusResponse struct {
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

func (y *Ytm) setStatus() error {
	// https://github.com/ytmdesktop/ytmdesktop/wiki/Remote-Control-API
	url := y.props.GetString(APIURL, "http://127.0.0.1:9863")
	httpTimeout := y.props.GetInt(APIURL, DefaultHTTPTimeout)
	body, err := y.env.HTTPRequest(url+"/query", httpTimeout)
	if err != nil {
		return err
	}
	q := new(ytmdaStatusResponse)
	err = json.Unmarshal(body, &q)
	if err != nil {
		return err
	}
	y.Status = playing
	y.Icon = y.props.GetString(PlayingIcon, "\uE602 ")
	if !q.player.HasSong {
		y.Status = stopped
		y.Icon = y.props.GetString(StoppedIcon, "\uF04D ")
	} else if q.player.IsPaused {
		y.Status = paused
		y.Icon = y.props.GetString(PausedIcon, "\uF8E3 ")
	}
	y.Artist = q.track.Author
	y.Track = q.track.Title
	return nil
}
