package segments

import (
	"encoding/json"
	"fmt"
	httplib "net/http"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

const YTMDATOKEN = "ytmda_token"

type Ytm struct {
	base

	MusicPlayer
}

func (y *Ytm) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }} "
}

func (y *Ytm) Enabled() bool {
	err := y.setStatus()
	// If we don't get a response back (error), the user isn't running
	// YTMDA, or they don't have the Companion API enabled.
	return err == nil
}

type ytmdaStatusResponse struct {
	Player struct {
		TrackState int  `json:"trackState"`
		AdPlaying  bool `json:"adPlaying"`
	} `json:"player"`
	Video struct {
		Author string `json:"author"`
		Title  string `json:"title"`
	} `json:"video"`
}

func (y *Ytm) setStatus() error {
	token, err := y.getToken()
	if err != nil {
		return err
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

func (y *Ytm) getToken() (string, error) {
	if token, OK := y.env.Cache().Get(YTMDATOKEN); OK && len(token) > 0 {
		return token, nil
	}

	code, err := y.requestCode()
	if err != nil {
		return "", err
	}

	token, err := y.requestToken(code)
	if err != nil {
		return "", err
	}

	y.env.Cache().Set(YTMDATOKEN, token, cache.INFINITE)

	return token, nil
}

const baseURL = "http://localhost:9863/api/v1"

func (y *Ytm) requestCode() (string, error) {
	url := baseURL + "/auth/requestcode"
	body := fmt.Sprintf(`{"appId": "ohmyposh", "appName": "oh-my-posh", "appVersion": "%s"}`, strings.TrimPrefix(build.Version, "v"))

	type codeResponse struct {
		Code string `json:"code"`
	}

	result, err := ytdmaRequest[codeResponse](httplib.MethodPost, url, body, y.env)

	return result.Code, err
}

func (y *Ytm) requestToken(code string) (string, error) {
	url := baseURL + "/auth/request"
	body := fmt.Sprintf(`{"appId": "ohmyposh", "code": "%s"}`, code)

	type tokenResponse struct {
		Token string `json:"token"`
	}

	result, err := ytdmaRequest[tokenResponse](httplib.MethodPost, url, body, y.env)

	return result.Token, err
}

func (y *Ytm) requestStatus(token string) (*ytmdaStatusResponse, error) {
	url := baseURL + "/state"

	setAuth := func(request *httplib.Request) {
		request.Header.Set("Authorization", token)
	}

	response, err := ytdmaRequest[ytmdaStatusResponse](httplib.MethodGet, url, "", y.env, setAuth)

	return &response, err
}

func ytdmaRequest[a any](method, url, body string, env runtime.Environment, requestModifiers ...http.RequestModifier) (a, error) {
	if requestModifiers == nil {
		requestModifiers = []http.RequestModifier{}
	}

	modifyRequest := func(request *httplib.Request) {
		request.Method = method
		request.Header.Set("Content-Type", "application/json")
	}

	requestModifiers = append(requestModifiers, modifyRequest)

	var result a

	response, err := env.HTTPRequest(url, strings.NewReader(body), 50000, requestModifiers...)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(response, &result)
	return result, err
}
