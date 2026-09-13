package segments

import (
	"errors"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type LastFM struct {
	Base

	MusicPlayer

	Full string
}

const (
	Username options.Option = "username"
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
	return d.Enable(d.setStatus)
}

func (d *LastFM) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Full }}{{ end }} "
}

func (d *LastFM) getResult() (*lfmDataResponse, error) {
	apikey := d.options.Template(APIKey, ".", d)
	username := d.options.Template(Username, ".", d)

	url := fmt.Sprintf("https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&api_key=%s&user=%s&format=json&limit=1", apikey, username)

	request := &http.Request{
		Env:         d.env,
		HTTPTimeout: d.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout),
	}

	response, err := request.Do[lfmDataResponse](url, nil)
	return &response, err
}

func (d *LastFM) setStatus() error {
	q, err := d.getResult()
	if err != nil {
		return err
	}

	if len(q.TracksInfo.Tracks) == 0 {
		return errors.New("no data found")
	}

	track := q.TracksInfo.Tracks[0]

	d.Artist = track.Artist.Name
	d.Track = track.Name
	d.Full = fmt.Sprintf("%s - %s", d.Artist, d.Track)

	isPlaying := track.Info != nil && track.Info.IsPlaying != nil && *track.Info.IsPlaying == "true"

	if isPlaying {
		d.Status = playing
	} else {
		d.Status = stopped
	}

	d.resolveIcon(d.options)

	return nil
}
