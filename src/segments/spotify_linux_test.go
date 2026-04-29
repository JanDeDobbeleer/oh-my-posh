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
		Expected        string
		ExpectedEnabled bool
	}{
		{Case: "no data", ExpectedEnabled: false},
		{Case: "error", ExpectedEnabled: false, Status: "Error.ServiceUnknown"},
		{Case: "paused", ExpectedEnabled: true, Expected: "\uf04c Candlemass - Spellbreaker", Status: "paused", Artist: "Candlemass", Track: "Spellbreaker"},
		{Case: "playing", ExpectedEnabled: true, Expected: "\ue602 Candlemass - Spellbreaker", Status: "playing", Artist: "Candlemass", Track: "Spellbreaker"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)

		dbusCMD := "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player"
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:PlaybackStatus | awk -F '\"' '/string/ {print tolower($2)}'").Return(tc.Status)
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'").Return(tc.Artist)
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'").Return(tc.Track)

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
