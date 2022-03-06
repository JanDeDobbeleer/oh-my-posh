//go:build windows

package segments

import (
	"errors"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

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
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("QueryWindowTitles", "spotify.exe", "^(Spotify.*)|(.*\\s-\\s.*)$").Return(tc.Title, tc.Error)
		env.On("QueryWindowTitles", "msedge.exe", "^(Spotify.*)").Return("", errors.New("not implemented"))
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
			ExpectedString:  "\ue602 Snow in Stockholm - Sarah, the Illstrumentalist",
			ExpectedEnabled: true,
			Title:           "Spotify - Snow in Stockholm · Sarah, the Illstrumentalist",
		},
		{
			Case:            "Stopped",
			ExpectedEnabled: false,
			Title:           "Spotify - Web Player",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("QueryWindowTitles", "spotify.exe", "^(Spotify.*)|(.*\\s-\\s.*)$").Return("", errors.New("not implemented"))
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
