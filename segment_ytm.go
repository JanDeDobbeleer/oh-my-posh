package main

import (
	"fmt"
)

type ytm struct {
	props   *properties
	env     environmentInfo
	service ytmStatusService
}

const (
	// YTMDARemoteControlAPIURL is the YTMDA Remote Control API URL property.
	YTMDARemoteControlAPIURL Property = "ytmda_remote_control_api_url"
)

func (y *ytm) string() string {
	// Hit the Remote Control API to find out what song is playing (or paused).
	// https://github.com/ytmdesktop/ytmdesktop/wiki/Remote-Control-API
	status, err := y.service.Get()
	if err != nil {
		return ""
	}

	if status.state == stopped {
		return y.props.getString(StoppedIcon, "\uF04D ")
	}

	icon := ""
	separator := y.props.getString(TrackSeparator, " - ")
	if status.state == paused {
		icon = y.props.getString(PausedIcon, "\uF8E3 ")
	} else {
		icon = y.props.getString(PlayingIcon, "\uE602 ")
	}
	return fmt.Sprintf("%s%s%s%s", icon, status.author, separator, status.title)
}

func (y *ytm) enabled() bool {
	// See if the Remote Control API returns a response.
	_, err := y.service.Get()

	// If we don't get a response back (error), the user isn't running
	// YTMDA, or they don't have the RC API enabled.
	if err != nil {
		return false
	}

	return true
}

func (y *ytm) init(props *properties, env environmentInfo) {
	y.props = props
	y.env = env
	y.service = newYTMDAStatusService(props.getString(YTMDARemoteControlAPIURL, "http://localhost:9863"))
}
