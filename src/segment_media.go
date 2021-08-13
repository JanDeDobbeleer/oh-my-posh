package main

import (
	"fmt"
	"strings"
	"time"
)

type media struct {
	props *properties
	env   environmentInfo
	info  NowPlayingSessionInfo
	other string
}

const (
	// PlayingIcon indicates a song is playing
	MediaPlayingIcon Property = "playing_icon"
	// PausedIcon indicates a song is paused
	MediaPausedIcon Property = "paused_icon"
	// StoppedIcon indicates a song is stopped
	MediaStoppedIcon Property = "stopped_icon"
	// TrackSeparator is put between the artist and the track
	MediaTrackSeparator Property = "track_separator"
)

func (s *media) string() string {
	separator := s.props.getString(TrackSeparator, " - ")
	if s.other != "" {
		spt := strings.Split(s.other, " - ")
		return fmt.Sprintf("%s%s%s", spt[0], separator, spt[1])
	}
	icon := ""
	switch s.info.Playback.PlaybackState {
	case 4:
		icon = s.props.getString(MediaStoppedIcon, "\uF04D ")
	case 5:
		icon = s.props.getString(MediaPlayingIcon, "\uE602 ")
	case 6:
		icon = s.props.getString(MediaPausedIcon, "\uF8E3 ")
	}
	return fmt.Sprintf("%s%s%s%s", icon, s.info.MediaInfo.Title, separator, s.info.MediaInfo.Artist)
}

func (n *media) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}

type NowPlayingSessionInfo struct {
	Session struct {
		Hwnd struct {
		} `json:"Hwnd"`
		PID            int         `json:"PID"`
		RenderDeviceID string      `json:"RenderDeviceId"`
		SourceAppID    string      `json:"SourceAppId"`
		SourceDeviceID string      `json:"SourceDeviceId"`
		Connection     interface{} `json:"Connection"`
	} `json:"Session"`
	Playback struct {
		PropsValid          int       `json:"PropsValid"`
		PlaybackCaps        int       `json:"PlaybackCaps"`
		PlaybackState       int       `json:"PlaybackState"`
		PlaybackMode        int       `json:"PlaybackMode"`
		RepeatMode          int       `json:"RepeatMode"`
		PlaybackRate        int       `json:"PlaybackRate"`
		ShuffleEnabled      bool      `json:"ShuffleEnabled"`
		LastPlayingFileTime time.Time `json:"LastPlayingFileTime"`
	} `json:"Playback"`
	MediaInfo struct {
		AlbumArtist         string   `json:"AlbumArtist"`
		AlbumTitle          string   `json:"AlbumTitle"`
		Subtitle            string   `json:"Subtitle"`
		Title               string   `json:"Title"`
		Artist              string   `json:"Artist"`
		MediaClassPrimaryID string   `json:"MediaClassPrimaryID"`
		Genres              []string `json:"Genres"`
		AlbumTrackCount     int      `json:"AlbumTrackCount"`
		TrackNumber         int      `json:"TrackNumber"`
	} `json:"MediaInfo"`
	Timeline struct {
		StartTime struct {
			Ticks             int `json:"Ticks"`
			Days              int `json:"Days"`
			Hours             int `json:"Hours"`
			Milliseconds      int `json:"Milliseconds"`
			Minutes           int `json:"Minutes"`
			Seconds           int `json:"Seconds"`
			TotalDays         int `json:"TotalDays"`
			TotalHours        int `json:"TotalHours"`
			TotalMilliseconds int `json:"TotalMilliseconds"`
			TotalMinutes      int `json:"TotalMinutes"`
			TotalSeconds      int `json:"TotalSeconds"`
		} `json:"StartTime"`
		EndTime struct {
			Ticks             int     `json:"Ticks"`
			Days              int     `json:"Days"`
			Hours             int     `json:"Hours"`
			Milliseconds      int     `json:"Milliseconds"`
			Minutes           int     `json:"Minutes"`
			Seconds           int     `json:"Seconds"`
			TotalDays         float64 `json:"TotalDays"`
			TotalHours        float64 `json:"TotalHours"`
			TotalMilliseconds int     `json:"TotalMilliseconds"`
			TotalMinutes      float64 `json:"TotalMinutes"`
			TotalSeconds      float64 `json:"TotalSeconds"`
		} `json:"EndTime"`
		MinSeekTime struct {
			Ticks             int `json:"Ticks"`
			Days              int `json:"Days"`
			Hours             int `json:"Hours"`
			Milliseconds      int `json:"Milliseconds"`
			Minutes           int `json:"Minutes"`
			Seconds           int `json:"Seconds"`
			TotalDays         int `json:"TotalDays"`
			TotalHours        int `json:"TotalHours"`
			TotalMilliseconds int `json:"TotalMilliseconds"`
			TotalMinutes      int `json:"TotalMinutes"`
			TotalSeconds      int `json:"TotalSeconds"`
		} `json:"MinSeekTime"`
		MaxSeekTime struct {
			Ticks             int     `json:"Ticks"`
			Days              int     `json:"Days"`
			Hours             int     `json:"Hours"`
			Milliseconds      int     `json:"Milliseconds"`
			Minutes           int     `json:"Minutes"`
			Seconds           int     `json:"Seconds"`
			TotalDays         float64 `json:"TotalDays"`
			TotalHours        float64 `json:"TotalHours"`
			TotalMilliseconds int     `json:"TotalMilliseconds"`
			TotalMinutes      float64 `json:"TotalMinutes"`
			TotalSeconds      float64 `json:"TotalSeconds"`
		} `json:"MaxSeekTime"`
		Position struct {
			Ticks             int     `json:"Ticks"`
			Days              int     `json:"Days"`
			Hours             int     `json:"Hours"`
			Milliseconds      int     `json:"Milliseconds"`
			Minutes           int     `json:"Minutes"`
			Seconds           int     `json:"Seconds"`
			TotalDays         float64 `json:"TotalDays"`
			TotalHours        float64 `json:"TotalHours"`
			TotalMilliseconds int     `json:"TotalMilliseconds"`
			TotalMinutes      float64 `json:"TotalMinutes"`
			TotalSeconds      float64 `json:"TotalSeconds"`
		} `json:"Position"`
		PositionSetFileTime time.Time `json:"PositionSetFileTime"`
	} `json:"Timeline"`
}
