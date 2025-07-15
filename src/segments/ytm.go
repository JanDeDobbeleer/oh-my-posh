package segments

import (
	"encoding/json"
	"errors"
	httplib "net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const (
	ytmdaStatusURL = auth.YTMDABASEURL + "/state"
)

type Ytm struct {
	base

	MusicPlayer
}

func (y *Ytm) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }} "
}

func (y *Ytm) Enabled() bool {
	err := y.setStatus()
	if err != nil {
		log.Error(err)
	}

	return err == nil
}

type ytmdaStatusResponse struct {
	Video struct {
		Author string `json:"author"`
		Title  string `json:"title"`
	} `json:"video"`
	Player struct {
		TrackState int  `json:"trackState"`
		AdPlaying  bool `json:"adPlaying"`
	} `json:"player"`
}

func (y *Ytm) setStatus() error {
	token, OK := y.env.Cache().Get(auth.YTMDATOKEN)
	if !OK || token == "" {
		return errors.New("YTMDA token not found, please authenticate using `oh-my-posh auth ytmda`")
	}

	status, err := y.requestStatus(token)
	if err != nil {
		return err
	}

	switch status.Player.TrackState {
	case 1, 2: // playing or buffering
		y.Status = playing
		y.Icon = y.props.GetString(PlayingIcon, "\uf04b ")
	case -1: // stopped
		y.Status = stopped
		y.Icon = y.props.GetString(StoppedIcon, "\uf04d ")
	default: // paused
		y.Status = paused
		y.Icon = y.props.GetString(PausedIcon, "\uf04c ")
	}

	if status.Player.AdPlaying {
		ad := y.props.GetString(AdIcon, "\ueebb ")
		y.Icon = ad + y.Icon
	}

	y.Artist = status.Video.Author
	y.Track = status.Video.Title

	return nil
}

func (y *Ytm) requestStatus(token string) (*ytmdaStatusResponse, error) {
	setHeaders := func(request *httplib.Request) {
		request.Header.Set("Authorization", token)
		request.Header.Set("Content-Type", "application/json")
	}

	httpTimeout := y.props.GetInt(properties.HTTPTimeout, 5000)
	response, err := y.env.HTTPRequest(ytmdaStatusURL, nil, httpTimeout, setHeaders)
	if err != nil {
		return nil, err
	}

	var result ytmdaStatusResponse
	err = json.Unmarshal(response, &result)
	return &result, err
}
