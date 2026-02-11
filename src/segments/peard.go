package segments

import (
	"encoding/json"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type PearDesktop struct {
	Base

	Status         string
	Artist         string
	Track          string
	TrackUrl       string
	ArtistUrl      string
	MediaType      string
	IsPaused       bool
	SongDuration   int
	ElapsedSeconds int
	Icon           string
}

const (
	// Port is the port to connect to Pear Desktop API
	Port options.Option = "port"
)

func (p *PearDesktop) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"paused\" }}{{ .Artist }} - {{ .Track }}{{ end }} "
}

func (p *PearDesktop) Enabled() bool {
	err := p.setStatus()
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}

type pearDesktopResponse struct {
	Title          string `json:"title"`
	Artist         string `json:"artist"`
	ArtistUrl      string `json:"artistUrl"`
	IsPaused       bool   `json:"isPaused"`
	SongDuration   int    `json:"songDuration"`
	ElapsedSeconds int    `json:"elapsedSeconds"`
	Url            string `json:"url"`
	// MediaType can be AUDIO, ORIGINAL_MUSIC_VIDEO, USER_GENERATED_CONTENT, PODCAST_EPISODE, or OTHER_VIDEO
	MediaType string `json:"mediaType"`
	// ignore everything else, they don't provide good info in a cli env
}

func (p *PearDesktop) setStatus() error {
	port := p.options.Int(Port, 26538)
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/song-info", port)

	httpTimeout := p.options.Int(options.HTTPTimeout, 5000)
	response, err := p.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return err
	}

	var result pearDesktopResponse
	if err := json.Unmarshal(response, &result); err != nil {
		return err
	}

	p.Track = result.Title
	p.Artist = result.Artist
	p.TrackUrl = result.Url
	p.ArtistUrl = result.ArtistUrl
	p.MediaType = result.MediaType
	p.IsPaused = result.IsPaused
	p.SongDuration = result.SongDuration
	p.ElapsedSeconds = result.ElapsedSeconds

	if result.IsPaused {
		p.Status = paused
		p.Icon = p.options.String(PausedIcon, "\uf04c ")
		return nil
	}

	p.Status = playing
	p.Icon = p.options.String(PlayingIcon, "\uf04b ")
	return nil
}
