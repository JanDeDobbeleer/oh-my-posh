//go:build linux && !darwin && !windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyLinux(t *testing.T) {
	cases := []struct {
		Case            string
		Status          string
		Artist          string
		Track           string
		Album           string
		TrackNumber     string
		Expected        string
		ExpectedEnabled bool
	}{
		{Case: "no data", ExpectedEnabled: false},
		{Case: "error", ExpectedEnabled: false, Status: "Error.ServiceUnknown"},
		{
			Case: "paused", ExpectedEnabled: true, Expected: "\uf04c Candlemass - Spellbreaker",
			Status: "paused", Artist: "Candlemass", Track: "Spellbreaker", Album: "Nightfall", TrackNumber: "3",
		},
		{
			Case: "playing", ExpectedEnabled: true, Expected: "\ue602 Candlemass - Spellbreaker",
			Status: "playing", Artist: "Candlemass", Track: "Spellbreaker", Album: "Nightfall", TrackNumber: "3",
		},
		{Case: "ad", ExpectedEnabled: true, Expected: "\ueebb Spotify - Try Premium for free", Status: "playing", Artist: "Spotify", Track: "Try Premium for free"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)

		env.On("RunShellCommand", shell.BASH, spotifyDBusCommand+spotifyPlaybackStatusCommand).Return(tc.Status)
		env.On("RunShellCommand", shell.BASH, spotifyDBusCommand+spotifyArtistCommand).Return(tc.Artist)
		env.On("RunShellCommand", shell.BASH, spotifyDBusCommand+spotifyTrackCommand).Return(tc.Track)
		env.On("RunShellCommand", shell.BASH, spotifyDBusCommand+spotifyAlbumCommand).Return(tc.Album)
		env.On("RunShellCommand", shell.BASH, spotifyDBusCommand+spotifyTrackNumberCommand).Return(tc.TrackNumber)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		got := s.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, got, tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}

func TestSpotifyWSL(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		Output          string
		Expected        string
		ExpectedEnabled bool
	}{
		{
			Case:            "playing",
			Output:          "playing|Spellbreaker|Candlemass|Nightfall|3",
			Expected:        "\ue602 Candlemass - Spellbreaker",
			ExpectedEnabled: true,
		},
		{
			Case:            "paused",
			Output:          "paused|Spellbreaker|Candlemass|Nightfall|3",
			Expected:        "\uf04c Candlemass - Spellbreaker",
			ExpectedEnabled: true,
		},
		{
			Case:            "ad (empty album, track number 0)",
			Output:          "playing|Try Premium for free|Spotify||0",
			Expected:        "\ueebb Spotify - Try Premium for free",
			ExpectedEnabled: true,
		},
		{
			Case:            "stopped",
			Output:          "stopped||||0",
			ExpectedEnabled: false,
		},
		{
			Case:            "closed (no Spotify session)",
			Output:          "closed||||0",
			ExpectedEnabled: false,
		},
		{
			Case:            "powershell error",
			Error:           errors.New("oops"),
			ExpectedEnabled: false,
		},
		{
			Case:            "malformed output",
			Output:          "garbage",
			ExpectedEnabled: false,
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(true)
		env.On("RunCommand", "powershell.exe", []string{"-NoProfile", "-NonInteractive", "-Command", spotifySMTCScript}).Return(tc.Output, tc.Error)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		got := s.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, got, tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}
