package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaInfo(t *testing.T) {
	cases := []struct {
		Case           string
		JsonText       string
		WindowTitle    string
		ExpectedString string
		TrackSeparator string
	}{
		{
			Case:           "No playing session and other player",
			JsonText:       "{}",
			ExpectedString: "",
			TrackSeparator: " - ",
		},
		{
			Case:           "Get info from media session",
			JsonText:       "{\"Playback\":{\"PlaybackState\":5},\"MediaInfo\":{\"Title\":\"Believe in you\",\"Artist\":\"nonoc\"},\"Timeline\":{\"EndTime\":{\"Seconds\":39,\"TotalMinutes\":4.65555},\"Position\":{\"Seconds\":3,\"TotalMinutes\":0.054933333333333334}}}",
			ExpectedString: "\uE602Believe in you - nonoc",
			TrackSeparator: " - ",
		},
		{
			Case:           "Get info from player window title",
			JsonText:       "{}",
			WindowTitle:    "Believe in you - nonoc",
			ExpectedString: "Believe in you - nonoc",
			TrackSeparator: " - ",
		},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("isWsl", nil).Return(false)
		env.On("hasCommand", "sys-media-info").Return(true)
		env.On("runCommand", "sys-media-info", []string{"--json"}).Return(tc.JsonText, nil)
		env.On("getWindowTitle", "cloudmusic.exe").Return(tc.WindowTitle, nil)
		env.On("getWindowTitle", "qqmusic.exe").Return(tc.WindowTitle, nil)
		env.On("getWindowTitle", "spotify.exe").Return(tc.WindowTitle, nil)
		props := &properties{
			values: map[Property]interface{}{
				MediaPlayingIcon:    "\uE602",
				MediaPausedIcon:     "\uF8E3",
				MediaStoppedIcon:    "\uF04D",
				MediaTrackSeparator: " - ",
				MediaTimeSeparator:  "/",
				MediaIsShowTime:     false,
			},
		}
		media := &media{
			env:   env,
			props: props,
		}
		enabled := media.enabled()
		if !enabled {
			continue
		}
		assert.Equal(t, tc.ExpectedString, media.string(), tc.Case)
	}
}

func TestMediaPlayState(t *testing.T) {
	cases := []struct {
		Case           string
		JsonText       string
		ExpectedString string
		TrackSeparator string
	}{
		{
			Case:           "Playing",
			JsonText:       "{\"Playback\":{\"PlaybackState\":5},\"MediaInfo\":{\"Title\":\"Believe in you\",\"Artist\":\"nonoc\"},\"Timeline\":{\"EndTime\":{\"Seconds\":39,\"TotalMinutes\":4.65555},\"Position\":{\"Seconds\":3,\"TotalMinutes\":0.054933333333333334}}}",
			ExpectedString: "\uE602Believe in you - nonoc",
			TrackSeparator: " - ",
		},
		{
			Case:           "Paused",
			JsonText:       "{\"Playback\":{\"PlaybackState\":6},\"MediaInfo\":{\"Title\":\"Believe in you\",\"Artist\":\"nonoc\"},\"Timeline\":{\"EndTime\":{\"Seconds\":39,\"TotalMinutes\":4.65555},\"Position\":{\"Seconds\":3,\"TotalMinutes\":0.054933333333333334}}}",
			ExpectedString: "\uF8E3Believe in you - nonoc",
			TrackSeparator: " - ",
		},
		{
			Case:           "Stopped",
			JsonText:       "{\"Playback\":{\"PlaybackState\":4},\"MediaInfo\":{\"Title\":\"Believe in you\",\"Artist\":\"nonoc\"},\"Timeline\":{\"EndTime\":{\"Seconds\":39,\"TotalMinutes\":4.65555},\"Position\":{\"Seconds\":3,\"TotalMinutes\":0.054933333333333334}}}",
			ExpectedString: "\uF04DBelieve in you - nonoc",
			TrackSeparator: " - ",
		},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("isWsl", nil).Return(false)
		env.On("hasCommand", "sys-media-info").Return(true)
		env.On("runCommand", "sys-media-info", []string{"--json"}).Return(tc.JsonText, nil)
		env.On("getWindowTitle", "cloudmusic.exe").Return("", nil)
		env.On("getWindowTitle", "qqmusic.exe").Return("", nil)
		env.On("getWindowTitle", "spotify.exe").Return("", nil)
		props := &properties{
			values: map[Property]interface{}{
				MediaPlayingIcon:    "\uE602",
				MediaPausedIcon:     "\uF8E3",
				MediaStoppedIcon:    "\uF04D",
				MediaTrackSeparator: " - ",
				MediaTimeSeparator:  "/",
				MediaIsShowTime:     false,
			},
		}
		media := &media{
			env:   env,
			props: props,
		}
		enabled := media.enabled()
		if !enabled {
			continue
		}
		assert.Equal(t, tc.ExpectedString, media.string(), tc.Case)
	}
}
