package segments

import (
	"errors"
	httplib "net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	ytmdaStatusURL = auth.YTMDABASEURL + "/state"
)

type Ytm struct {
	Base

	MusicPlayer
}

func (y *Ytm) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }} "
}

func (y *Ytm) Enabled() bool {
	return y.Enable(y.setStatus)
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
	token, OK := cache.Device.Get[string](auth.YTMDATOKEN)
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
	case -1: // stopped
		y.Status = stopped
	default: // paused
		y.Status = paused
	}

	if status.Player.AdPlaying {
		y.Status = ad
	}

	y.Artist = status.Video.Author
	y.Track = status.Video.Title
	y.resolveIcon(y.options)

	return nil
}

func (y *Ytm) requestStatus(token string) (*ytmdaStatusResponse, error) {
	setHeaders := func(request *httplib.Request) {
		request.Header.Set("Authorization", token)
		request.Header.Set("Content-Type", "application/json")
	}

	request := &http.Request{
		Env:         y.env,
		HTTPTimeout: y.options.Int(options.HTTPTimeout, 5000),
	}

	status, err := request.Do[ytmdaStatusResponse](ytmdaStatusURL, nil, setHeaders)
	return &status, err
}
