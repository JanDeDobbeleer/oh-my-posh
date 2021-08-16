package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// MediaPlayingIcon indicates a song is playing
	MediaPlayingIcon Property = "playing_icon"
	// MediaPausedIcon indicates a song is paused
	MediaPausedIcon Property = "paused_icon"
	// MediaStoppedIcon indicates a song is stopped
	MediaStoppedIcon Property = "stopped_icon"
	// MediaTrackSeparator is put between the artist and the track
	MediaTrackSeparator Property = "track_separator"
	// MediaTimeSeparator is put between the media position and total time
	MediaTimeSeparator Property = "time_separator"
	// MediaShowTime is media time show or hidden switch
	MediaIsShowTime Property = "is_show_time"
)

type CShapTimeSpan struct {
	Ticks             int     `json:"Ticks"`
	Days              int     `json:"Days"`
	Hours             int     `json:"Hours"`
	Milliseconds      int     `json:"Milliseconds"`
	Minutes           int     `json:"Minutes"`
	Seconds           int     `json:"Seconds"`
	TotalDays         float64 `json:"TotalDays"`
	TotalHours        float64 `json:"TotalHours"`
	TotalMilliseconds float64 `json:"TotalMilliseconds"`
	TotalMinutes      float64 `json:"TotalMinutes"`
	TotalSeconds      float64 `json:"TotalSeconds"`
}

type NowPlayingSessionInfo struct {
	Session struct {
		PID            int    `json:"PID"`
		RenderDeviceID string `json:"RenderDeviceId"`
		SourceAppID    string `json:"SourceAppId"`
		SourceDeviceID string `json:"SourceDeviceId"`
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
		StartTime           CShapTimeSpan `json:"StartTime"`
		EndTime             CShapTimeSpan `json:"EndTime"`
		MinSeekTime         CShapTimeSpan `json:"MinSeekTime"`
		MaxSeekTime         CShapTimeSpan `json:"MaxSeekTime"`
		Position            CShapTimeSpan `json:"Position"`
		PositionSetFileTime time.Time     `json:"PositionSetFileTime"`
	} `json:"Timeline"`
}

type media struct {
	props *properties
	env   environmentInfo
	info  NowPlayingSessionInfo
	other string
}

func (s *media) enabled() bool {
	tool := "sys-media-info"
	if s.env.isWsl() {
		tool += ".exe"
	}
	if s.env.hasCommand(tool) {
		str, err := s.env.runCommand(tool, "--json")
		if err == nil && str != "{}" {
			json.Unmarshal([]byte(str), &s.info)
			return true
		}
	}
	players := [...]string{"cloudmusic.exe", "qqmusic.exe", "spotify.exe"}
	for _, player := range players {
		title, err := s.env.getWindowTitle(player, "^(.*\\s-\\s.*)$")
		if err == nil && title != "" {
			s.other = title
			return true
		}
	}
	return false
}

func (s *media) string() string {
	separator := s.props.getString(TrackSeparator, " - ")
	time_separator := s.props.getString(MediaTimeSeparator, "/")
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
	str := icon
	if s.props.getBool(MediaIsShowTime, true) && s.info.Timeline.Position.TotalSeconds > 1 && s.info.Timeline.EndTime.TotalSeconds > 1 {
		str = str + fmt.Sprintf("[%d:%02d%s%d:%02d] ", int64(s.info.Timeline.Position.TotalMinutes), s.info.Timeline.Position.Seconds, time_separator, int64(s.info.Timeline.EndTime.TotalMinutes), s.info.Timeline.EndTime.Seconds)
	}
	str = str + fmt.Sprintf("%s%s%s", s.info.MediaInfo.Title, separator, s.info.MediaInfo.Artist)
	return str
}

func (n *media) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}
