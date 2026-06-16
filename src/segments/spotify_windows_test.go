//go:build windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWindowsSMTC(t *testing.T) {
	cases := []struct {
		Info            *runtime.MediaInfo
		Error           error
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "playing",
			Info:            &runtime.MediaInfo{Status: "playing", Title: "Spellbreaker", Artist: "Candlemass", Album: "Nightfall", TrackNumber: 3},
			ExpectedString:  " Candlemass - Spellbreaker",
			ExpectedEnabled: true,
		},
		{
			Case:            "paused",
			Info:            &runtime.MediaInfo{Status: "paused", Title: "Spellbreaker", Artist: "Candlemass", Album: "Nightfall", TrackNumber: 3},
			ExpectedString:  " Candlemass - Spellbreaker",
			ExpectedEnabled: true,
		},
		{
			Case:            "ad (empty album, track number 0)",
			Info:            &runtime.MediaInfo{Status: "playing", Title: "君のこころが観たいもの　プライムビデオ", Artist: "プライムビデオ", Album: "", TrackNumber: 0},
			ExpectedString:  " プライムビデオ - 君のこころが観たいもの　プライムビデオ",
			ExpectedEnabled: true,
		},
		{
			Case:            "stopped",
			Info:            &runtime.MediaInfo{Status: "stopped"},
			ExpectedEnabled: false,
		},
		{
			Case:            "closed (no Spotify session)",
			Info:            &runtime.MediaInfo{Status: "closed"},
			ExpectedEnabled: false,
		},
		{
			Case:            "track with parentheses and quotes",
			Info:            &runtime.MediaInfo{Status: "playing", Title: `Collapsing (feat. Björn "Speed" Strid)`, Artist: "Demon Hunter", Album: "The World Is a Thorn", TrackNumber: 9},
			ExpectedString:  " Demon Hunter - Collapsing (feat. Björn \"Speed\" Strid)",
			ExpectedEnabled: true,
		},
		{
			Case:            "smtc error",
			Error:           errors.New("oops"),
			ExpectedEnabled: false,
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("QueryMediaPlayer", "spotify").Return(tc.Info, tc.Error)
		// Edge PWA fallback never matches in this group of tests.
		env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return("", &runtime.NotImplemented{})

		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, s.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}

func TestSpotifyWindowsPWA(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		ExpectedString  string
		Title           string
		ExpectedEnabled bool
	}{
		{
			Case:            "playing",
			ExpectedString:  " Sarah, the Illstrumentalist - Snow in Stockholm",
			ExpectedEnabled: true,
			Title:           "Spotify - Snow in Stockholm • Sarah, the Illstrumentalist",
		},
		{
			Case:            "stopped",
			ExpectedEnabled: false,
			Title:           "Spotify - Web Player",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		// SMTC unavailable — falls back to Edge PWA window title parsing.
		env.On("QueryMediaPlayer", "spotify").Return((*runtime.MediaInfo)(nil), errors.New("smtc unavailable"))
		env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return(tc.Title, tc.Error)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, s.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}
