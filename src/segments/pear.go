package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	httplib "net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// PearPort is the port on which the Pear Desktop API Server listens.
const PearPort options.Option = "port"

const defaultPearPort = 26538

type Pear struct {
	Base

	MusicPlayer
}

func (p *Pear) Template() string {
	return " {{ .Icon }}{{ .Artist }} - {{ .Track }} "
}

func (p *Pear) Enabled() bool {
	err := p.setStatus()
	if err != nil {
		log.Error(err)
	}

	return err == nil
}

type pearSongInfo struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	IsPaused bool   `json:"isPaused"`
}

func (p *Pear) setStatus() error {
	token, OK := cache.Device.Get[string](auth.PEARTOKEN)
	if !OK || token == "" {
		return errors.New("pear token not found, please authenticate using `oh-my-posh auth pear`")
	}

	status, err := p.requestSongInfo(token)
	if err != nil {
		return err
	}

	p.Status = playing
	p.Icon = p.options.Markup(PlayingIcon, "\uf04b ")

	if status.IsPaused {
		p.Status = paused
		p.Icon = p.options.Markup(PausedIcon, "\uf04c ")
	}

	p.Artist = status.Artist
	p.Track = status.Title

	return nil
}

func (p *Pear) requestSongInfo(token string) (*pearSongInfo, error) {
	port := p.options.Int(PearPort, defaultPearPort)
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/song-info", port)

	setHeaders := func(request *httplib.Request) {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	httpTimeout := p.options.Int(options.HTTPTimeout, 5000)
	response, err := p.env.HTTPRequest(url, nil, httpTimeout, setHeaders)
	if err != nil {
		return nil, err
	}

	var result pearSongInfo
	err = json.Unmarshal(response, &result)
	return &result, err
}
