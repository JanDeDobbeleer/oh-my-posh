//go:build windows

package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWindowsNative(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Title           string
		Error           error
	}{
		{
			Case:            "Playing",
			ExpectedString:  "\ue602 Candlemass - Spellbreaker",
			ExpectedEnabled: true,
			Title:           "Candlemass - Spellbreaker",
		},
		{
			Case:            "Stopped",
			ExpectedEnabled: false,
			Title:           "Spotify premium",
		},
		{
			Case:            "Playing - new",
			ExpectedString:  "\ue602 Demon Hunter - Collapsing (feat. Björn \"Speed\" Strid)",
			ExpectedEnabled: true,
			Title:           `Demon Hunter - Collapsing (feat. Björn "Speed" Strid)`,
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("QueryWindowTitles", "spotify.exe", `^(Spotify.*)|(.*\s-\s.*)$`).Return(tc.Title, tc.Error)
		env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return("", &platform.NotImplemented{})
		s := &Spotify{
			env:   env,
			props: properties.Map{},
		}
		assert.Equal(t, tc.ExpectedEnabled, s.Enabled())
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s))
		}
	}
}

func TestSpotifyWindowsPWA(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Title           string
		Error           error
	}{
		{
			Case:            "Playing",
			ExpectedString:  "\ue602 Sarah, the Illstrumentalist - Snow in Stockholm",
			ExpectedEnabled: true,
			Title:           "Spotify - Snow in Stockholm • Sarah, the Illstrumentalist",
		},
		{
			Case:            "Playing",
			ExpectedString:  "\ue602 Main one - Bring the drama",
			ExpectedEnabled: true,
			Title:           "Spotify - Bring the drama • Main one",
		},
		{
			Case:            "Stopped",
			ExpectedEnabled: false,
			Title:           "Spotify - Web Player",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("QueryWindowTitles", "spotify.exe", "^(Spotify.*)|(.*\\s-\\s.*)$").Return("", &platform.NotImplemented{})
		env.On("QueryWindowTitles", "msedge.exe", "^(Spotify.*)").Return(tc.Title, tc.Error)
		s := &Spotify{
			env:   env,
			props: properties.Map{},
		}
		assert.Equal(t, tc.ExpectedEnabled, s.Enabled())
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s))
		}
	}
}
