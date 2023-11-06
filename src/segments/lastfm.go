package segments

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type LastFM struct {
	props properties.Properties
	env   platform.Environment

	Artist string
	Track  string
	Full   string
	Icon   string
	Status string
}

const (
	// LastFM username
	Username properties.Property = "username"
)

type lmfDate struct {
	UnixString string `json:"uts"`
}

type lfmTrackInfo struct {
	IsPlaying *string `json:"nowplaying,omitempty"`
}

type Artist struct {
	Name string `json:"#text"`
}

type lfmTrack struct {
	Artist `json:"artist"`
	Name   string        `json:"name"`
	Info   *lfmTrackInfo `json:"@attr"`
	Date   lmfDate       `json:"date"`
}

type tracks struct {
	Tracks []*lfmTrack `json:"track"`
}

type lfmDataResponse struct {
	TracksInfo tracks `json:"recenttracks"`
}

func (d *LastFM) Enabled() bool {
	err := d.setStatus()

	if err != nil {
		d.env.Error(err)
		return false
	}

	return true
}

func (d *LastFM) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Full }}{{ end }} "
}

func (d *LastFM) getResult() (*lfmDataResponse, error) {
	cacheTimeout := d.props.GetInt(properties.CacheTimeout, 0)
	response := new(lfmDataResponse)

	apikey := d.props.GetString(APIKey, ".")
	username := d.props.GetString(Username, ".")
	httpTimeout := d.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	url := fmt.Sprintf("https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&api_key=%s&user=%s&format=json&limit=1", apikey, username)

	if cacheTimeout > 0 {
		val, found := d.env.Cache().Get(url)

		if found {
			err := json.Unmarshal([]byte(val), response)
			if err != nil {
				return nil, err
			}
			return response, nil
		}
	}

	body, err := d.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return new(lfmDataResponse), err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return new(lfmDataResponse), err
	}

	if cacheTimeout > 0 {
		d.env.Cache().Set(url, string(body), cacheTimeout)
	}
	return response, nil
}

func (d *LastFM) setStatus() error {
	q, err := d.getResult()
	if err != nil {
		return err
	}

	if len(q.TracksInfo.Tracks) == 0 {
		return errors.New("No data found")
	}

	track := q.TracksInfo.Tracks[0]

	d.Artist = track.Artist.Name
	d.Track = track.Name
	d.Full = fmt.Sprintf("%s - %s", d.Artist, d.Track)

	isPlaying := false
	if track.Info != nil && track.Info.IsPlaying != nil && *track.Info.IsPlaying == "true" {
		isPlaying = true
	}

	if isPlaying {
		d.Icon = d.props.GetString(PlayingIcon, "\uE602 ")
		d.Status = "playing"
	} else {
		d.Icon = d.props.GetString(StoppedIcon, "\uF04D ")
		d.Status = "stopped"
	}

	return nil
}

func (d *LastFM) Init(props properties.Properties, env platform.Environment) {
	d.props = props
	d.env = env
}
